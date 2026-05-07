package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strings"
	"sync/atomic"
	"time"
)

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const (
	defaultPort         = 1448
	defaultSpeedMPS     = float32(15.0) // metres per second
	tickMS              = 5             // milliseconds per tick
	batteryDrainPerSec  = float32(0.05)
	batteryChargePerSec = float32(0.5)
	batteryLowThreshold = float32(15.0)
	batteryCritical     = float32(5.0)

	jackPosUp   = 26000001
	jackPosDown = 31
)

var jackStepPerTick = (jackPosUp - jackPosDown) * tickMS / 3000

// ---------------------------------------------------------------------------
// MockRobot
// ---------------------------------------------------------------------------

const (
	tickActive = time.Duration(tickMS) * time.Millisecond // 5 ms  — moving / jack animating
	tickIdle   = 200 * time.Millisecond                   // 200 ms — stationary
)

type MockRobot struct {
	ID              int
	state           *RobotState
	actionIDCounter atomic.Int32
	batteryAccum    float32
	stopCh          chan struct{}
}

// newRobot creates and starts the simulation goroutine for a single robot.
// HTTP routing is handled by RobotManager — not started here.
func newRobot(id int, state *RobotState) *MockRobot {
	r := &MockRobot{
		ID:           id,
		state:        state,
		batteryAccum: float32(state.Battery.BatteryPercentage),
		stopCh:       make(chan struct{}),
	}
	r.actionIDCounter.Store(1)
	go r.startSimulationLoop()
	log.Printf("Robot #%d started — pose (%.2f, %.2f) battery %d%%",
		id, state.Pose.X, state.Pose.Y, state.Battery.BatteryPercentage)
	return r
}

func (r *MockRobot) Stop() {
	close(r.stopCh)
}

// isActive returns true when the robot needs high-frequency ticks.
// Called without the state lock — reads are benign race on bool/int fields.
func (r *MockRobot) isActive() bool {
	s := r.state
	s.mu.RLock()
	active := (s.CurrentAction != nil && s.CurrentAction.State.Status == StatusRunning) ||
		s.JackCommand != ""
	s.mu.RUnlock()
	return active
}

// ---------------------------------------------------------------------------
// Simulation loop — adaptive tick rate
// ---------------------------------------------------------------------------

func (r *MockRobot) startSimulationLoop() {
	interval := tickIdle
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.tick()
			// Switch interval based on activity
			desired := tickIdle
			if r.isActive() {
				desired = tickActive
			}
			if desired != interval {
				interval = desired
				ticker.Reset(interval)
			}
		}
	}
}

func (r *MockRobot) tick() {
	defer func() {
		if rec := recover(); rec != nil {
			log.Printf("tick panic: %v", rec)
		}
	}()
	s := r.state
	s.mu.Lock()
	defer s.mu.Unlock()

	r.tickMovement()
	r.tickBattery()
	r.tickJack()
}

func (r *MockRobot) tickMovement() {
	s := r.state
	current := s.CurrentAction
	if current == nil {
		return
	}
	if current.State.Status != StatusRunning {
		return
	}

	// Two-phase docking
	if s.DockingPhase != nil {
		r.tickDockingMovement(current)
		return
	}

	if s.MovementTarget == nil {
		return
	}
	// Respect soft-brake or emergency stop
	if s.SoftBrakeActive || s.Health.HasSystemEmergencyStop {
		return
	}

	tx := s.MovementTarget.X
	ty := s.MovementTarget.Y
	dx := tx - s.Pose.X
	dy := ty - s.Pose.Y
	dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

	step := defaultSpeedMPS * tickMS / 1000.0

	if dist <= step {
		s.Pose.X = tx
		s.Pose.Y = ty
		if s.MovementTarget.YawSet {
			s.Pose.Yaw = s.MovementTarget.Yaw
		}
		s.MovementTarget = nil
		current.State.Status = StatusDone
		current.State.Result = ResultSuccess
		current.Stage = "DONE"

		if s.GoingHome {
			s.GoingHome = false
			s.Battery.DockingStatus = "on_dock"
			s.Battery.IsCharging = true
			s.Battery.IsDCConnected = true
			r.addEvent("charging.docked", fmt.Sprintf("Robot docked and charging at (%.2f, %.2f)", tx, ty))
			log.Printf("MockRobot: action %d – docked and charging at (%.2f, %.2f)", current.ActionID, tx, ty)
		} else {
			r.addEvent("navigation.arrived", fmt.Sprintf("Arrived at target (%.2f, %.2f)", tx, ty))
			log.Printf("MockRobot: action %d done – arrived at (%.2f, %.2f)", current.ActionID, tx, ty)
		}
	} else {
		ratio := step / dist
		s.Pose.X += dx * ratio
		s.Pose.Y += dy * ratio
		s.Pose.Yaw = float32(math.Atan2(float64(dy), float64(dx)))
	}
}

