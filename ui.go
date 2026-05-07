package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ---------------------------------------------------------------------------
// View model
// ---------------------------------------------------------------------------

type RobotCardData struct {
	ID            int
	Label         string
	PoseX         float32
	PoseY         float32
	PoseYaw       float32
	BatteryPct    int
	BatteryClass  string // "bat-high" | "bat-mid" | "bat-low"
	IsCharging    bool
	DockingStatus string
	HasEmergency  bool
	SoftBrake     bool
	ActionStatus  string // "IDLE" | "RUNNING" | "DONE"
	ActionName    string
	ActionStage   string
	DeliveryStage string
	Floor         string
	JackStage     int
	JackPos       int
	LidarOn       bool
	FrontCamOn    bool
	BackCamOn     bool
	CliffSafe     bool
	Events        []RobotEvent
}

type dashboardData struct {
	Robots []RobotCardData
}

// buildCardData reads all display fields from a robot under RLock.
func buildCardData(rb *MockRobot) RobotCardData {
	s := rb.state
	s.mu.RLock()
	defer s.mu.RUnlock()

	actionStatus := "IDLE"
	actionName := ""
	actionStage := ""
	if s.CurrentAction != nil {
		actionName = s.CurrentAction.ActionName
		actionStage = s.CurrentAction.Stage
		switch s.CurrentAction.State.Status {
		case StatusRunning:
			actionStatus = "RUNNING"
		case StatusDone:
			actionStatus = "DONE"
		default:
			actionStatus = "NEW"
		}
	}

	batClass := "bat-high"
	if s.Battery.BatteryPercentage < 20 {
		batClass = "bat-low"
	} else if s.Battery.BatteryPercentage < 50 {
		batClass = "bat-mid"
	}

	events := s.Events
	if len(events) > 5 {
		events = events[:5]
	}

	floor := fmt.Sprintf("%s / %s", s.CurrentFloor.Building, s.CurrentFloor.FloorID)

	// Shorten action name for display
	displayName := actionName
	if idx := strings.LastIndex(displayName, "."); idx >= 0 {
		displayName = displayName[idx+1:]
	}

	return RobotCardData{
		ID:            rb.ID,
		Label:         fmt.Sprintf("Robot #%d", rb.ID),
		PoseX:         s.Pose.X,
		PoseY:         s.Pose.Y,
		PoseYaw:       s.Pose.Yaw,
		BatteryPct:    s.Battery.BatteryPercentage,
		BatteryClass:  batClass,
		IsCharging:    s.Battery.IsCharging,
		DockingStatus: s.Battery.DockingStatus,
		HasEmergency:  s.Health.HasSystemEmergencyStop,
		SoftBrake:     s.SoftBrakeActive,
		ActionStatus:  actionStatus,
		ActionName:    displayName,
		ActionStage:   actionStage,
		DeliveryStage: s.DeliveryStage,
		Floor:         floor,
		JackStage:     s.JackStage,
		JackPos:       s.JackActualPos,
		LidarOn:       s.LidarOn,
		FrontCamOn:    s.FrontCamOn,
		BackCamOn:     s.BackCamOn,
		CliffSafe:     s.CliffSafe,
		Events:        events,
	}
}

// ---------------------------------------------------------------------------
// Templates — defined inline (no embed needed, no external files to manage)
// ---------------------------------------------------------------------------

