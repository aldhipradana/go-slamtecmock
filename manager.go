package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ---------------------------------------------------------------------------
// RobotManager
// ---------------------------------------------------------------------------

type RobotManager struct {
	mu     sync.RWMutex
	robots map[int]*MockRobot
	nextID atomic.Int32
}

func NewRobotManager() *RobotManager {
	m := &RobotManager{
		robots: make(map[int]*MockRobot),
	}
	// Default Robot #1
	m.createRobot(defaultRobotState())
	return m
}

// createRobot allocates the next ID, starts the robot, stores it.
func (m *RobotManager) createRobot(state *RobotState) *MockRobot {
	id := int(m.nextID.Add(1))
	rb := newRobot(id, state)
	m.mu.Lock()
	m.robots[id] = rb
	m.mu.Unlock()
	log.Printf("RobotManager: created Robot #%d", id)
	return rb
}

// deleteRobot stops and removes the robot. Returns false if not found.
func (m *RobotManager) deleteRobot(id int) bool {
	m.mu.Lock()
	rb, ok := m.robots[id]
	if ok {
		delete(m.robots, id)
	}
	m.mu.Unlock()
	if ok {
		rb.Stop()
		log.Printf("RobotManager: deleted Robot #%d", id)
	}
	return ok
}

// getRobot returns the robot for the given id, or nil.
func (m *RobotManager) getrobot(id int) *MockRobot {
	m.mu.RLock()
	rb := m.robots[id]
	m.mu.RUnlock()
	return rb
}

// listRobots returns all robots sorted by ID.
func (m *RobotManager) listRobots() []*MockRobot {
	m.mu.RLock()
	list := make([]*MockRobot, 0, len(m.robots))
	for _, rb := range m.robots {
		list = append(list, rb)
	}
	m.mu.RUnlock()
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list
}

// ---------------------------------------------------------------------------
// withRobot middleware — extracts {robotId}, looks up robot, calls fn
// ---------------------------------------------------------------------------

type robotHandlerFunc func(rb *MockRobot, w http.ResponseWriter, req *http.Request)

func (m *RobotManager) withRobot(fn robotHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		idStr := chi.URLParam(req, "robotId")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, `{"error":"invalid robot id"}`, http.StatusBadRequest)
			return
		}
		rb := m.getrobot(id)
		if rb == nil {
			http.Error(w, fmt.Sprintf(`{"error":"robot %d not found"}`, id), http.StatusNotFound)
			return
		}
		fn(rb, w, req)
	}
}

// ---------------------------------------------------------------------------
// BuildRouter — single router for the entire server
// ---------------------------------------------------------------------------

func (m *RobotManager) BuildRouter() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	// ── Dashboard + HTMX UI ────────────────────────────────────────────────
	r.Get("/", m.handleDashboard)
	r.Get("/ui/robots", m.handleGetRobotListUI)
	r.Post("/ui/robots", m.handleCreateRobotUI)
	r.Delete("/ui/robots/{id}", m.handleDeleteRobotUI)
	r.Get("/ui/robots/{id}/status", m.handleGetRobotStatusUI)
	r.Post("/ui/robots/{id}/action", m.handleUIAction)
	r.Post("/ui/robots/{id}/jack", m.handleUIJack)
	r.Post("/ui/robots/{id}/brake", m.handleUIBrake)
	r.Post("/ui/robots/{id}/sensor", m.handleUISensor)

	// ── Management REST API ────────────────────────────────────────────────
	r.Get("/robots", m.handleListRobots)
	r.Post("/robots", m.handleCreateRobot)
	r.Delete("/robots/{id}", m.handleDeleteRobotAPI)

	// ── Per-robot Slamtec sub-router  /robot/{robotId}/... ─────────────────
	r.Route("/robot/{robotId}", func(r chi.Router) {
		m.registerSlamtecRoutes(r)
	})

	// ── Backward-compat: /api/... → Robot #1 ──────────────────────────────
	r.Mount("/api", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		rb := m.getrobot(1)
		if rb == nil {
			http.Error(w, `{"error":"default robot not found"}`, http.StatusServiceUnavailable)
			return
		}
		// Re-dispatch through a temporary sub-router bound to Robot #1
		sr := chi.NewRouter()
		bindSlamtecRoutes(sr, rb)
		sr.ServeHTTP(w, req)
	}))

	return r
}