func (r *MockRobot) tickDockingMovement(current *ActionInfo) {
	s := r.state
	phase := s.DockingPhase
	if phase == nil {
		return
	}
	step := defaultSpeedMPS * tickMS / 1000.0

	switch phase.Phase {
	case DockingApproach:
		dx := phase.ApproachX - s.Pose.X
		dy := phase.ApproachY - s.Pose.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		if dist <= step {
			s.Pose.X = phase.ApproachX
			s.Pose.Y = phase.ApproachY

			// transition to docking phase
			phase.Phase = DockingDocking
			current.Stage = "DOCKING_SCAN"

			dockOffset := phase.DockAllowance
			yaw := s.Pose.Yaw

			if phase.BackwardDocking {
				s.MovementTarget = &MovementTarget{
					X:      s.Pose.X - float32(math.Cos(float64(yaw)))*dockOffset,
					Y:      s.Pose.Y - float32(math.Sin(float64(yaw)))*dockOffset,
					Yaw:    yaw,
					YawSet: true,
				}
				log.Printf("MockRobot: action %d – DOCKING phase (BACKWARD) backing %.2fm", current.ActionID, dockOffset)
			} else {
				s.MovementTarget = &MovementTarget{
					X:      s.Pose.X + float32(math.Cos(float64(yaw)))*dockOffset,
					Y:      s.Pose.Y + float32(math.Sin(float64(yaw)))*dockOffset,
					Yaw:    yaw,
					YawSet: true,
				}
				log.Printf("MockRobot: action %d – DOCKING phase (FORWARD) moving %.2fm", current.ActionID, dockOffset)
			}
			r.addEvent("docking.approach_complete", "Arrived at landing point, starting tag alignment")
		} else {
			ratio := step / dist
			s.Pose.X += dx * ratio
			s.Pose.Y += dy * ratio
			s.Pose.Yaw = float32(math.Atan2(float64(dy), float64(dx)))
		}

	case DockingDocking:
		if s.MovementTarget == nil {
			// docking complete
			s.DockingPhase = nil
			current.State.Status = StatusDone
			current.State.Result = ResultSuccess
			current.Stage = "DOCKING_COMPLETE"

			tagTypeName := map[int]string{0: "Visual", 1: "Laser", 2: "Reflector", 3: "Shelf"}[phase.TagType]
			if tagTypeName == "" {
				tagTypeName = "Unknown"
			}
			log.Printf("MockRobot: action %d – MoveToTag COMPLETE, aligned with %s tag (type=%d)",
				current.ActionID, tagTypeName, phase.TagType)
			r.addEvent("docking.complete",
				fmt.Sprintf("Successfully docked with %s tag, tag_ids=%v", tagTypeName, phase.TagIDs))
			return
		}

		tx := s.MovementTarget.X
		ty := s.MovementTarget.Y
		dx := tx - s.Pose.X
		dy := ty - s.Pose.Y
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy)))

		if dist <= step {
			s.Pose.X = tx
			s.Pose.Y = ty
			if s.MovementTarget.YawSet {
				s.Pose.Yaw = s.MovementTarget.Yaw
			}
			s.MovementTarget = nil // triggers completion on next tick
			log.Printf("MockRobot: action %d – tag alignment complete", current.ActionID)
		} else {
			dockSpeed := step * 0.5
			ratio := dockSpeed / dist
			s.Pose.X += dx * ratio
			s.Pose.Y += dy * ratio
		}

	case DockingComplete:
		s.DockingPhase = nil
		s.MovementTarget = nil
	}
}