var pageTmpl = template.Must(template.New("page").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>Slamtec Mock Manager</title>
<script src="https://unpkg.com/htmx.org@2.0.4/dist/htmx.min.js"></script>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:system-ui,sans-serif;background:#0f172a;color:#e2e8f0;min-height:100vh}
header{background:#1e293b;padding:1rem 1.5rem;display:flex;align-items:center;gap:1rem;border-bottom:1px solid #334155}
header h1{font-size:1.25rem;font-weight:700;color:#f8fafc}
header h1 span{color:#38bdf8}
.add-btn{margin-left:auto;background:#0ea5e9;color:#fff;border:none;padding:.5rem 1.2rem;border-radius:.5rem;cursor:pointer;font-size:.9rem;font-weight:600;transition:background .2s}
.add-btn:hover{background:#0284c7}
#robot-list{display:flex;flex-wrap:wrap;gap:1.25rem;padding:1.5rem}
.card{background:#1e293b;border:1px solid #334155;border-radius:.75rem;width:320px;flex-shrink:0;overflow:hidden}
.card-header{display:flex;align-items:center;padding:.75rem 1rem;background:#0f172a;border-bottom:1px solid #334155}
.card-header h2{font-size:1rem;font-weight:600;color:#f1f5f9;flex:1}
.del-btn{background:#ef4444;border:none;color:#fff;padding:.25rem .6rem;border-radius:.375rem;cursor:pointer;font-size:.8rem;transition:background .2s}
.del-btn:hover{background:#b91c1c}
.card-body{padding:.875rem 1rem;font-size:.8rem}
.row{display:flex;justify-content:space-between;padding:.25rem 0;border-bottom:1px solid #1e293b}
.row .label{color:#94a3b8}
.row .val{color:#f1f5f9;font-weight:500;text-align:right}
.badge{display:inline-block;padding:.1rem .45rem;border-radius:.3rem;font-size:.7rem;font-weight:700}
.badge-run{background:#16a34a;color:#fff}
.badge-idle{background:#475569;color:#fff}
.badge-done{background:#334155;color:#94a3b8}
.bat-bar{height:8px;border-radius:4px;background:#1e293b;margin:.3rem 0}
.bat-fill{height:100%;border-radius:4px;transition:width .5s}
.bat-high .bat-fill{background:#22c55e}
.bat-mid .bat-fill{background:#eab308}
.bat-low .bat-fill{background:#ef4444}
.section{margin-top:.75rem;padding-top:.5rem;border-top:1px solid #334155}
.section-title{font-size:.7rem;color:#64748b;font-weight:700;text-transform:uppercase;letter-spacing:.05em;margin-bottom:.4rem}
.ctrl-row{display:flex;gap:.4rem;flex-wrap:wrap;margin-bottom:.4rem;align-items:center}
.ctrl-row label{color:#94a3b8;font-size:.75rem;white-space:nowrap}
input[type=number],input[type=text]{background:#0f172a;border:1px solid #334155;color:#f1f5f9;padding:.25rem .4rem;border-radius:.3rem;width:70px;font-size:.78rem}
.btn{border:none;padding:.3rem .7rem;border-radius:.375rem;cursor:pointer;font-size:.78rem;font-weight:600;transition:background .2s}
.btn-blue{background:#0ea5e9;color:#fff}.btn-blue:hover{background:#0284c7}
.btn-green{background:#16a34a;color:#fff}.btn-green:hover{background:#15803d}
.btn-yellow{background:#ca8a04;color:#fff}.btn-yellow:hover{background:#a16207}
.btn-red{background:#ef4444;color:#fff}.btn-red:hover{background:#b91c1c}
.btn-gray{background:#475569;color:#fff}.btn-gray:hover{background:#334155}
.btn-orange{background:#f97316;color:#fff}.btn-orange:hover{background:#ea580c}
.events{margin-top:.5rem;font-size:.72rem;color:#64748b;line-height:1.6}
.events span{color:#38bdf8}
.emergency{background:#7f1d1d;border:1px solid #ef4444;color:#fca5a5;padding:.3rem .6rem;border-radius:.375rem;margin-bottom:.4rem;font-size:.78rem;font-weight:600}
.card-controls{padding:.875rem 1rem;border-top:2px solid #334155;font-size:.8rem}
.card-status{padding:.875rem 1rem;font-size:.8rem}
</style>
</head>
<body>
<header>
  <h1>🤖 <span>Slamtec</span> Mock Manager</h1>
  <button class="add-btn"
    hx-post="/ui/robots"
    hx-target="#robot-list"
    hx-swap="beforeend">+ Add Robot</button>
</header>
<div id="robot-list">
{{range .Robots}}{{template "card" .}}{{end}}
</div>
</body>
</html>
`))

var cardTmpl = template.Must(pageTmpl.New("card").Parse(`<div class="card" id="card-{{.ID}}">
  <div class="card-header">
    <h2>{{.Label}}</h2>
    <button class="del-btn"
      hx-delete="/ui/robots/{{.ID}}"
      hx-target="#card-{{.ID}}"
      hx-swap="outerHTML"
      hx-confirm="Delete {{.Label}}?">✕</button>
  </div>
  <div id="card-status-{{.ID}}" class="card-status"
    hx-get="/ui/robots/{{.ID}}/status"
    hx-trigger="every 3s"
    hx-swap="innerHTML">
    {{template "card-status" .}}
  </div>
  {{template "card-controls" .}}
</div>
`))

var cardStatusTmpl = template.Must(pageTmpl.New("card-status").Parse(`{{if .HasEmergency}}<div class="emergency">⚠ Emergency Stop Active</div>{{end}}
<div class="row"><span class="label">Pose</span><span class="val">({{printf "%.1f" .PoseX}}, {{printf "%.1f" .PoseY}}) {{printf "%.2f" .PoseYaw}}°</span></div>
<div class="row"><span class="label">Battery</span><span class="val {{.BatteryClass}}">{{.BatteryPct}}% {{if .IsCharging}}⚡{{end}}</span></div>
<div class="{{.BatteryClass}} bat-bar"><div class="bat-fill" style="width:{{.BatteryPct}}%"></div></div>
<div class="row"><span class="label">Dock</span><span class="val">{{.DockingStatus}}</span></div>
<div class="row"><span class="label">Action</span><span class="val">
  {{if eq .ActionStatus "RUNNING"}}<span class="badge badge-run">▶ RUNNING</span>{{else if eq .ActionStatus "IDLE"}}<span class="badge badge-idle">IDLE</span>{{else}}<span class="badge badge-done">DONE</span>{{end}}
</span></div>
{{if .ActionName}}<div class="row"><span class="label">Task</span><span class="val">{{.ActionName}}</span></div>{{end}}
{{if .ActionStage}}<div class="row"><span class="label">Stage</span><span class="val">{{.ActionStage}}</span></div>{{end}}
<div class="row"><span class="label">Delivery</span><span class="val">{{.DeliveryStage}}</span></div>
<div class="row"><span class="label">Floor</span><span class="val">{{.Floor}}</span></div>
<div class="row"><span class="label">Jack</span><span class="val">stage={{.JackStage}} pos={{.JackPos}}</span></div>
<div class="row"><span class="label">Brake</span><span class="val">{{if .SoftBrake}}ON🔴{{else}}off{{end}}</span></div>
{{if .Events}}
<div class="section">
  <div class="section-title">Recent Events</div>
  <div class="events">{{range .Events}}<div><span>{{.Type}}</span> — {{.Message}}</div>{{end}}</div>
</div>
{{end}}
`))

var cardControlsTmpl = template.Must(pageTmpl.New("card-controls").Parse(`<div class="card-controls">
<div class="section">
  <div class="section-title">Move To</div>
  <form hx-post="/ui/robots/{{.ID}}/action" hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">
    <input type="hidden" name="action" value="move">
    <div class="ctrl-row">
      <label>x</label><input type="number" name="x" value="0" step="0.5">
      <label>y</label><input type="number" name="y" value="0" step="0.5">
      <label>yaw</label><input type="number" name="yaw" value="0" step="0.1">
    </div>
    <div class="ctrl-row">
      <button class="btn btn-blue" type="submit">Move</button>
      <button class="btn btn-green" type="button"
        hx-post="/ui/robots/{{.ID}}/action"
        hx-vals='{"action":"home"}'
        hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Go Home</button>
      <button class="btn btn-red" type="button"
        hx-post="/ui/robots/{{.ID}}/action"
        hx-vals='{"action":"abort"}'
        hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Abort</button>
    </div>
  </form>
  <form hx-post="/ui/robots/{{.ID}}/action" hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">
    <input type="hidden" name="action" value="rotate">
    <div class="ctrl-row">
      <label>angle (rad)</label><input type="number" name="angle" value="1.57" step="0.1">
      <button class="btn btn-yellow" type="submit">Rotate</button>
    </div>
  </form>
</div>

<div class="section">
  <div class="section-title">Jack</div>
  <div class="ctrl-row">
    <button class="btn btn-blue"
      hx-post="/ui/robots/{{.ID}}/jack"
      hx-vals='{"cmd":"Up"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Up</button>
    <button class="btn btn-gray"
      hx-post="/ui/robots/{{.ID}}/jack"
      hx-vals='{"cmd":"Down"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Down</button>
    <button class="btn btn-yellow"
      hx-post="/ui/robots/{{.ID}}/jack"
      hx-vals='{"cmd":"Stop"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Stop</button>
    <button class="btn btn-red"
      hx-post="/ui/robots/{{.ID}}/jack"
      hx-vals='{"cmd":"ClearAlarm"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Clear Alarm</button>
  </div>
</div>

<div class="section">
  <div class="section-title">Emergency Brake</div>
  <div class="ctrl-row">
    <button class="btn btn-red"
      hx-post="/ui/robots/{{.ID}}/brake"
      hx-vals='{"value":"on"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Enable Brake</button>
    <button class="btn btn-green"
      hx-post="/ui/robots/{{.ID}}/brake"
      hx-vals='{"value":"off"}'
      hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Release Brake</button>
  </div>
</div>

<div class="section">
  <div class="section-title">Sensors</div>
  <div class="ctrl-row">
    <label>Lidar</label>
    <button class="btn btn-green" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"lidar","val":"on"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">On</button>
    <button class="btn btn-red" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"lidar","val":"off"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Off</button>
    <label>Front</label>
    <button class="btn btn-green" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"front_cam","val":"on"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">On</button>
    <button class="btn btn-red" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"front_cam","val":"off"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Off</button>
    <label>Back</label>
    <button class="btn btn-green" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"back_cam","val":"on"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">On</button>
    <button class="btn btn-red" hx-post="/ui/robots/{{.ID}}/sensor" hx-vals='{"sensor":"back_cam","val":"off"}' hx-target="#card-status-{{.ID}}" hx-swap="innerHTML">Off</button>
  </div>
</div>
</div>
`))

// ---------------------------------------------------------------------------
// UI Handlers
// ---------------------------------------------------------------------------

func (m *RobotManager) handleDashboard(w http.ResponseWriter, req *http.Request) {
	robots := m.listRobots()
	cards := make([]RobotCardData, 0, len(robots))
	for _, rb := range robots {
		cards = append(cards, buildCardData(rb))
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := pageTmpl.Execute(w, dashboardData{Robots: cards}); err != nil {
		log.Printf("ui: dashboard template error: %v", err)
	}
}

func (m *RobotManager) handleGetRobotListUI(w http.ResponseWriter, req *http.Request) {
	robots := m.listRobots()
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, rb := range robots {
		if err := cardTmpl.ExecuteTemplate(w, "card", buildCardData(rb)); err != nil {
			log.Printf("ui: card template error: %v", err)
		}
	}
}

func (m *RobotManager) handleCreateRobotUI(w http.ResponseWriter, req *http.Request) {
	rb := m.createRobot(defaultRobotState())
	data := buildCardData(rb)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := cardTmpl.ExecuteTemplate(w, "card", data); err != nil {
		log.Printf("ui: new card template error: %v", err)
	}
}

func (m *RobotManager) handleDeleteRobotUI(w http.ResponseWriter, req *http.Request) {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil || !m.deleteRobot(id) {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	// Empty response — HTMX outerHTML swap removes the card
	w.WriteHeader(http.StatusOK)
}

func (m *RobotManager) handleGetRobotStatusUI(w http.ResponseWriter, req *http.Request) {
	rb := m.uiRobot(w, req)
	if rb == nil {
		return
	}
	data := buildCardData(rb)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := cardStatusTmpl.ExecuteTemplate(w, "card-status", data); err != nil {
		log.Printf("ui: status template error: %v", err)
	}
}

func (m *RobotManager) handleUIAction(w http.ResponseWriter, req *http.Request) {
	rb := m.uiRobot(w, req)
	if rb == nil {
		return
	}
	req.ParseForm()
	action := req.FormValue("action")

	rb.state.mu.Lock()
	switch action {
	case "move":
		x, _ := strconv.ParseFloat(req.FormValue("x"), 32)
		y, _ := strconv.ParseFloat(req.FormValue("y"), 32)
		yaw, _ := strconv.ParseFloat(req.FormValue("yaw"), 32)
		body, _ := json.Marshal(map[string]any{
			"action_name": "slamtec.agent.actions.MoveToAction",
			"options": map[string]any{
				"target": map[string]any{"x": x, "y": y, "yaw": yaw},
			},
		})
		rb.createActionFromBody(body)
	case "home":
		body, _ := json.Marshal(map[string]any{
			"action_name": "slamtec.agent.actions.GoHomeAction",
			"options":     map[string]any{},
		})
		rb.createActionFromBody(body)
	case "abort":
		if rb.state.CurrentAction != nil && rb.state.CurrentAction.State.Status == StatusRunning {
			rb.state.CurrentAction.State.Status = StatusDone
			rb.state.CurrentAction.State.Result = ResultAborted
			rb.state.CurrentAction.Stage = "ABORTED"
			rb.state.MovementTarget = nil
			rb.addEvent("navigation.aborted", "Action aborted via UI")
		}
	case "rotate":
		angle, _ := strconv.ParseFloat(req.FormValue("angle"), 32)
		body, _ := json.Marshal(map[string]any{
			"action_name": "slamtec.agent.actions.RotateAction",
			"options":     map[string]any{"angle": angle},
		})
		rb.createActionFromBody(body)
	}
	rb.state.mu.Unlock()

	m.renderCardStatus(w, rb)
}

func (m *RobotManager) handleUIJack(w http.ResponseWriter, req *http.Request) {
	rb := m.uiRobot(w, req)
	if rb == nil {
		return
	}
	req.ParseForm()
	cmd := req.FormValue("cmd")
	switch cmd {
	case "Up", "Down", "Stop", "ClearAlarm":
		rb.state.mu.Lock()
		rb.state.JackCommand = cmd
		rb.state.mu.Unlock()
	}
	m.renderCardStatus(w, rb)
}

func (m *RobotManager) handleUIBrake(w http.ResponseWriter, req *http.Request) {
	rb := m.uiRobot(w, req)
	if rb == nil {
		return
	}
	req.ParseForm()
	val := req.FormValue("value")

	rb.state.mu.Lock()
	if val == "on" {
		rb.state.SoftBrakeActive = true
		rb.state.Health.HasSystemEmergencyStop = true
		rb.state.Health.HasError = true
		if rb.state.CurrentAction != nil && rb.state.CurrentAction.State.Status == StatusRunning {
			rb.state.CurrentAction.State.Status = StatusDone
			rb.state.CurrentAction.State.Result = ResultAborted
			rb.state.CurrentAction.Stage = "ABORTED"
			rb.state.MovementTarget = nil
		}
		rb.addEvent("system.emergency_stop", "Soft-brake activated via UI")
	} else {
		rb.state.SoftBrakeActive = false
		if rb.batteryAccum > batteryCritical {
			rb.state.Health.HasSystemEmergencyStop = false
			rb.state.Health.HasError = false
		}
		rb.addEvent("system.emergency_stop_released", "Soft-brake released via UI")
	}
	rb.state.mu.Unlock()

	m.renderCardStatus(w, rb)
}

func (m *RobotManager) handleUISensor(w http.ResponseWriter, req *http.Request) {
	rb := m.uiRobot(w, req)
	if rb == nil {
		return
	}
	req.ParseForm()
	sensor := req.FormValue("sensor")
	val := req.FormValue("val") == "on"

	rb.state.mu.Lock()
	switch sensor {
	case "lidar":
		rb.state.LidarOn = val
	case "front_cam":
		rb.state.FrontCamOn = val
	case "back_cam":
		rb.state.BackCamOn = val
	}
	rb.state.mu.Unlock()

	m.renderCardStatus(w, rb)
}

// uiRobot is a helper that looks up the robot from URL param "id" and writes
// a 404 if not found. Returns nil on failure.
func (m *RobotManager) uiRobot(w http.ResponseWriter, req *http.Request) *MockRobot {
	idStr := chi.URLParam(req, "id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return nil
	}
	rb := m.getrobot(id)
	if rb == nil {
		http.Error(w, "robot not found", http.StatusNotFound)
		return nil
	}
	return rb
}

func (m *RobotManager) renderCardStatus(w http.ResponseWriter, rb *MockRobot) {
	data := buildCardData(rb)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := cardStatusTmpl.ExecuteTemplate(w, "card-status", data); err != nil {
		log.Printf("ui: render card-status error: %v", err)
	}
}











