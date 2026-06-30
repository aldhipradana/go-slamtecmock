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
	defaultPort          = 1448
	defaultSpeedMPS      = float32(15.0) // metres per second
	tickMS               = 5             // milliseconds per tick
	batteryDrainPerSec   = float32(0.05)
	batteryChargePerSec  = float32(0.5)
	batteryLowThreshold  = float32(15.0)
	batteryCritical      = float32(5.0)
	cameraQrDetectRadius = float32(45.0)

	jackPosUp   = 26000001
	jackPosDown = 31
)

var jackStepPerTick = (jackPosUp - jackPosDown) * tickMS / 3000

// ---------------------------------------------------------------------------
// MockRobot
// ---------------------------------------------------------------------------

type MockRobot struct {
	state           *RobotState
	actionIDCounter atomic.Int32
	batteryAccum    float32
	stopCh          chan struct{}
}

func NewMockRobot(port int, state *RobotState) *MockRobot {
	r := &MockRobot{
		state:        state,
		batteryAccum: float32(state.Battery.BatteryPercentage),
		stopCh:       make(chan struct{}),
	}
	r.actionIDCounter.Store(1)

	router := r.buildRouter()

	log.Printf("MockRobot starting on port %d — initial pose (%.2f, %.2f), battery %d%%",
		port, state.Pose.X, state.Pose.Y, state.Battery.BatteryPercentage)

	go r.startSimulationLoop()
	go func() {
		addr := fmt.Sprintf(":%d", port)
		if err := serveHTTP(addr, router); err != nil {
			log.Fatalf("HTTP server error: %v", err)
		}
	}()
	return r
}

func (r *MockRobot) Stop() {
	close(r.stopCh)
}

// ---------------------------------------------------------------------------
// Simulation loop
// ---------------------------------------------------------------------------