// registerSlamtecRoutes mounts all Slamtec routes on r using m.withRobot.
func (m *RobotManager) registerSlamtecRoutes(r chi.Router) {
	// System
	r.Get("/api/core/system/v1/power/status", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetPowerStatus(w, req) }))
	r.Get("/api/core/system/v1/network/status", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetNetworkStatus(w, req) }))
	r.Get("/api/core/system/v1/robot/info", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetRobotInfo(w, req) }))
	r.Get("/api/core/system/v1/robot/health", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetRobotHealth(w, req) }))
	r.Get("/api/core/system/v1/capabilities", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetCapabilities(w, req) }))
	r.Get("/api/core/system/v1/parameter", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetParameter(w, req) }))
	r.Put("/api/core/system/v1/parameter", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutParameter(w, req) }))
	r.Post("/api/core/system/v1/power/{cmd}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePowerCmd(w, req) }))
	r.Get("/api/core/system/v1/jack/status", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetJackStatus(w, req) }))
	r.Post("/api/core/system/v1/jack/status", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePostJackStatus(w, req) }))
	// SLAM
	r.Get("/api/core/slam/v1/localization/pose", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetPose(w, req) }))
	r.Get("/api/core/slam/v1/homepose", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetHomePose(w, req) }))
	r.Put("/api/core/slam/v1/homepose", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutHomePose(w, req) }))
	r.Put("/api/multi-floor/localization/v1/pose", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutPose(w, req) }))
	r.Get("/api/core/slam/v1/maps/explore", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetMapExplore(w, req) }))
	r.Get("/api/core/slam/v1/maps/stcm", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetMapStcm(w, req) }))
	r.Put("/api/core/slam/v1/mapping/{enable}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutMapping(w, req) }))
	// Artifact POIs
	r.Get("/api/core/artifact/v1/pois", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetArtifactPois(w, req) }))
	r.Post("/api/core/artifact/v1/pois", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePostArtifactPoi(w, req) }))
	r.Delete("/api/core/artifact/v1/pois/{id}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleDeleteArtifactPoi(w, req) }))
	// Motion
	r.Get("/api/core/motion/v1/action-factories", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetActionFactories(w, req) }))
	r.Get("/api/core/motion/v1/strategies", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetStrategies(w, req) }))
	r.Get("/api/core/motion/v1/strategies/{id}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetCurrentStrategy(w, req) }))
	r.Put("/api/core/motion/v1/strategies/{id}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutStrategy(w, req) }))
	r.Post("/api/core/slam/v1/action/create", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleCreateAction(w, req) }))
	r.Post("/api/core/motion/v1/actions", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleCreateAction(w, req) }))
	r.Get("/api/core/motion/v1/actions/{actionId}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetAction(w, req) }))
	r.Delete("/api/core/motion/v1/actions/{actionId}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleDeleteAction(w, req) }))
	// Multi-floor
	r.Post("/api/multi-floor/motion/v1/movetoaction", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleCreateAction(w, req) }))
	r.Get("/api/multi-floor/map/v1/floors", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetFloors(w, req) }))
	r.Get("/api/multi-floor/map/v1/floors/{floorId}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetCurrentFloor(w, req) }))
	r.Put("/api/multi-floor/map/v1/floors/{floorId}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutFloor(w, req) }))
	r.Get("/api/multi-floor/map/v1/pois", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetMultiFloorPois(w, req) }))
	r.Get("/api/multi-floor/map/v1/elevators/{elevatorId}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetElevator(w, req) }))
	// Events
	r.Get("/api/platform/v1/events", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetEvents(w, req) }))
	// Delivery
	r.Get("/api/delivery/v1/stage", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetDeliveryStage(w, req) }))
	r.Get("/api/delivery/v1/settings", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetDeliverySettings(w, req) }))
	r.Put("/api/delivery/v1/settings/timeout", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutDeliveryTimeout(w, req) }))
	r.Put("/api/delivery/v1/tasks/{cmd}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutDeliveryTask(w, req) }))
	r.Get("/api/delivery/v1/cargos", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetCargos(w, req) }))
	r.Put("/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handlePutCargoBoxOp(w, req) }))
	r.Get("/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/operation_result", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleGetCargoOpResult(w, req) }))
	// External sensors
	r.Get("/front_cam", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleFrontCam(w, req) }))
	r.Get("/back_cam", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleBackCam(w, req) }))
	r.Get("/cliff_safe", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleCliffSafe(w, req) }))
	r.Get("/sensor/status", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleSensorStatus(w, req) }))
	r.Post("/sensor/lidar/on", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleLidarOn(w, req) }))
	r.Post("/sensor/lidar/off", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleLidarOff(w, req) }))
	r.Post("/sensor/front_cam/on", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleFrontCamOn(w, req) }))
	r.Post("/sensor/front_cam/off", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleFrontCamOff(w, req) }))
	r.Post("/sensor/back_cam/on", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleBackCamOn(w, req) }))
	r.Post("/sensor/back_cam/off", m.withRobot(func(rb *MockRobot, w http.ResponseWriter, req *http.Request) { rb.handleBackCamOff(w, req) }))
}

