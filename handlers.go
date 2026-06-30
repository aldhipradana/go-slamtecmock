package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ---------------------------------------------------------------------------
// HTTP helpers
// ---------------------------------------------------------------------------

func jsonResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("jsonResponse encode error: %v", err)
	}
}

func readBody(r *http.Request) []byte {
	b, _ := io.ReadAll(r.Body)
	return b
}

func serveHTTP(addr string, handler http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: handler}
	return srv.ListenAndServe()
}

// ---------------------------------------------------------------------------
// Router
// ---------------------------------------------------------------------------

func (r *MockRobot) buildRouter() http.Handler {
	mux := chi.NewRouter()
	mux.Use(middleware.Recoverer)
	mux.Use(corsMiddleware)

	mux.Get("/", r.handleUIRedirect)
	mux.Get("/ui", r.handleUIDashboard)
	mux.Get("/ui/partials/summary", r.handleUISummary)
	mux.Get("/ui/partials/events", r.handleUIEvents)
	mux.Get("/ui/partials/pois", r.handleUIPois)
	mux.Post("/ui/actions/cliff/{mode}", r.handleUISetCliff)
	mux.Post("/ui/actions/physical-estop/{mode}", r.handleUISetPhysicalEStop)

	// 3.1 System Resources
	mux.Get("/api/core/system/v1/power/status", r.handleGetPowerStatus)
	mux.Get("/api/core/system/v1/network/status", r.handleGetNetworkStatus)
	mux.Get("/api/core/system/v1/robot/info", r.handleGetRobotInfo)
	mux.Get("/api/core/system/v1/robot/health", r.handleGetRobotHealth)
	mux.Get("/api/core/system/v1/capabilities", r.handleGetCapabilities)
	mux.Get("/api/core/system/v1/parameter", r.handleGetParameter)
	mux.Put("/api/core/system/v1/parameter", r.handlePutParameter)
	mux.Post("/api/core/system/v1/power/{cmd}", r.handlePowerCmd)
	mux.Get("/api/core/system/v1/jack/status", r.handleGetJackStatus)
	mux.Post("/api/core/system/v1/jack/status", r.handlePostJackStatus)

	// 3.2 SLAM
	mux.Get("/api/core/slam/v1/localization/pose", r.handleGetPose)
	mux.Get("/api/core/slam/v1/homepose", r.handleGetHomePose)
	mux.Put("/api/core/slam/v1/homepose", r.handlePutHomePose)
	mux.Put("/api/multi-floor/localization/v1/pose", r.handlePutPose)
	mux.Get("/api/core/slam/v1/maps/explore", r.handleGetMapExplore)
	mux.Get("/api/core/slam/v1/maps/stcm", r.handleGetMapStcm)
	mux.Put("/api/core/slam/v1/mapping/{enable}", r.handlePutMapping)

	// 3.3 Artifacts – POIs
	mux.Get("/api/core/artifact/v1/pois", r.handleGetArtifactPois)
	mux.Post("/api/core/artifact/v1/pois", r.handlePostArtifactPoi)
	mux.Delete("/api/core/artifact/v1/pois/{id}", r.handleDeleteArtifactPoi)

	// 3.4 Motion – actions
	mux.Get("/api/core/motion/v1/action-factories", r.handleGetActionFactories)
	mux.Get("/api/core/motion/v1/strategies", r.handleGetStrategies)
	mux.Get("/api/core/motion/v1/strategies/{id}", r.handleGetCurrentStrategy)
	mux.Put("/api/core/motion/v1/strategies/{id}", r.handlePutStrategy)
	mux.Post("/api/core/slam/v1/action/create", r.handleCreateAction)
	mux.Post("/api/core/motion/v1/actions", r.handleCreateAction)
	mux.Get("/api/core/motion/v1/actions/{actionId}", r.handleGetAction)
	mux.Delete("/api/core/motion/v1/actions/{actionId}", r.handleDeleteAction)

	// 3.9 Multi-floor
	mux.Post("/api/multi-floor/motion/v1/movetoaction", r.handleCreateAction)
	mux.Get("/api/multi-floor/map/v1/floors", r.handleGetFloors)
	mux.Get("/api/multi-floor/map/v1/floors/{floorId}", r.handleGetCurrentFloor)
	mux.Put("/api/multi-floor/map/v1/floors/{floorId}", r.handlePutFloor)
	mux.Get("/api/multi-floor/map/v1/pois", r.handleGetMultiFloorPois)
	mux.Get("/api/multi-floor/map/v1/elevators/{elevatorId}", r.handleGetElevator)

	// 3.8 Platform – events
	mux.Get("/api/platform/v1/events", r.handleGetEvents)

	// 3.10 Delivery
	mux.Get("/api/delivery/v1/stage", r.handleGetDeliveryStage)
	mux.Get("/api/delivery/v1/settings", r.handleGetDeliverySettings)
	mux.Put("/api/delivery/v1/settings/timeout", r.handlePutDeliveryTimeout)
	mux.Put("/api/delivery/v1/tasks/{cmd}", r.handlePutDeliveryTask)
	mux.Get("/api/delivery/v1/cargos", r.handleGetCargos)
	mux.Put("/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}", r.handlePutCargoBoxOp)
	mux.Get("/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/operation_result", r.handleGetCargoOpResult)

	// External sensors
	mux.Get("/front_cam", r.handleFrontCam)
	mux.Get("/back_cam", r.handleBackCam)
	mux.Get("/cliff_safe", r.handleCliffSafe)
	mux.Get("/sensor/status", r.handleSensorStatus)
	mux.Post("/sensor/lidar/on", r.handleLidarOn)
	mux.Post("/sensor/lidar/off", r.handleLidarOff)
	mux.Post("/sensor/front_cam/on", r.handleFrontCamOn)
	mux.Post("/sensor/front_cam/off", r.handleFrontCamOff)
	mux.Post("/sensor/back_cam/on", r.handleBackCamOn)
	mux.Post("/sensor/back_cam/off", r.handleBackCamOff)

	return mux
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if req.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, req)
	})
}