func (r *MockRobot) tickBattery() {
	s := r.state
	tickSec := float32(tickMS) / 1000.0
	moving := s.CurrentAction != nil && s.CurrentAction.State.Status == StatusRunning

	if s.Battery.DockingStatus == "on_dock" && s.Battery.IsCharging {
		r.batteryAccum += batteryChargePerSec * tickSec
		if r.batteryAccum > 100 {
			r.batteryAccum = 100
		}
	} else if moving {
		r.batteryAccum -= batteryDrainPerSec * tickSec
		if r.batteryAccum < 0 {
			r.batteryAccum = 0
		}

		if r.batteryAccum <= batteryCritical && !s.Health.HasSystemEmergencyStop {
			s.Health.HasSystemEmergencyStop = true
			s.Health.HasError = true
			r.addEvent("system.critical_low_battery", "Battery critically low – emergency stop triggered")
			log.Printf("MockRobot: critical low battery – emergency stop")
		} else if r.batteryAccum <= batteryLowThreshold {
			now := time.Now().UnixMilli()
			if now-s.LastLowBatteryEventMs > 30_000 {
				s.LastLowBatteryEventMs = now
				r.addEvent("system.power_low", fmt.Sprintf("Battery low (%d%%)", int(r.batteryAccum)))
			}
		}
	}
	s.Battery.BatteryPercentage = int(r.batteryAccum)
}

func (r *MockRobot) tickJack() {
	s := r.state
	cmd := s.JackCommand
	if cmd == "" {
		return
	}

	switch cmd {
	case "Up":
		if s.JackActualPos >= jackPosUp {
			s.JackActualPos = jackPosUp
			s.JackStage = 5
			s.JackCommand = ""
			log.Printf("MockRobot: jack fully UP (stage=5)")
			r.addEvent("jack.up", "Jack fully raised")
		} else {
			s.JackActualPos += jackStepPerTick
			if s.JackActualPos > jackPosUp {
				s.JackActualPos = jackPosUp
			}
			s.JackStage = 3
		}
	case "Down":
		if s.JackActualPos <= jackPosDown {
			s.JackActualPos = jackPosDown
			s.JackStage = 2
			s.JackCommand = ""
			log.Printf("MockRobot: jack fully DOWN (stage=2)")
			r.addEvent("jack.down", "Jack fully lowered")
		} else {
			s.JackActualPos -= jackStepPerTick
			if s.JackActualPos < jackPosDown {
				s.JackActualPos = jackPosDown
			}
			s.JackStage = 3
		}
	case "Stop":
		s.JackCommand = ""
		log.Printf("MockRobot: jack STOPPED at pos=%d", s.JackActualPos)
	case "ClearAlarm":
		s.JackAlarm = 0
		s.JackCommand = ""
		log.Printf("MockRobot: jack alarm cleared")
	default:
		s.JackCommand = ""
	}
}

// ---------------------------------------------------------------------------
// addEvent — must be called with state lock held
// ---------------------------------------------------------------------------

