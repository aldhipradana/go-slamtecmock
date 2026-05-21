package main

import "sync"

// ---------------------------------------------------------------------------
// Primitive structs — JSON field names mirror the Slamtec REST API
// ---------------------------------------------------------------------------

type Pose struct {
	X     float32 `json:"x"`
	Y     float32 `json:"y"`
	Yaw   float32 `json:"yaw"`
	Pitch float32 `json:"pitch"`
	Roll  float32 `json:"roll"`
	Z     float32 `json:"z"`
}

func newPose(x, y, yaw float32) Pose {
	return Pose{X: x, Y: y, Yaw: yaw}
}

type MovementTarget struct {
	X      float32
	Y      float32
	Yaw    float32
	YawSet bool
}

// DockingPhaseEnum is the state of the two-phase MoveToTag docking process.
type DockingPhaseEnum int

const (
	DockingApproach DockingPhaseEnum = iota
	DockingDocking
	DockingComplete
)

type DockingPhaseState struct {
	Phase           DockingPhaseEnum
	ApproachX       float32
	ApproachY       float32
	TagX            float32
	TagY            float32
	TagYaw          float32
	BackwardDocking bool
	DockAllowance   float32
	TagIDs          []int
	TagType         int
	ReflectTagNum   int
}

type PowerStatus struct {
	BatteryPercentage int    `json:"batteryPercentage"`
	IsCharging        bool   `json:"isCharging"`
	IsDCConnected     bool   `json:"isDCConnected"`
	DockingStatus     string `json:"dockingStatus"` // "on_dock" | "not_on_dock" | "docking"
	PowerStage        string `json:"powerStage"`    // "running" | "sleeping"
	SleepMode         string `json:"sleepMode"`
}

func defaultPowerStatus() PowerStatus {
	return PowerStatus{
		BatteryPercentage: 85,
		DockingStatus:     "not_on_dock",
		PowerStage:        "running",
		SleepMode:         "awake",
	}
}

type NetworkStatus struct {
	Quality  int    `json:"quality"`
	SSID     string `json:"ssid"`
	IP       string `json:"ip"`
	SignalDB int    `json:"signal_db"`
}

func defaultNetworkStatus() NetworkStatus {
	return NetworkStatus{Quality: 90, SSID: "MockWifi", IP: "192.168.1.100", SignalDB: -55}
}

type RobotInfo struct {
	Model           string `json:"model"`
	SerialNumber    string `json:"serial_number"`
	FirmwareVersion string `json:"firmware_version"`
	HardwareVersion string `json:"hardware_version"`
	Manufacturer    string `json:"manufacturer"`
}

func defaultRobotInfo() RobotInfo {
	return RobotInfo{
		Model:           "MockRobot-4000",
		SerialNumber:    "MOCK-0001",
		FirmwareVersion: "4.6.0",
		HardwareVersion: "2.0",
		Manufacturer:    "MockTec",
	}
}

type HealthError struct {
	Description string `json:"description"`
	ErrorCode   int64  `json:"error_code"`
	Level       string `json:"level"` // "ERROR" | "WARNING" | "FATAL"
}

type RobotHealth struct {
	HasDepthCameraDisconnected bool          `json:"has_depth_camera_disconnected"`
	HasError                   bool          `json:"has_error"`
	HasFatal                   bool          `json:"has_fatal"`
	HasLidarDisconnected       bool          `json:"has_lidar_disconnected"`
	HasSdpDisconnected         bool          `json:"has_sdp_disconnected"`
	HasSystemEmergencyStop     bool          `json:"has_system_emergency_stop"`
	HasWarning                 bool          `json:"has_warning"`
	BaseError                  []HealthError `json:"base_error"`
}

type Capability struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type MapMeta struct {
	GridX      int     `json:"grid_x"`
	GridY      int     `json:"grid_y"`
	OriginX    float32 `json:"origin_x"`
	OriginY    float32 `json:"origin_y"`
	Resolution float32 `json:"resolution"`
}

func defaultMapMeta() MapMeta {
	return MapMeta{GridX: 200, GridY: 200, OriginX: -5, OriginY: -5, Resolution: 0.05}
}