func (r *MockRobot) startSimulationLoop() {
	ticker := time.NewTicker(tickMS * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			r.tick()
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
		arrivedTarget := s.MovementTarget
		s.Pose.X = tx
		s.Pose.Y = ty
		if arrivedTarget.YawSet {
			s.Pose.Yaw = arrivedTarget.Yaw
		}
		if arrivedTarget != nil {
			r.syncCurrentFloorLocked(arrivedTarget.Building, arrivedTarget.FloorID)
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
	wasBatteryCritical := r.batteryAccum <= batteryCritical

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

		if r.batteryAccum <= batteryCritical && !wasBatteryCritical {
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
	r.recomputeEmergencyStopLocked()
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

func (r *MockRobot) abortCurrentActionLocked(eventType, message string) {
	if r.state.CurrentAction != nil && r.state.CurrentAction.State.Status == StatusRunning {
		r.state.CurrentAction.State.Status = StatusDone
		r.state.CurrentAction.State.Result = ResultAborted
		r.state.CurrentAction.Stage = "ABORTED"
		r.state.MovementTarget = nil
		r.state.DockingPhase = nil
		r.state.GoingHome = false
		if eventType != "" && message != "" {
			r.addEvent(eventType, message)
		}
	}
}

func (r *MockRobot) recomputeEmergencyStopLocked() {
	batteryCriticalActive := r.batteryAccum <= batteryCritical || float32(r.state.Battery.BatteryPercentage) <= batteryCritical
	hasEmergencyStop := r.state.SoftBrakeActive || r.state.PhysicalEStopActive || batteryCriticalActive
	r.state.Health.HasSystemEmergencyStop = hasEmergencyStop
	r.state.Health.HasError = hasEmergencyStop
	if !r.state.CliffSafe {
		r.state.Health.HasWarning = true
	} else if !hasEmergencyStop {
		r.state.Health.HasWarning = false
	}
}

func (r *MockRobot) setSoftBrakeLocked(active bool, source string) {
	if r.state.SoftBrakeActive == active {
		return
	}

	r.state.SoftBrakeActive = active
	if active {
		r.abortCurrentActionLocked("system.emergency_stop", "Current action aborted due to soft-brake activation")
		r.addEvent("system.emergency_stop", "Soft-brake activated via "+source)
		log.Printf("MockRobot: soft-brake ACTIVATED via %s", source)
	} else {
		r.addEvent("system.emergency_stop_released", "Soft-brake released via "+source)
		log.Printf("MockRobot: soft-brake RELEASED via %s", source)
	}
	r.recomputeEmergencyStopLocked()
}

func (r *MockRobot) setPhysicalEStopLocked(active bool, source string) {
	if r.state.PhysicalEStopActive == active {
		return
	}

	r.state.PhysicalEStopActive = active
	if active {
		r.abortCurrentActionLocked("system.physical_estop_action_aborted", "Current action aborted due to physical e-stop")
		r.addEvent("system.physical_estop", "Physical e-stop pressed via "+source)
		log.Printf("MockRobot: physical e-stop ACTIVATED via %s", source)
	} else {
		r.addEvent("system.physical_estop_released", "Physical e-stop released via "+source)
		log.Printf("MockRobot: physical e-stop RELEASED via %s", source)
	}
	r.recomputeEmergencyStopLocked()
}

func (r *MockRobot) setCliffSafeLocked(safe bool, source string) {
	if r.state.CliffSafe == safe {
		return
	}

	r.state.CliffSafe = safe
	r.state.Health.HasWarning = !safe || r.state.Health.HasWarning
	if safe {
		r.addEvent("sensor.cliff_safe", "Cliff sensor returned SAFE via "+source)
		log.Printf("MockRobot: cliff SAFE via %s", source)
	} else {
		r.addEvent("sensor.cliff_unsafe", "Cliff sensor reported UNSAFE via "+source)
		log.Printf("MockRobot: cliff UNSAFE via %s", source)
	}
	r.recomputeEmergencyStopLocked()
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

func (r *MockRobot) findPoiByPose(x, y float32) *MultiFloorPoi {
	const tolerance = 0.001
	for i, p := range r.state.MultiFloorPois {
		if math.Abs(float64(p.Pose.X-x)) <= tolerance && math.Abs(float64(p.Pose.Y-y)) <= tolerance {
			return &r.state.MultiFloorPois[i]
		}
	}
	return nil
}

func (r *MockRobot) syncCurrentFloorLocked(building, floorID string) {
	if building != "" {
		r.state.CurrentFloor.Building = building
	}
	if floorID != "" {
		r.state.CurrentFloor.FloorID = floorID
	}
}

func (r *MockRobot) currentVisibleFrontQrLocked() string {
	if !r.state.FrontCamOn {
		return ""
	}
	if r.state.FrontCamQrID != "" {
		return r.state.FrontCamQrID
	}
	return r.currentVisibleQrByPoseLocked()
}

func (r *MockRobot) currentVisibleBackQrLocked() string {
	if !r.state.BackCamOn {
		return ""
	}
	if r.state.BackCamQrID != "" {
		return r.state.BackCamQrID
	}
	return ""
}

func (r *MockRobot) currentVisibleQrByPoseLocked() string {
	currentBuilding := r.state.CurrentFloor.Building
	currentFloor := r.state.CurrentFloor.FloorID
	bestDistance := float32(math.MaxFloat32)
	bestQr := ""

	for _, poi := range r.state.MultiFloorPois {
		if !isCameraVisiblePoiType(poi.Type) {
			continue
		}
		if currentBuilding != "" && poi.Building != currentBuilding {
			continue
		}
		if currentFloor != "" && poi.Floor != currentFloor {
			continue
		}

		dx := poi.Pose.X - r.state.Pose.X
		dy := poi.Pose.Y - r.state.Pose.Y
		distance := float32(math.Sqrt(float64(dx*dx + dy*dy)))
		if distance <= cameraQrDetectRadius && distance < bestDistance {
			bestDistance = distance
			bestQr = poi.ID
		}
	}

	return bestQr
}

func isCameraVisiblePoiType(poiType string) bool {
	switch poiType {
	case "TROLLEY_POINT", "AV_LOADING_POINT", "LIFT_TROLLEY_POINT", "QR_POINT":
		return true
	default:
		return false
	}
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
				targetBuilding := ""
				targetFloor := ""

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
				if v, ok := target["building"]; ok {
					json.Unmarshal(v, &targetBuilding)
				}
				if v, ok := target["floor"]; ok {
					json.Unmarshal(v, &targetFloor)
				}
				if poseRaw, ok := target["pose"]; ok {
					var pose map[string]json.RawMessage
					if err := json.Unmarshal(poseRaw, &pose); err == nil {
						if v, ok := pose["x"]; ok {
							json.Unmarshal(v, &tx)
						}
						if v, ok := pose["y"]; ok {
							json.Unmarshal(v, &ty)
						}
						if v, ok := pose["yaw"]; ok {
							json.Unmarshal(v, &tyaw)
							hasYaw = true
						}
					}
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

				movementTarget := &MovementTarget{X: tx, Y: ty, Yaw: tyaw, YawSet: hasYaw}
				if targetBuilding != "" || targetFloor != "" {
					movementTarget.Building = targetBuilding
					movementTarget.FloorID = targetFloor
				}
				if poiRaw, ok := target["poi_name"]; ok {
					var poiName string
					json.Unmarshal(poiRaw, &poiName)
					if poi := r.findPoiByName(poiName); poi != nil {
						movementTarget.Building = poi.Building
						movementTarget.FloorID = poi.Floor
					}
				} else if poi := r.findPoiByPose(tx, ty); poi != nil {
					movementTarget.Building = poi.Building
					movementTarget.FloorID = poi.Floor
				}

				r.state.MovementTarget = movementTarget
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
			Phase:           DockingApproach,
			ApproachX:       approachX,
			ApproachY:       approachY,
			TagX:            approachX,
			TagY:            approachY,
			TagYaw:          approachYaw,
			BackwardDocking: backwardDocking,
			DockAllowance:   dockAllowance,
			TagIDs:          tagIDs,
			TagType:         tagType,
			ReflectTagNum:   reflectTagNum,
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