func (r *MockRobot) addEvent(typ, message string) {
	ev := RobotEvent{
		Type:      typ,
		Message:   message,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	r.state.Events = append([]RobotEvent{ev}, r.state.Events...)
	if len(r.state.Events) > 50 {
		r.state.Events = r.state.Events[:50]
	}
}

// ---------------------------------------------------------------------------
// Action helpers
// ---------------------------------------------------------------------------

func (r *MockRobot) newAction(name string) *ActionInfo {
	id := int(r.actionIDCounter.Add(1)) - 1
	info := &ActionInfo{
		ActionID:   id,
		ActionName: name,
		Stage:      "GOING_TO_TARGET",
		State: ActionState{
			Status: StatusRunning,
			Result: ResultSuccess,
		},
	}
	r.state.CurrentAction = info
	r.state.ActionHistory[id] = info
	log.Printf("MockRobot: created action id=%d name=%s", id, name)
	return info
}

func (r *MockRobot) idleAction() *ActionInfo {
	return &ActionInfo{
		ActionID:   0,
		ActionName: "idle",
		Stage:      "IDLE",
		State:      ActionState{Status: StatusDone, Result: ResultSuccess},
	}
}

func (r *MockRobot) findAction(id int) *ActionInfo {
	if info, ok := r.state.ActionHistory[id]; ok {
		return info
	}
	return &ActionInfo{
		ActionID:   id,
		ActionName: "unknown",
		State:      ActionState{Status: StatusDone, Result: ResultSuccess},
	}
}

func (r *MockRobot) findPoiByName(name string) *MultiFloorPoi {
	for i, p := range r.state.MultiFloorPois {
		if p.ID == name || p.PoiName == name || p.DisplayName == name {
			return &r.state.MultiFloorPois[i]
		}
	}
	return nil
}

func (r *MockRobot) applyGoHome(info *ActionInfo) {
	s := r.state
	var target *Pose
	if s.HomePose != nil {
		target = s.HomePose
	} else {
		target = s.DockPose
	}
	if target != nil {
		s.Battery.DockingStatus = "not_on_dock"
		s.Battery.IsCharging = false
		s.Battery.IsDCConnected = false
		s.MovementTarget = &MovementTarget{X: target.X, Y: target.Y, Yaw: target.Yaw, YawSet: true}
		s.GoingHome = true
		s.DeliveryStage = "GOING_HOME"
		log.Printf("MockRobot: go_home – heading to (%.2f, %.2f)", target.X, target.Y)
	} else {
		info.State.Status = StatusDone
		info.State.Result = ResultSuccess
		s.Battery.DockingStatus = "on_dock"
		s.Battery.IsCharging = true
		s.Battery.IsDCConnected = true
	}
}

// ---------------------------------------------------------------------------
// createActionFromBody — parses action JSON and starts simulation
// ---------------------------------------------------------------------------

func (r *MockRobot) createActionFromBody(body []byte) *ActionInfo {
	actionName := r.resolveActionName(body)
	info := r.newAction(actionName)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(body, &raw); err != nil {
		return info
	}

	// Determine options object — either {"options": {...}} or flat
	options := raw
	if optRaw, ok := raw["options"]; ok {
		var opts map[string]json.RawMessage
		if err := json.Unmarshal(optRaw, &opts); err == nil {
			options = opts
		}
	}

	isSchedulableMoveToTag := strings.Contains(actionName, "SchedulableMoveToTag")
	isMoveToTag := strings.Contains(actionName, "MoveToTag")

	// ---- MoveToAction / MultiFloorMoveAction ----
	if !isSchedulableMoveToTag && (strings.Contains(actionName, "move_to") ||
		strings.Contains(actionName, "MoveToAction") ||
		strings.Contains(actionName, "MultiFloorMove") ||
		strings.Contains(actionName, "multi_floor_move") ||
		hasKey(raw, "target")) {

		targetRaw, ok := options["target"]
		if ok {
			var target map[string]json.RawMessage
			if err := json.Unmarshal(targetRaw, &target); err == nil {
				tx := r.state.Pose.X
				ty := r.state.Pose.Y
				tyaw := r.state.Pose.Yaw
				hasYaw := false

				if v, ok := target["x"]; ok {
					json.Unmarshal(v, &tx)
				}
				if v, ok := target["y"]; ok {
					json.Unmarshal(v, &ty)
				}
				if v, ok := target["yaw"]; ok {
					json.Unmarshal(v, &tyaw)
					hasYaw = true
				}

				if poiRaw, ok := target["poi_name"]; ok {
					var poiName string
					json.Unmarshal(poiRaw, &poiName)
					poi := r.findPoiByName(poiName)
					if poi != nil {
						tx = poi.Pose.X
						ty = poi.Pose.Y
						tyaw = poi.Pose.Yaw
						hasYaw = true
						log.Printf("MockRobot: navigate to POI '%s' at (%.2f, %.2f)", poiName, tx, ty)
					} else {
						log.Printf("MockRobot: POI '%s' not found – action will fail", poiName)
						info.State.Status = StatusDone
						info.State.Result = ResultFailed
						info.State.Reason = "POI not found: " + poiName
						return info
					}
				}

				r.state.MovementTarget = &MovementTarget{X: tx, Y: ty, Yaw: tyaw, YawSet: hasYaw}
				r.state.GoingHome = false
				r.state.DeliveryStage = "GOING_TO_TASK_POINT"
				log.Printf("MockRobot: action %d – move_to (%.2f, %.2f)", info.ActionID, tx, ty)
			}
		}
	}

	// ---- GoHomeAction ----
	if strings.Contains(actionName, "go_home") ||
		strings.Contains(actionName, "GoHome") ||
		strings.Contains(actionName, "BackHome") {
		r.applyGoHome(info)
	}

	// ---- RotateAction ----
	if strings.Contains(actionName, "rotate") || strings.Contains(actionName, "Rotate") {
		var angle float32 = 3.14
		if v, ok := options["angle"]; ok {
			json.Unmarshal(v, &angle)
		}
		r.state.Pose.Yaw = float32(math.Mod(float64(r.state.Pose.Yaw+angle), 2*math.Pi))
		info.State.Status = StatusDone
		info.State.Result = ResultSuccess
		r.state.MovementTarget = nil
		log.Printf("MockRobot: action %d – rotate by %.2f rad → new yaw=%.2f", info.ActionID, angle, r.state.Pose.Yaw)
	}

	// ---- EnterElevatorAction ----
	if strings.Contains(actionName, "enter_elevator") || strings.Contains(actionName, "EnterElevator") {
		go func(actionID int) {
			time.Sleep(1500 * time.Millisecond)
			r.state.mu.Lock()
			defer r.state.mu.Unlock()
			if r.state.CurrentAction != nil && r.state.CurrentAction.ActionID == actionID {
				r.state.CurrentAction.State.Status = StatusDone
				r.state.CurrentAction.State.Result = ResultSuccess
				r.addEvent("elevator.entered", "Robot entered elevator")
			}
		}(info.ActionID)
	}

	// ---- LeaveElevatorAction ----
	if strings.Contains(actionName, "leave_elevator") || strings.Contains(actionName, "LeaveElevator") {
		go func(actionID int) {
			time.Sleep(1500 * time.Millisecond)
			r.state.mu.Lock()
			defer r.state.mu.Unlock()
			if r.state.CurrentAction != nil && r.state.CurrentAction.ActionID == actionID {
				r.state.CurrentAction.State.Status = StatusDone
				r.state.CurrentAction.State.Result = ResultSuccess
				r.addEvent("elevator.left", "Robot left elevator")
			}
		}(info.ActionID)
	}

	// ---- SchedulableMoveToTagAction / MoveToTagAction ----
	if isSchedulableMoveToTag || isMoveToTag {
		approachX := r.state.Pose.X
		approachY := r.state.Pose.Y
		approachYaw := r.state.Pose.Yaw

		if targetRaw, ok := options["target"]; ok {
			var target map[string]json.RawMessage
			if err := json.Unmarshal(targetRaw, &target); err == nil {
				if v, ok := target["x"]; ok {
					json.Unmarshal(v, &approachX)
				}
				if v, ok := target["y"]; ok {
					json.Unmarshal(v, &approachY)
				}
				if v, ok := target["yaw"]; ok {
					json.Unmarshal(v, &approachYaw)
				}
			}
		}

		backwardDocking := false
		dockAllowance := float32(0.275)
		tagType := 3
		reflectTagNum := 2
		var tagIDs []int

		if tagOptsRaw, ok := options["move_to_tag_options"]; ok {
			var tagOpts map[string]json.RawMessage
			if err := json.Unmarshal(tagOptsRaw, &tagOpts); err == nil {
				if v, ok := tagOpts["backward_docking"]; ok {
					json.Unmarshal(v, &backwardDocking)
				}
				if v, ok := tagOpts["dock_allowance"]; ok {
					json.Unmarshal(v, &dockAllowance)
				}
				if v, ok := tagOpts["tag_type"]; ok {
					json.Unmarshal(v, &tagType)
				}
				if v, ok := tagOpts["reflect_tag_num"]; ok {
					json.Unmarshal(v, &reflectTagNum)
				}
				if v, ok := tagOpts["tag_ids"]; ok {
					json.Unmarshal(v, &tagIDs)
				}
			}
		}

		r.state.DockingPhase = &DockingPhaseState{
			Phase:          DockingApproach,
			ApproachX:      approachX,
			ApproachY:      approachY,
			TagX:           approachX,
			TagY:           approachY,
			TagYaw:         approachYaw,
			BackwardDocking: backwardDocking,
			DockAllowance:  dockAllowance,
			TagIDs:         tagIDs,
			TagType:        tagType,
			ReflectTagNum:  reflectTagNum,
		}
		info.Stage = "DOCKING_APPROACH"
		r.state.MovementTarget = &MovementTarget{X: approachX, Y: approachY, Yaw: approachYaw, YawSet: true}
		log.Printf("MockRobot: action %d – MoveToTag APPROACH → (%.2f, %.2f) yaw=%.2f backward=%v",
			info.ActionID, approachX, approachY, approachYaw, backwardDocking)
	}

	return info
}

func (r *MockRobot) resolveActionName(body []byte) string {
	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err == nil {
		if v, ok := m["action_name"]; ok {
			var name string
			if json.Unmarshal(v, &name) == nil && name != "" {
				return name
			}
		}
		if _, ok := m["target"]; ok {
			return "slamtec.agent.actions.MultiFloorMoveAction"
		}
	}
	bs := string(body)
	switch {
	case strings.Contains(bs, "SchedulableMoveToTagAction"):
		return "slamtec.agent.actions.SchedulableMoveToTagAction"
	case strings.Contains(bs, "MoveToTagAction"):
		return "slamtec.agent.actions.MoveToTagAction"
	case strings.Contains(bs, "MoveToAction"):
		return "slamtec.agent.actions.MoveToAction"
	case strings.Contains(bs, "MultiFloorMoveAction"):
		return "slamtec.agent.actions.MultiFloorMoveAction"
	case strings.Contains(bs, "EnterElevatorAction"):
		return "slamtec.agent.actions.EnterElevatorAction"
	case strings.Contains(bs, "LeaveElevatorAction"):
		return "slamtec.agent.actions.LeaveElevatorAction"
	case strings.Contains(bs, "RotateAction"):
		return "slamtec.agent.actions.RotateAction"
	case strings.Contains(bs, "GoHomeAction"), strings.Contains(bs, "BackHomeAction"):
		return "slamtec.agent.actions.GoHomeAction"
	case strings.Contains(bs, "move_to"):
		return "slamtec.agent.actions.MoveToAction"
	case strings.Contains(bs, "go_home"):
		return "slamtec.agent.actions.GoHomeAction"
	case strings.Contains(bs, "rotate"):
		return "slamtec.agent.actions.RotateAction"
	default:
		return "slamtec.agent.actions.UnknownAction"
	}
}

func hasKey(m map[string]json.RawMessage, key string) bool {
	_, ok := m[key]
	return ok
}

// ---------------------------------------------------------------------------
// Map header builder
// ---------------------------------------------------------------------------

func buildMapHeader(gridX, gridY int, originX, originY, resolution float32) []byte {
	buf := make([]byte, 32)
	binary.LittleEndian.PutUint32(buf[0:], uint32(gridX))
	binary.LittleEndian.PutUint32(buf[4:], uint32(gridY))
	putFloat32LE(buf[8:], originX)
	putFloat32LE(buf[12:], originY)
	putFloat32LE(buf[16:], resolution)
	return buf
}

func putFloat32LE(b []byte, f float32) {
	bits := math.Float32bits(f)
	binary.LittleEndian.PutUint32(b, bits)
}