// bindSlamtecRoutes binds all routes directly to a specific robot (used for /api compat alias).
func bindSlamtecRoutes(r chi.Router, rb *MockRobot) {
	r.Get("/core/system/v1/power/status", rb.handleGetPowerStatus)
	r.Get("/core/system/v1/network/status", rb.handleGetNetworkStatus)
	r.Get("/core/system/v1/robot/info", rb.handleGetRobotInfo)
	r.Get("/core/system/v1/robot/health", rb.handleGetRobotHealth)
	r.Get("/core/system/v1/capabilities", rb.handleGetCapabilities)
	r.Get("/core/system/v1/parameter", rb.handleGetParameter)
	r.Put("/core/system/v1/parameter", rb.handlePutParameter)
	r.Post("/core/system/v1/power/{cmd}", rb.handlePowerCmd)
	r.Get("/core/system/v1/jack/status", rb.handleGetJackStatus)
	r.Post("/core/system/v1/jack/status", rb.handlePostJackStatus)
	r.Get("/core/slam/v1/localization/pose", rb.handleGetPose)
	r.Get("/core/slam/v1/homepose", rb.handleGetHomePose)
	r.Put("/core/slam/v1/homepose", rb.handlePutHomePose)
	r.Put("/multi-floor/localization/v1/pose", rb.handlePutPose)
	r.Get("/core/slam/v1/maps/explore", rb.handleGetMapExplore)
	r.Get("/core/slam/v1/maps/stcm", rb.handleGetMapStcm)
	r.Put("/core/slam/v1/mapping/{enable}", rb.handlePutMapping)
	r.Get("/core/artifact/v1/pois", rb.handleGetArtifactPois)
	r.Post("/core/artifact/v1/pois", rb.handlePostArtifactPoi)
	r.Delete("/core/artifact/v1/pois/{id}", rb.handleDeleteArtifactPoi)
	r.Get("/core/motion/v1/action-factories", rb.handleGetActionFactories)
	r.Get("/core/motion/v1/strategies", rb.handleGetStrategies)
	r.Get("/core/motion/v1/strategies/{id}", rb.handleGetCurrentStrategy)
	r.Put("/core/motion/v1/strategies/{id}", rb.handlePutStrategy)
	r.Post("/core/slam/v1/action/create", rb.handleCreateAction)
	r.Post("/core/motion/v1/actions", rb.handleCreateAction)
	r.Get("/core/motion/v1/actions/{actionId}", rb.handleGetAction)
	r.Delete("/core/motion/v1/actions/{actionId}", rb.handleDeleteAction)
	r.Post("/multi-floor/motion/v1/movetoaction", rb.handleCreateAction)
	r.Get("/multi-floor/map/v1/floors", rb.handleGetFloors)
	r.Get("/multi-floor/map/v1/floors/{floorId}", rb.handleGetCurrentFloor)
	r.Put("/multi-floor/map/v1/floors/{floorId}", rb.handlePutFloor)
	r.Get("/multi-floor/map/v1/pois", rb.handleGetMultiFloorPois)
	r.Get("/multi-floor/map/v1/elevators/{elevatorId}", rb.handleGetElevator)
	r.Get("/platform/v1/events", rb.handleGetEvents)
	r.Get("/delivery/v1/stage", rb.handleGetDeliveryStage)
	r.Get("/delivery/v1/settings", rb.handleGetDeliverySettings)
	r.Put("/delivery/v1/settings/timeout", rb.handlePutDeliveryTimeout)
	r.Put("/delivery/v1/tasks/{cmd}", rb.handlePutDeliveryTask)
	r.Get("/delivery/v1/cargos", rb.handleGetCargos)
	r.Put("/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}", rb.handlePutCargoBoxOp)
	r.Get("/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/operation_result", rb.handleGetCargoOpResult)
}