// ---------------------------------------------------------------------------
// 3.1 System Resources handlers
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetPowerStatus(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Battery
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetNetworkStatus(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Network
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetRobotInfo(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.RobotInfo
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetRobotHealth(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Health
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetCapabilities(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Capabilities
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetParameter(w http.ResponseWriter, req *http.Request) {
	param := req.URL.Query().Get("param")
	if param == "" {
		http.Error(w, "Parameter is required", http.StatusBadRequest)
		return
	}
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()
	if param == "base.emergency_stop" {
		val := "off"
		if r.state.SoftBrakeActive {
			val = "on"
		}
		jsonResponse(w, val)
		return
	}
	val, ok := r.state.SystemParameters[param]
	if !ok {
		http.Error(w, "Unknown parameter: "+param, http.StatusBadRequest)
		return
	}
	jsonResponse(w, val)
}

func (r *MockRobot) handlePutParameter(w http.ResponseWriter, req *http.Request) {
	body := readBody(req)
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	var m map[string]json.RawMessage
	if err := json.Unmarshal(body, &m); err != nil {
		jsonResponse(w, true)
		return
	}

	paramRaw, hasParam := m["param"]
	valueRaw, hasValue := m["value"]

	if hasParam && hasValue {
		var param, value string
		json.Unmarshal(paramRaw, &param)
		json.Unmarshal(valueRaw, &value)
		if param != "" {
			r.state.SystemParameters[param] = value

			if param == "base.emergency_stop" {
				r.setSoftBrakeLocked(strings.EqualFold(value, "on"), "parameter API")
			}

			if param == "base.brake_release" {
				released := strings.EqualFold(value, "on")
				msg := "restored"
				if released {
					msg = "released"
				}
				r.addEvent("system.brake_release", "Brake "+msg+" via parameter API")
				log.Printf("MockRobot: brake %s", strings.ToUpper(msg))
			}
		}
	} else {
		// legacy map-style body
		for k, v := range m {
			var sv string
			if json.Unmarshal(v, &sv) == nil {
				r.state.SystemParameters[k] = sv
			}
		}
	}
	jsonResponse(w, true)
}

func (r *MockRobot) handlePowerCmd(w http.ResponseWriter, req *http.Request) {
	cmd := chi.URLParam(req, "cmd")
	log.Printf("MockRobot: system command '%s'", cmd)
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleGetJackStatus(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	s := JackStatus{
		ActualPos: r.state.JackActualPos,
		Alarm:     r.state.JackAlarm,
		DrvStatus: r.state.JackDrvStatus,
		Stage:     r.state.JackStage,
	}
	r.state.mu.RUnlock()
	jsonResponse(w, s)
}

func (r *MockRobot) handlePostJackStatus(w http.ResponseWriter, req *http.Request) {
	raw := strings.TrimSpace(string(readBody(req)))
	// strip surrounding quotes if present
	if len(raw) >= 2 && raw[0] == '"' && raw[len(raw)-1] == '"' {
		raw = raw[1 : len(raw)-1]
	}
	log.Printf("MockRobot: jack command received: '%s'", raw)
	switch raw {
	case "Up", "Down", "Stop", "ClearAlarm":
		r.state.mu.Lock()
		r.state.JackCommand = raw
		r.state.mu.Unlock()
		w.WriteHeader(http.StatusOK)
	default:
		log.Printf("MockRobot: unknown jack command '%s'", raw)
		http.Error(w, "Unknown jack command: "+raw, http.StatusBadRequest)
	}
}

// ---------------------------------------------------------------------------
// 3.2 SLAM handlers
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetPose(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Pose
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetHomePose(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	hp := r.state.HomePose
	r.state.mu.RUnlock()
	if hp == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, "{}")
		return
	}
	jsonResponse(w, hp)
}

func (r *MockRobot) handlePutHomePose(w http.ResponseWriter, req *http.Request) {
	var p Pose
	if err := json.NewDecoder(req.Body).Decode(&p); err == nil {
		r.state.mu.Lock()
		r.state.HomePose = &p
		dock := newPose(p.X, p.Y, p.Yaw)
		r.state.DockPose = &dock
		r.state.mu.Unlock()
		log.Printf("MockRobot: home pose registered at (%.2f, %.2f)", p.X, p.Y)
	}
	jsonResponse(w, true)
}

func (r *MockRobot) handlePutPose(w http.ResponseWriter, req *http.Request) {
	var p Pose
	if err := json.NewDecoder(req.Body).Decode(&p); err == nil {
		r.state.mu.Lock()
		r.state.Pose.X = p.X
		r.state.Pose.Y = p.Y
		r.state.Pose.Yaw = p.Yaw
		r.addEvent("localization.updated", fmt.Sprintf("Pose manually set to (%.2f, %.2f)", p.X, p.Y))
		r.state.mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleGetMapExplore(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	m := r.state.MapMeta
	r.state.mu.RUnlock()
	header := buildMapHeader(m.GridX, m.GridY, m.OriginX, m.OriginY, m.Resolution)
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(header)
}

func (r *MockRobot) handleGetMapStcm(w http.ResponseWriter, req *http.Request) {
	r.handleGetMapExplore(w, req)
}

func (r *MockRobot) handlePutMapping(w http.ResponseWriter, req *http.Request) {
	enable := chi.URLParam(req, "enable")
	r.state.mu.Lock()
	r.state.MappingEnabled = strings.EqualFold(enable, "true")
	r.state.mu.Unlock()
	log.Printf("MockRobot: mapping %v", r.state.MappingEnabled)
	fmt.Fprint(w, "true")
}

// ---------------------------------------------------------------------------
// 3.3 Artifacts – POIs
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetArtifactPois(w http.ResponseWriter, req *http.Request) {
	typ := req.URL.Query().Get("type")
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()
	if typ != "" {
		var filtered []ArtifactPoi
		for _, p := range r.state.ArtifactPois {
			if p.Type == typ {
				filtered = append(filtered, p)
			}
		}
		if filtered == nil {
			filtered = []ArtifactPoi{}
		}
		jsonResponse(w, filtered)
		return
	}
	jsonResponse(w, r.state.ArtifactPois)
}

func (r *MockRobot) handlePostArtifactPoi(w http.ResponseWriter, req *http.Request) {
	var poi ArtifactPoi
	if err := json.NewDecoder(req.Body).Decode(&poi); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	r.state.mu.Lock()
	if poi.ID == "" {
		poi.ID = fmt.Sprintf("poi-%d", time.Now().UnixNano())
	}
	// set pose to current if not provided
	if poi.Pose.X == 0 && poi.Pose.Y == 0 {
		poi.Pose = r.state.Pose
	}
	r.state.ArtifactPois = append(r.state.ArtifactPois, poi)
	r.state.mu.Unlock()
	jsonResponse(w, poi)
}

func (r *MockRobot) handleDeleteArtifactPoi(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "id")
	r.state.mu.Lock()
	pois := r.state.ArtifactPois[:0]
	for _, p := range r.state.ArtifactPois {
		if p.ID != id {
			pois = append(pois, p)
		}
	}
	r.state.ArtifactPois = pois
	r.state.mu.Unlock()
	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// 3.4 Motion handlers
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetActionFactories(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.SupportedActions
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetStrategies(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.MotionStrategies
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetCurrentStrategy(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.CurrentMotionStrategy
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handlePutStrategy(w http.ResponseWriter, req *http.Request) {
	var s MotionStrategy
	if err := json.NewDecoder(req.Body).Decode(&s); err == nil {
		r.state.mu.Lock()
		r.state.CurrentMotionStrategy = s
		r.state.mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleCreateAction(w http.ResponseWriter, req *http.Request) {
	body := readBody(req)
	r.state.mu.Lock()
	info := r.createActionFromBody(body)
	r.state.mu.Unlock()
	jsonResponse(w, info)
}

func (r *MockRobot) handleGetAction(w http.ResponseWriter, req *http.Request) {
	seg := chi.URLParam(req, "actionId")
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()
	if seg == ":current" || seg == "current" {
		if r.state.CurrentAction != nil {
			jsonResponse(w, r.state.CurrentAction)
		} else {
			jsonResponse(w, r.idleAction())
		}
		return
	}
	id, err := strconv.Atoi(seg)
	if err != nil {
		jsonResponse(w, r.idleAction())
		return
	}
	jsonResponse(w, r.findAction(id))
}

func (r *MockRobot) handleDeleteAction(w http.ResponseWriter, req *http.Request) {
	log.Printf("MockRobot: abort action requested")
	r.state.mu.Lock()
	r.abortCurrentActionLocked("navigation.aborted", "Current action aborted by caller")
	r.state.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

// ---------------------------------------------------------------------------
// 3.9 Multi-floor handlers
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetFloors(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Floors
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handleGetCurrentFloor(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.CurrentFloor
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handlePutFloor(w http.ResponseWriter, req *http.Request) {
	var f Floor
	body := readBody(req)
	log.Printf("MockRobot: payload for floor update: %s", body)
	if err := json.Unmarshal(body, &f); err == nil {
		r.state.mu.Lock()
		if f.Building != "" {
			r.state.CurrentFloor.Building = f.Building
		}
		if f.FloorID != "" {
			r.state.CurrentFloor.FloorID = f.FloorID
		}
		r.addEvent("floor.updated", "Floor updated to "+r.state.CurrentFloor.Building+"/"+r.state.CurrentFloor.FloorID)
		r.state.mu.Unlock()
	}
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleGetMultiFloorPois(w http.ResponseWriter, req *http.Request) {
	building := req.URL.Query().Get("building")
	floor := req.URL.Query().Get("floor")
	typ := req.URL.Query().Get("type")
	group := req.URL.Query().Get("group")

	r.state.mu.RLock()
	defer r.state.mu.RUnlock()
	var result []MultiFloorPoi
	for _, p := range r.state.MultiFloorPois {
		if building != "" && p.Building != building {
			continue
		}
		if floor != "" && p.Floor != floor {
			continue
		}
		if typ != "" && p.Type != typ {
			continue
		}
		if group != "" && p.Group != group {
			continue
		}
		result = append(result, p)
	}
	if result == nil {
		result = []MultiFloorPoi{}
	}
	jsonResponse(w, result)
}

func (r *MockRobot) handleGetElevator(w http.ResponseWriter, req *http.Request) {
	id := chi.URLParam(req, "elevatorId")
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()
	for _, el := range r.state.Elevators {
		if el.ElevatorID == id {
			jsonResponse(w, el)
			return
		}
	}
	// default
	el := Elevator{
		ElevatorID:           id,
		DoorType:             "single",
		Optional:             false,
		FrontSchedulingPoses: []Pose{newPose(r.state.Pose.X+1.0, r.state.Pose.Y, 0)},
	}
	jsonResponse(w, el)
}

// ---------------------------------------------------------------------------
// 3.8 Events
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetEvents(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Events
	r.state.mu.RUnlock()
	if v == nil {
		v = []RobotEvent{}
	}
	jsonResponse(w, v)
}

// ---------------------------------------------------------------------------
// 3.10 Delivery
// ---------------------------------------------------------------------------

func (r *MockRobot) handleGetDeliveryStage(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	stage := r.state.DeliveryStage
	r.state.mu.RUnlock()
	jsonResponse(w, map[string]string{"stage": stage})
}

func (r *MockRobot) handleGetDeliverySettings(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.DeliverySettings
	r.state.mu.RUnlock()
	jsonResponse(w, v)
}

func (r *MockRobot) handlePutDeliveryTimeout(w http.ResponseWriter, req *http.Request) {
	var body map[string]json.RawMessage
	if err := json.NewDecoder(req.Body).Decode(&body); err == nil {
		if v, ok := body["food_pickup_timeout"]; ok {
			var timeout int
			if json.Unmarshal(v, &timeout) == nil {
				r.state.mu.Lock()
				r.state.DeliverySettings["food_pickup_timeout"] = strconv.Itoa(timeout)
				r.state.mu.Unlock()
			}
		}
	}
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handlePutDeliveryTask(w http.ResponseWriter, req *http.Request) {
	cmd := chi.URLParam(req, "cmd")
	body := readBody(req)
	r.state.mu.Lock()
	defer r.state.mu.Unlock()

	switch cmd {
	case ":task_execution", "task_execution":
		var m map[string]json.RawMessage
		if err := json.Unmarshal(body, &m); err == nil {
			if v, ok := m["enable_task_execution"]; ok {
				var enabled bool
				if json.Unmarshal(v, &enabled) == nil {
					r.state.TaskExecutionEnabled = enabled
					log.Printf("MockRobot: task execution %v", enabled)
				}
			}
		}
	case ":start_pickup", "start_pickup":
		r.state.DeliveryStage = "ARRIVED_AT_TARGET"
		r.addEvent("delivery.start_pickup", "Pickup started")
	case ":end_pickup", "end_pickup":
		r.state.DeliveryStage = "ON_RETURNING"
		r.addEvent("delivery.end_pickup", "Pickup completed")
	}
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleGetCargos(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.Cargos
	r.state.mu.RUnlock()
	if v == nil {
		v = []Cargo{}
	}
	jsonResponse(w, v)
}

func (r *MockRobot) handlePutCargoBoxOp(w http.ResponseWriter, req *http.Request) {
	cargoID := chi.URLParam(req, "cargo_id")
	boxID := chi.URLParam(req, "box_id")
	op := chi.URLParam(req, "op")
	log.Printf("MockRobot: cargo op=%s cargoId=%s boxId=%s", op, cargoID, boxID)

	r.state.mu.Lock()
	for i, c := range r.state.Cargos {
		if c.ID == cargoID {
			for j, b := range c.Boxes {
				if b.ID == boxID {
					switch op {
					case "open":
						r.state.Cargos[i].Boxes[j].DoorStatus = "OPEN"
					case "close":
						r.state.Cargos[i].Boxes[j].DoorStatus = "CLOSED"
					case "lock":
						r.state.Cargos[i].Boxes[j].LockStatus = "LOCKED"
					case "unlock":
						r.state.Cargos[i].Boxes[j].LockStatus = "UNLOCKED"
					}
					break
				}
			}
			break
		}
	}
	r.state.mu.Unlock()
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleGetCargoOpResult(w http.ResponseWriter, req *http.Request) {
	cargoID := chi.URLParam(req, "cargo_id")
	status := CargoOperationStatus{
		Type:    "OPEN",
		Stage:   "DONE",
		CargoID: cargoID,
	}
	jsonResponse(w, status)
}

// ---------------------------------------------------------------------------
// External sensor handlers
// ---------------------------------------------------------------------------

func (r *MockRobot) handleFrontCam(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.currentVisibleFrontQrLocked()
	r.state.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, v)
}

func (r *MockRobot) handleBackCam(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.currentVisibleBackQrLocked()
	r.state.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, v)
}

func (r *MockRobot) handleCliffSafe(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	v := r.state.CliffSafe
	r.state.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, strconv.FormatBool(v))
}

func (r *MockRobot) handleSensorStatus(w http.ResponseWriter, req *http.Request) {
	r.state.mu.RLock()
	lidar := r.state.LidarOn
	front := r.state.FrontCamOn
	back := r.state.BackCamOn
	r.state.mu.RUnlock()
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "lidar=%v\nfront_cam=%v\nback_cam=%v", lidar, front, back)
}

func (r *MockRobot) handleLidarOn(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.LidarOn = true
	r.state.mu.Unlock()
	log.Printf("MockRobot: lidar on via actions API")
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleLidarOff(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.LidarOn = false
	r.state.mu.Unlock()
	log.Printf("MockRobot: lidar off via actions API")
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleFrontCamOn(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.FrontCamOn = true
	r.state.mu.Unlock()
	log.Printf("MockRobot: front_cam on")
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleFrontCamOff(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.FrontCamOn = false
	r.state.mu.Unlock()
	log.Printf("MockRobot: front_cam off")
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleBackCamOn(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.BackCamOn = true
	r.state.mu.Unlock()
	log.Printf("MockRobot: back_cam on")
	w.WriteHeader(http.StatusOK)
}

func (r *MockRobot) handleBackCamOff(w http.ResponseWriter, req *http.Request) {
	r.state.mu.Lock()
	r.state.BackCamOn = false
	r.state.mu.Unlock()
	log.Printf("MockRobot: back_cam off")
	w.WriteHeader(http.StatusOK)
}