// ActionState status / result constants
const (
	StatusNew     = 0
	StatusRunning = 1
	StatusPaused  = 3
	StatusDone    = 4
	ResultSuccess = 0
	ResultFailed  = -1
	ResultAborted = -2
)

type ActionState struct {
	Status int    `json:"status"`
	Result int    `json:"result"`
	Reason string `json:"reason"`
}

type ActionInfo struct {
	ActionID   int         `json:"action_id"`
	ActionName string      `json:"action_name"`
	Stage      string      `json:"stage"`
	State      ActionState `json:"state"`
}

type Floor struct {
	Building       string `json:"building"`
	FloorID        string `json:"floor"`
	IsDefaultFloor bool   `json:"is_default_floor"`
	Pose           *Pose  `json:"pose,omitempty"`
}

type ArtifactPoiMetadata struct {
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
}

type ArtifactPoi struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Metadata ArtifactPoiMetadata `json:"metadata"`
	Pose     Pose                `json:"pose"`
}

type MultiFloorPoi struct {
	ID          string `json:"id"`
	PoiName     string `json:"poi_name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Building    string `json:"building"`
	Floor       string `json:"floor"`
	Group       string `json:"group,omitempty"`
	Pose        Pose   `json:"pose"`
}

type Elevator struct {
	ElevatorID           string `json:"elevator_id"`
	DoorType             string `json:"door_type"`
	FrontSchedulingPoses []Pose `json:"front_scheduling_poses"`
	Optional             bool   `json:"optional"`
}

type MotionStrategy struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type RobotEvent struct {
	Type      string `json:"type"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

type CargoBox struct {
	ID          string `json:"id"`
	DoorStatus  string `json:"door_status"`
	LockStatus  string `json:"lock_status"`
	StockStatus string `json:"stock_status"`
	Status      string `json:"status"`
}

type Cargo struct {
	ID          string     `json:"id"`
	Orientation string     `json:"orientation"`
	Type        string     `json:"type"`
	Boxes       []CargoBox `json:"boxes"`
}

type CargoOperationStatus struct {
	Type    string    `json:"type"`
	Stage   string    `json:"stage"`
	CargoID string    `json:"cargo_id"`
	Box     *CargoBox `json:"box,omitempty"`
}

type JackStatus struct {
	ActualPos int `json:"actual_pos"`
	Alarm     int `json:"alarm"`
	DrvStatus int `json:"drv_status"`
	Stage     int `json:"stage"`
}

// ---------------------------------------------------------------------------
// RobotState — single source of truth for all mutable robot state
// ---------------------------------------------------------------------------

type RobotState struct {
	mu sync.RWMutex

	// sensor state
	LidarOn      bool
	FrontCamOn   bool
	BackCamOn    bool
	CliffSafe    bool
	FrontCamQrID string
	BackCamQrID  string

	// pose
	Pose Pose

	// movement
	MovementTarget *MovementTarget
	DockingPhase   *DockingPhaseState
	GoingHome      bool
	DockPose       *Pose
	HomePose       *Pose

	// power
	Battery PowerStatus

	// network
	Network NetworkStatus

	// robot info
	RobotInfo RobotInfo

	// health
	Health RobotHealth

	// capabilities
	Capabilities []Capability

	// system parameters
	SystemParameters map[string]string

	// mapping
	MappingEnabled bool
	MapMeta        MapMeta

	// artifact POIs
	ArtifactPois []ArtifactPoi

	// multi-floor POIs
	MultiFloorPois []MultiFloorPoi

	// floors
	Floors       []Floor
	CurrentFloor Floor

	// elevators
	Elevators []Elevator

	// actions
	CurrentAction *ActionInfo
	ActionHistory map[int]*ActionInfo

	// motion strategies
	MotionStrategies      []MotionStrategy
	CurrentMotionStrategy MotionStrategy

	// supported actions
	SupportedActions []string

	// events ring buffer
	Events []RobotEvent

	// delivery
	DeliveryStage        string
	TaskExecutionEnabled bool
	DeliverySettings     map[string]string
	Cargos               []Cargo

	// soft brake
	SoftBrakeActive     bool
	PhysicalEStopActive bool

	// jack (lift) state
	JackActualPos int
	JackStage     int
	JackAlarm     int
	JackDrvStatus int
	JackCommand   string // "Up", "Down", "Stop", "ClearAlarm", or ""

	// internal timing
	LastLowBatteryEventMs int64
}

func defaultRobotState() *RobotState {
	s := &RobotState{
		CliffSafe: true,
		Pose:      newPose(1.0, 2.0, 0.0),
		DockPose:  posePtr(0, 0, 0),
		HomePose:  posePtr(564, 406, 0),
		Battery:   defaultPowerStatus(),
		Network:   defaultNetworkStatus(),
		RobotInfo: defaultRobotInfo(),
		Health:    RobotHealth{BaseError: []HealthError{}},
		MapMeta:   defaultMapMeta(),
		SystemParameters: map[string]string{
			"max_speed":          "0.8",
			"min_speed":          "0.1",
			"obstacle_avoidance": "true",
		},
		Capabilities: []Capability{
			{Name: "slam", Enabled: true},
			{Name: "motion", Enabled: true},
			{Name: "delivery", Enabled: true},
		},
		SupportedActions: []string{
			"slamtec.agent.actions.MoveToAction",
			"slamtec.agent.actions.MultiFloorMoveAction",
			"slamtec.agent.actions.MultiFloorBackHomeAction",
			"slamtec.agent.actions.GoHomeAction",
			"slamtec.agent.actions.RotateAction",
			"slamtec.agent.actions.RecoverLocalizationAction",
			"slamtec.agent.actions.ManualRelocalizationAction",
		},
		MotionStrategies: []MotionStrategy{
			{Name: "default", Description: "Default navigation strategy"},
			{Name: "aggressive", Description: "Faster speed, tighter obstacle avoidance"},
			{Name: "conservative", Description: "Slower speed, wider clearance"},
		},
		CurrentMotionStrategy: MotionStrategy{Name: "default", Description: "Default navigation strategy"},
		ActionHistory:         map[int]*ActionInfo{},
		Events:                []RobotEvent{},
		DeliveryStage:         "IDLE",
		TaskExecutionEnabled:  true,
		DeliverySettings: map[string]string{
			"food_pickup_timeout": "60",
			"low_battery_action":  "go_home",
		},
		Cargos:        []Cargo{},
		JackActualPos: 31,
		JackStage:     2,
		JackDrvStatus: 17975,
	}
	s.ArtifactPois = defaultArtifactPois()
	s.MultiFloorPois = defaultMultiFloorPois()
	s.Floors = defaultFloors()
	s.CurrentFloor = Floor{Building: "Kitchen", FloorID: "1", IsDefaultFloor: true}
	s.Elevators = []Elevator{defaultElevatorPose("elev-1", 4.0, -1.0)}
	return s
}

func posePtr(x, y, yaw float32) *Pose {
	p := newPose(x, y, yaw)
	return &p
}

func defaultArtifactPois() []ArtifactPoi {
	return []ArtifactPoi{
		makeArtifactPoi("poi-lobby", "Lobby", "navigation", 5, 0, 0),
		makeArtifactPoi("poi-dock", "Dock", "PARKING", 0, 0, 0),
	}
}

func makeArtifactPoi(id, name, typ string, x, y, yaw float32) ArtifactPoi {
	return ArtifactPoi{
		ID:   id,
		Type: typ,
		Metadata: ArtifactPoiMetadata{
			DisplayName: name,
			Type:        typ,
		},
		Pose: newPose(x, y, yaw),
	}
}

func makeMFPoi(id, name, typ, building, floor string, x, y, yaw float32) MultiFloorPoi {
	return MultiFloorPoi{
		ID:          id,
		PoiName:     name,
		DisplayName: name,
		Type:        typ,
		Building:    building,
		Floor:       floor,
		Pose:        newPose(x, y, yaw),
	}
}

func defaultMultiFloorPois() []MultiFloorPoi {
	return []MultiFloorPoi{
		// Kitchen layout
		makeMFPoi("KITCHEN_HOME_1", "Kitchen Home", "INTERSECTION", "Kitchen", "1", 457, 282, 0),
		makeMFPoi("KITCHEN_HOME_2", "Kitchen Home 2", "INTERSECTION", "Kitchen", "1", 459, 222, 0),
		makeMFPoi("KITCHEN_VEHICLE_PORT_1", "Kitchen Vehicle Port 1", "INTERSECTION", "Kitchen", "1", 441, 199, 0),
		makeMFPoi("KITCHEN_LOADING_1_APPROACH", "Front of Kitchen Loading 1", "INTERSECTION", "Kitchen", "1", 505, 279, 0),
		makeMFPoi("KITCHEN_LOADING_1", "Kitchen Loading 1", "TROLLEY_POINT", "Kitchen", "1", 533, 277, 0),
		makeMFPoi("KITCHEN_LOADING_2_APPROACH", "Front of Kitchen Loading 2", "INTERSECTION", "Kitchen", "1", 414, 280, 0),
		makeMFPoi("KITCHEN_LOADING_2", "Kitchen Loading 2", "TROLLEY_POINT", "Kitchen", "1", 386, 278, 0),
		makeMFPoi("KITCHEN_LOADING_3_APPROACH", "Front of Kitchen Loading 3", "INTERSECTION", "Kitchen", "1", 416, 222, 0),
		makeMFPoi("KITCHEN_LOADING_3", "Kitchen Loading 3", "TROLLEY_POINT", "Kitchen", "1", 388, 222, 0),
		makeMFPoi("KITCHEN_LOADING_4_APPROACH", "Front of Kitchen Loading 4", "INTERSECTION", "Kitchen", "1", 506, 220, 0),
		makeMFPoi("KITCHEN_LOADING_4", "Kitchen Loading 4", "TROLLEY_POINT", "Kitchen", "1", 534, 218, 0),
		makeMFPoi("KITCHEN_VEHICLE_1_APPROACH", "Front of Kitchen Vehicle 1", "INTERSECTION", "Kitchen", "1", 442, 50, 0),
		makeMFPoi("KITCHEN_VEHICLE_1", "Kitchen Vehicle 1", "AV_LOADING_POINT", "Kitchen", "1", 442, 32, 0),
		makeMFPoi("KITCHEN_VEHICLE_2_APPROACH", "Front of Kitchen Vehicle 2", "INTERSECTION", "Kitchen", "1", 442, 90, 0),
		makeMFPoi("KITCHEN_VEHICLE_2", "Kitchen Vehicle 2", "AV_LOADING_POINT", "Kitchen", "1", 442, 73, 0),
		makeMFPoi("KITCHEN_VEHICLE_3_APPROACH", "Front of Kitchen Vehicle 3", "INTERSECTION", "Kitchen", "1", 440, 130, 0),
		makeMFPoi("KITCHEN_VEHICLE_3", "Kitchen Vehicle 3", "AV_LOADING_POINT", "Kitchen", "1", 440, 113, 0),
		makeMFPoi("KITCHEN_VEHICLE_4_APPROACH", "Front of Kitchen Vehicle 4", "INTERSECTION", "Kitchen", "1", 440, 180, 0),
		makeMFPoi("KITCHEN_VEHICLE_4", "Kitchen Vehicle 4", "AV_LOADING_POINT", "Kitchen", "1", 439, 163, 0),
		makeMFPoi("KITCHEN_CHARGING_DOCK_POINT", "Kitchen Charging Dock", "CHARGER_DOCK_POINT", "Kitchen", "1", 564, 406, 0),
		makeMFPoi("KITCHEN_EMERGENCY_POINT", "Kitchen Emergency Point", "EMERGENCY_POINT", "Kitchen", "1", 300, 350, 0),
		// Institution Ground
		makeMFPoi("INSTITUTION_GROUND_HOME", "Ground Home", "CHARGER_DOCK_POINT", "Institution", "1", 507, 361, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_1_APPROACH", "Front of Ground Loading 1", "INTERSECTION", "Institution", "1", 726, 155, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_1", "Ground Loading 1", "TROLLEY_POINT", "Institution", "1", 726, 122, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_2_APPROACH", "Front of Ground Loading 2", "INTERSECTION", "Institution", "1", 633, 155, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_2", "Ground Loading 2", "TROLLEY_POINT", "Institution", "1", 633, 122, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_3_APPROACH", "Front of Ground Loading 3", "INTERSECTION", "Institution", "1", 543, 155, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_3", "Ground Loading 3", "TROLLEY_POINT", "Institution", "1", 543, 122, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_4_APPROACH", "Front of Ground Loading 4", "INTERSECTION", "Institution", "1", 459, 155, 0),
		makeMFPoi("INSTITUTION_GROUND_LOADING_4", "Ground Loading 4", "TROLLEY_POINT", "Institution", "1", 459, 122, 0),
		makeMFPoi("PORT_INSTITUTION_GROUND_LOADING_1", "Port Ground Loading 1", "INTERSECTION", "Institution", "1", 726, 187, 0),
		makeMFPoi("PORT_INSTITUTION_GROUND_LOADING_2", "Port Ground Loading 2", "INTERSECTION", "Institution", "1", 632, 187, 0),
		makeMFPoi("PORT_INSTITUTION_GROUND_LOADING_3", "Port Ground Loading 3", "INTERSECTION", "Institution", "1", 543, 187, 0),
		makeMFPoi("PORT_INSTITUTION_GROUND_LOADING_4", "Port Ground Loading 4", "INTERSECTION", "Institution", "1", 459, 187, 0),
		makeMFPoi("PORT_INSTITUTION_GROUND_VEHICLE_1", "Port Ground Vehicle 1", "INTERSECTION", "Institution", "1", 682, 187, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_1_APPROACH", "Front of Ground Vehicle 1", "INTERSECTION", "Institution", "1", 682, 397, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_1", "Ground Vehicle 1", "AV_LOADING_POINT", "Institution", "1", 682, 421, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_2_APPROACH", "Front of Ground Vehicle 2", "INTERSECTION", "Institution", "1", 682, 337, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_2", "Ground Vehicle 2", "AV_LOADING_POINT", "Institution", "1", 682, 361, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_3_APPROACH", "Front of Ground Vehicle 3", "INTERSECTION", "Institution", "1", 682, 277, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_3", "Ground Vehicle 3", "AV_LOADING_POINT", "Institution", "1", 682, 301, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_4_APPROACH", "Front of Ground Vehicle 4", "INTERSECTION", "Institution", "1", 682, 218, 0),
		makeMFPoi("INSTITUTION_GROUND_VEHICLE_4", "Ground Vehicle 4", "AV_LOADING_POINT", "Institution", "1", 682, 242, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_1_APPROACH", "Front of Ground Lift 1", "INTERSECTION", "Institution", "1", 256, 181, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_1", "Ground Lift 1", "LIFT_TROLLEY_POINT", "Institution", "1", 260, 149, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_2_APPROACH", "Front of Ground Lift 2", "INTERSECTION", "Institution", "1", 211, 183, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_2", "Ground Lift 2", "LIFT_TROLLEY_POINT", "Institution", "1", 207, 151, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_3_APPROACH", "Front of Ground Lift 3", "INTERSECTION", "Institution", "1", 256, 242, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_3", "Ground Lift 3", "LIFT_TROLLEY_POINT", "Institution", "1", 263, 211, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_4_APPROACH", "Front of Ground Lift 4", "INTERSECTION", "Institution", "1", 215, 247, 0),
		makeMFPoi("INSTITUTION_GROUND_LIFT_4", "Ground Lift 4", "LIFT_TROLLEY_POINT", "Institution", "1", 211, 215, 0),
		makeMFPoi("CARGO_LIFT_ENTRY_GROUND", "Cargo Lift Entry Ground", "LIFT_WAITING_POINT", "Institution", "1", 231, 361, 0),
		makeMFPoi("GROUND_EMERGENCY_POINT", "Ground Emergency Point", "EMERGENCY_POINT", "Institution", "1", 400, 300, 0),
		// Institution Level 5
		makeMFPoi("INSTITUTION_TOP_LOADING_1_APPROACH", "Front of Top Loading 1", "INTERSECTION", "Institution", "5", 1107, 185, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_1", "Top Loading 1", "TROLLEY_POINT", "Institution", "5", 1107, 129, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_2_APPROACH", "Front of Top Loading 2", "INTERSECTION", "Institution", "5", 937, 185, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_2", "Top Loading 2", "TROLLEY_POINT", "Institution", "5", 937, 129, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_3_APPROACH", "Front of Top Loading 3", "INTERSECTION", "Institution", "5", 753, 185, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_3", "Top Loading 3", "TROLLEY_POINT", "Institution", "5", 753, 126, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_4_APPROACH", "Front of Top Loading 4", "INTERSECTION", "Institution", "5", 564, 185, 0),
		makeMFPoi("INSTITUTION_TOP_LOADING_4", "Top Loading 4", "TROLLEY_POINT", "Institution", "5", 564, 128, 0),
		makeMFPoi("PORT_INSTITUTION_TOP_LOADING_1", "Port Top Loading 1", "INTERSECTION", "Institution", "5", 1107, 255, 0),
		makeMFPoi("PORT_INSTITUTION_TOP_LOADING_2", "Port Top Loading 2", "INTERSECTION", "Institution", "5", 937, 255, 0),
		makeMFPoi("PORT_INSTITUTION_TOP_LOADING_3", "Port Top Loading 3", "INTERSECTION", "Institution", "5", 753, 255, 0),
		makeMFPoi("PORT_INSTITUTION_TOP_LOADING_4", "Port Top Loading 4", "INTERSECTION", "Institution", "5", 564, 255, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_1_APPROACH", "Front of Top Lift 1", "INTERSECTION", "Institution", "5", 580, 536, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_1", "Top Lift 1", "LIFT_TROLLEY_POINT", "Institution", "5", 593, 590, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_2_APPROACH", "Front of Top Lift 2", "INTERSECTION", "Institution", "5", 473, 543, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_2", "Top Lift 2", "LIFT_TROLLEY_POINT", "Institution", "5", 461, 598, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_3_APPROACH", "Front of Top Lift 3", "INTERSECTION", "Institution", "5", 559, 362, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_3", "Top Lift 3", "LIFT_TROLLEY_POINT", "Institution", "5", 589, 409, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_4_APPROACH", "Front of Top Lift 4", "INTERSECTION", "Institution", "5", 494, 363, 0),
		makeMFPoi("INSTITUTION_TOP_LIFT_4", "Top Lift 4", "LIFT_TROLLEY_POINT", "Institution", "5", 466, 412, 0),
		makeMFPoi("CARGO_LIFT_EXIT_LEVEL5", "Cargo Lift Exit Level 5", "LIFT_WAITING_POINT", "Institution", "5", 525, 308, 0),
		makeMFPoi("DESTINATION_AREA", "Destination Area", "INTERSECTION", "Institution", "5", 390, 259, 0),
		makeMFPoi("LEVEL5_CHARGER", "Level 5 Robot Charger", "CHARGER_DOCK_POINT", "Institution", "5", 421, 38, 0),
		makeMFPoi("LEVEL5_EMERGENCY_POINT", "Level 5 Emergency Point", "EMERGENCY_POINT", "Institution", "5", 700, 400, 0),
		// AV Map
		makeMFPoi("AV_HOME", "AV Home", "CHARGER_DOCK_POINT", "AVMap", "1", 292, 396, 0),
		makeMFPoi("KITCHEN_ENTRANCE", "Kitchen Entrance", "WAYPOINT", "AVMap", "1", 267, 289, 0),
		makeMFPoi("KITCHEN_EXIT", "Kitchen Exit", "WAYPOINT", "AVMap", "1", 231, 256, 0),
		makeMFPoi("KITCHEN_DOCK", "Kitchen Dock", "AV_LOADING_POINT", "AVMap", "1", 245, 298, 0),
		makeMFPoi("INSTITUTION_DOCK", "Institution Dock", "AV_LOADING_POINT", "AVMap", "1", 262, 102, 0),
		makeMFPoi("PRISON_GATE", "Prison Gate", "GATE", "AVMap", "1", 232, 91, 0),
		makeMFPoi("AV_MAP_EMERGENCY_POINT", "AV Map Emergency Point", "EMERGENCY_POINT", "AVMap", "1", 260, 200, 0),
	}
}

func defaultFloors() []Floor {
	return []Floor{
		{Building: "Kitchen", FloorID: "1", IsDefaultFloor: true},
		{Building: "Institution", FloorID: "1"},
		{Building: "Institution", FloorID: "5"},
		{Building: "AVMap", FloorID: "1"},
	}
}

func defaultElevatorPose(id string, frontX, frontY float32) Elevator {
	return Elevator{
		ElevatorID:           id,
		DoorType:             "single",
		Optional:             false,
		FrontSchedulingPoses: []Pose{newPose(frontX, frontY, 0)},
	}
}