// ---------------------------------------------------------------------------
// CORS middleware
// ---------------------------------------------------------------------------

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
// Management REST API handlers
// ---------------------------------------------------------------------------

type robotSummary struct {
	ID            int     `json:"id"`
	PoseX         float32 `json:"pose_x"`
	PoseY         float32 `json:"pose_y"`
	BatteryPct    int     `json:"battery_pct"`
	IsCharging    bool    `json:"is_charging"`
	ActionStatus  string  `json:"action_status"`
}

func (m *RobotManager) handleListRobots(w http.ResponseWriter, req *http.Request) {
	list := m.listRobots()
	summaries := make([]robotSummary, 0, len(list))
	for _, rb := range list {
		rb.state.mu.RLock()
		status := "IDLE"
		if rb.state.CurrentAction != nil {
			switch rb.state.CurrentAction.State.Status {
			case StatusRunning:
				status = "RUNNING"
			case StatusDone:
				status = "DONE"
			}
		}
		s := robotSummary{
			ID:           rb.ID,
			PoseX:        rb.state.Pose.X,
			PoseY:        rb.state.Pose.Y,
			BatteryPct:   rb.state.Battery.BatteryPercentage,
			IsCharging:   rb.state.Battery.IsCharging,
			ActionStatus: status,
		}
		rb.state.mu.RUnlock()
		summaries = append(summaries, s)
	}
	jsonResponse(w, summaries)
}

func (m *RobotManager) handleCreateRobot(w http.ResponseWriter, req *http.Request) {
	state := defaultRobotState()
	// Optional init body: {"x":0,"y":0,"yaw":0,"battery":85}
	var init struct {
		X       *float32 `json:"x"`
		Y       *float32 `json:"y"`
		Yaw     *float32 `json:"yaw"`
		Battery *int     `json:"battery"`
	}
	if err := json.NewDecoder(req.Body).Decode(&init); err == nil {
		if init.X != nil {
			state.Pose.X = *init.X
		}
		if init.Y != nil {
			state.Pose.Y = *init.Y
		}
		if init.Yaw != nil {
			state.Pose.Yaw = *init.Yaw
		}
		if init.Battery != nil {
			state.Battery.BatteryPercentage = *init.Battery
		}
	}
	rb := m.createRobot(state)
	w.WriteHeader(http.StatusCreated)
	jsonResponse(w, map[string]int{"id": rb.ID})
}

func (m *RobotManager) handleDeleteRobotAPI(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || !m.deleteRobot(id) {
		http.Error(w, `{"error":"robot not found"}`, http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

