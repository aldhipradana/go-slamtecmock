package main

import (
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type dashboardViewModel struct {
	GeneratedAt         string
	Pose                Pose
	CurrentFloor        Floor
	CurrentAction       ActionInfo
	HasCurrentAction    bool
	MovementTarget      *MovementTarget
	Battery             PowerStatus
	Health              RobotHealth
	DeliveryStage       string
	SoftBrakeActive     bool
	PhysicalEStopActive bool
	CliffSafe           bool
	LidarOn             bool
	FrontCamOn          bool
	BackCamOn           bool
	FrontVisibleQR      string
	BackVisibleQR       string
	Events              []RobotEvent
	Pois                []MultiFloorPoi
	PoiSearch           string
	PoiTotalCount       int
}

var dashboardTemplates = template.Must(template.New("dashboard").Funcs(template.FuncMap{
	"boolLabel": func(v bool) string {
		if v {
			return "YES"
		}
		return "NO"
	},
	"stateClass": func(v bool) string {
		if v {
			return "ok"
		}
		return "danger"
	},
	"qrLabel": func(v string) string {
		if strings.TrimSpace(v) == "" {
			return "—"
		}
		return v
	},
	"actionStatus": func(status int) string {
		switch status {
		case StatusRunning:
			return "RUNNING"
		case StatusDone:
			return "DONE"
		case StatusPaused:
			return "PAUSED"
		case StatusNew:
			return "NEW"
		default:
			return "UNKNOWN"
		}
	},
	"actionResult": func(result int) string {
		switch result {
		case ResultSuccess:
			return "SUCCESS"
		case ResultAborted:
			return "ABORTED"
		case ResultFailed:
			return "FAILED"
		default:
			return "UNKNOWN"
		}
	},
}).Parse(`
{{define "page"}}
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>go-slamtecmock Dashboard</title>
  <script src="https://unpkg.com/htmx.org@1.9.12"></script>
  <style>
    :root {
      color-scheme: dark;
      --bg: #0f172a;
      --card: #111827;
      --muted: #94a3b8;
      --text: #e5e7eb;
      --ok: #16a34a;
      --warn: #d97706;
      --danger: #dc2626;
      --line: #1f2937;
      --accent: #2563eb;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      font-family: Inter, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background: var(--bg);
      color: var(--text);
    }
    main {
      max-width: 1200px;
      margin: 0 auto;
      padding: 24px;
    }
    .hero {
      display: flex;
      justify-content: space-between;
      gap: 16px;
      align-items: flex-start;
      margin-bottom: 20px;
      flex-wrap: wrap;
    }
    .hero h1 { margin: 0 0 8px; }
    .hero p { margin: 0; color: var(--muted); }
    .links {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
    }
    .links a {
      color: #bfdbfe;
      text-decoration: none;
      border: 1px solid var(--line);
      padding: 8px 12px;
      border-radius: 10px;
      background: rgba(37, 99, 235, 0.1);
    }
    .grid {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: 20px;
    }
    .card {
      background: var(--card);
      border: 1px solid var(--line);
      border-radius: 16px;
      padding: 18px;
      box-shadow: 0 12px 24px rgba(0,0,0,0.18);
    }
    .section-title {
      margin: 0 0 14px;
      font-size: 1rem;
      color: #cbd5e1;
      letter-spacing: .02em;
      text-transform: uppercase;
    }
    .stats {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
      margin-bottom: 18px;
    }
    .stat {
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 12px;
      background: rgba(255,255,255,0.02);
    }
    .label {
      font-size: .76rem;
      text-transform: uppercase;
      color: var(--muted);
      margin-bottom: 6px;
      letter-spacing: .04em;
    }
    .value {
      font-weight: 700;
      font-size: 1.05rem;
      word-break: break-word;
    }
    .ok { color: #86efac; }
    .danger { color: #fca5a5; }
    .warn { color: #fdba74; }
    .controls {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
    }
    .control-group {
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 14px;
    }
    .control-group h3 {
      margin: 0 0 10px;
      font-size: .95rem;
    }
    .buttons {
      display: flex;
      flex-wrap: wrap;
      gap: 8px;
    }
    button {
      border: none;
      border-radius: 10px;
      padding: 10px 12px;
      font-weight: 700;
      cursor: pointer;
      color: white;
    }
    button.safe { background: var(--ok); }
    button.warn { background: var(--warn); }
    button.danger { background: var(--danger); }
    .events {
      display: flex;
      flex-direction: column;
      gap: 10px;
      max-height: 760px;
      overflow: auto;
    }
    .event {
      border: 1px solid var(--line);
      border-radius: 12px;
      padding: 12px;
      background: rgba(255,255,255,0.02);
    }
    .event-type {
      font-weight: 700;
      color: #bfdbfe;
      margin-bottom: 6px;
    }
    .event-time {
      color: var(--muted);
      font-size: .8rem;
      margin-top: 8px;
    }
    .poi-panel {
      margin-top: 20px;
    }
    .poi-toolbar {
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 12px;
      flex-wrap: wrap;
    }
    .poi-search {
      width: min(460px, 100%);
      border: 1px solid var(--line);
      border-radius: 10px;
      padding: 10px 12px;
      background: rgba(255,255,255,0.04);
      color: var(--text);
      font: inherit;
    }
    .poi-search:focus {
      outline: 2px solid rgba(37, 99, 235, 0.45);
      outline-offset: 2px;
    }
    .poi-count {
      color: var(--muted);
      font-size: .88rem;
    }
    .poi-table-wrap {
      overflow: auto;
      max-height: 520px;
      border: 1px solid var(--line);
      border-radius: 12px;
    }
    table {
      width: 100%;
      border-collapse: collapse;
      min-width: 900px;
    }
    th, td {
      padding: 10px 12px;
      text-align: left;
      border-bottom: 1px solid var(--line);
      vertical-align: top;
    }
    th {
      position: sticky;
      top: 0;
      z-index: 1;
      background: #111827;
      color: #cbd5e1;
      font-size: .76rem;
      letter-spacing: .04em;
      text-transform: uppercase;
    }
    td {
      color: #e5e7eb;
      font-size: .9rem;
    }
    tbody tr:last-child td {
      border-bottom: none;
    }
    .poi-id {
      font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", monospace;
      color: #bfdbfe;
      white-space: nowrap;
    }
    .pill {
      display: inline-flex;
      align-items: center;
      min-height: 24px;
      border: 1px solid var(--line);
      border-radius: 999px;
      padding: 3px 8px;
      color: #cbd5e1;
      background: rgba(255,255,255,0.03);
      white-space: nowrap;
    }
    @media (max-width: 900px) {
      .grid { grid-template-columns: 1fr; }
      .stats, .controls { grid-template-columns: 1fr; }
    }
  </style>
</head>
<body>
  <main>
    <section class="hero">
      <div>
        <h1>go-slamtecmock Dashboard</h1>
        <p>Real-time HTMX monitor for robot state, cliff safety, and physical e-stop simulation.</p>
      </div>
      <div class="links">
        <a href="/api/core/slam/v1/localization/pose" target="_blank" rel="noreferrer">Pose API</a>
        <a href="/api/platform/v1/events" target="_blank" rel="noreferrer">Events API</a>
        <a href="/cliff_safe" target="_blank" rel="noreferrer">Cliff API</a>
      </div>
    </section>

    <section class="grid">
      <div id="summary" hx-get="/ui/partials/summary" hx-trigger="load, every 1s" hx-swap="outerHTML">
        {{template "summary" .}}
      </div>
      <div id="events" hx-get="/ui/partials/events" hx-trigger="load, every 1s" hx-swap="outerHTML">
        {{template "events" .}}
      </div>
    </section>

    <section>
      {{template "pois" .}}
    </section>
  </main>
</body>
</html>
{{end}}

{{define "summary"}}
<div id="summary" class="card">
  <h2 class="section-title">Live robot summary</h2>
  <div class="stats">
    <div class="stat">
      <div class="label">Generated at</div>
      <div class="value">{{.GeneratedAt}}</div>
    </div>
    <div class="stat">
      <div class="label">Current floor</div>
      <div class="value">{{.CurrentFloor.Building}} / {{.CurrentFloor.FloorID}}</div>
    </div>
    <div class="stat">
      <div class="label">Pose</div>
      <div class="value">x={{printf "%.2f" .Pose.X}}, y={{printf "%.2f" .Pose.Y}}, yaw={{printf "%.2f" .Pose.Yaw}}</div>
    </div>
    <div class="stat">
      <div class="label">Battery</div>
      <div class="value">{{.Battery.BatteryPercentage}}% · charging={{boolLabel .Battery.IsCharging}}</div>
    </div>
    <div class="stat">
      <div class="label">Current action</div>
      <div class="value">
        {{if .HasCurrentAction}}
          {{.CurrentAction.ActionName}} · {{actionStatus .CurrentAction.State.Status}} / {{actionResult .CurrentAction.State.Result}}
          <div class="label" style="margin-top:8px; text-transform:none; letter-spacing:0;">Stage: {{.CurrentAction.Stage}}</div>
        {{else}}
          idle
        {{end}}
      </div>
    </div>
    <div class="stat">
      <div class="label">Movement target</div>
      <div class="value">
        {{if .MovementTarget}}
          x={{printf "%.2f" .MovementTarget.X}}, y={{printf "%.2f" .MovementTarget.Y}}, yaw={{printf "%.2f" .MovementTarget.Yaw}}
        {{else}}
          none
        {{end}}
      </div>
    </div>
    <div class="stat">
      <div class="label">Cliff safe</div>
      <div class="value {{stateClass .CliffSafe}}">{{boolLabel .CliffSafe}}</div>
    </div>
    <div class="stat">
      <div class="label">Physical e-stop</div>
      <div class="value {{if .PhysicalEStopActive}}danger{{else}}ok{{end}}">{{boolLabel .PhysicalEStopActive}}</div>
    </div>
    <div class="stat">
      <div class="label">Soft brake</div>
      <div class="value {{if .SoftBrakeActive}}warn{{else}}ok{{end}}">{{boolLabel .SoftBrakeActive}}</div>
    </div>
    <div class="stat">
      <div class="label">Health flags</div>
      <div class="value">error={{boolLabel .Health.HasError}} · warning={{boolLabel .Health.HasWarning}} · emergency={{boolLabel .Health.HasSystemEmergencyStop}}</div>
    </div>
    <div class="stat">
      <div class="label">Sensors</div>
      <div class="value">lidar={{boolLabel .LidarOn}} · front={{boolLabel .FrontCamOn}} · back={{boolLabel .BackCamOn}}</div>
    </div>
    <div class="stat">
      <div class="label">Visible QR</div>
      <div class="value">front={{qrLabel .FrontVisibleQR}} · back={{qrLabel .BackVisibleQR}}</div>
    </div>
  </div>

  <div class="controls">
    <div class="control-group">
      <h3>Cliff sensor</h3>
      <div class="buttons">
        <button class="safe" hx-post="/ui/actions/cliff/safe" hx-target="#summary" hx-swap="outerHTML">Set SAFE</button>
        <button class="danger" hx-post="/ui/actions/cliff/unsafe" hx-target="#summary" hx-swap="outerHTML">Set UNSAFE</button>
      </div>
    </div>
    <div class="control-group">
      <h3>Physical e-stop</h3>
      <div class="buttons">
        <button class="danger" hx-post="/ui/actions/physical-estop/press" hx-target="#summary" hx-swap="outerHTML">Press button</button>
        <button class="safe" hx-post="/ui/actions/physical-estop/release" hx-target="#summary" hx-swap="outerHTML">Release button</button>
      </div>
    </div>
  </div>
</div>
{{end}}

{{define "events"}}
<div id="events" class="card">
  <h2 class="section-title">Recent events</h2>
  <div class="events">
    {{if .Events}}
      {{range .Events}}
        <div class="event">
          <div class="event-type">{{.Type}}</div>
          <div>{{.Message}}</div>
          <div class="event-time">{{.Timestamp}}</div>
        </div>
      {{end}}
    {{else}}
      <div class="event">
        <div class="event-type">No events yet</div>
        <div>Trigger cliff or e-stop actions from the dashboard to simulate robot-side incidents.</div>
      </div>
    {{end}}
  </div>
</div>
{{end}}

{{define "pois"}}
<div id="pois" class="card poi-panel">
  <div class="poi-toolbar">
    <h2 class="section-title" style="margin:0;">Available POIs</h2>
  </div>
  <input
    class="poi-search"
    type="search"
    name="q"
    value="{{.PoiSearch}}"
    placeholder="Search POI id, name, type, building, floor"
    hx-get="/ui/partials/pois"
    hx-trigger="keyup changed delay:180ms, search"
    hx-target="#poi-results"
    hx-swap="outerHTML">
  {{template "poi-results" .}}
</div>
{{end}}

{{define "poi-results"}}
<div id="poi-results">
  <div class="poi-count" style="margin-top:12px;">{{len .Pois}} of {{.PoiTotalCount}} shown</div>
  <div class="poi-table-wrap" style="margin-top:12px;">
    <table>
      <thead>
        <tr>
          <th>ID</th>
          <th>Name</th>
          <th>Type</th>
          <th>Building</th>
          <th>Floor</th>
          <th>Pose</th>
        </tr>
      </thead>
      <tbody>
        {{if .Pois}}
          {{range .Pois}}
            <tr>
              <td class="poi-id">{{.ID}}</td>
              <td>{{.PoiName}}</td>
              <td><span class="pill">{{.Type}}</span></td>
              <td>{{.Building}}</td>
              <td>{{.Floor}}</td>
              <td>x={{printf "%.2f" .Pose.X}}, y={{printf "%.2f" .Pose.Y}}, yaw={{printf "%.2f" .Pose.Yaw}}</td>
            </tr>
          {{end}}
        {{else}}
          <tr>
            <td colspan="6">No POIs match the current search.</td>
          </tr>
        {{end}}
      </tbody>
    </table>
  </div>
</div>
{{end}}
`))

func (r *MockRobot) handleUIRedirect(w http.ResponseWriter, req *http.Request) {
	http.Redirect(w, req, "/ui", http.StatusFound)
}

func (r *MockRobot) handleUIDashboard(w http.ResponseWriter, req *http.Request) {
	r.renderDashboardTemplate(w, "page")
}

func (r *MockRobot) handleUISummary(w http.ResponseWriter, req *http.Request) {
	r.renderDashboardTemplate(w, "summary")
}

func (r *MockRobot) handleUIEvents(w http.ResponseWriter, req *http.Request) {
	r.renderDashboardTemplate(w, "events")
}

func (r *MockRobot) handleUIPois(w http.ResponseWriter, req *http.Request) {
	r.renderDashboardTemplateWithData(w, "poi-results", r.dashboardViewModel(req.URL.Query().Get("q")))
}

func (r *MockRobot) handleUISetCliff(w http.ResponseWriter, req *http.Request) {
	mode := strings.ToLower(chi.URLParam(req, "mode"))
	r.state.mu.Lock()
	switch mode {
	case "safe":
		r.setCliffSafeLocked(true, "dashboard UI")
	case "unsafe":
		r.setCliffSafeLocked(false, "dashboard UI")
	default:
		r.state.mu.Unlock()
		http.Error(w, "unknown cliff mode", http.StatusBadRequest)
		return
	}
	r.state.mu.Unlock()
	r.handleUISummary(w, req)
}

func (r *MockRobot) handleUISetPhysicalEStop(w http.ResponseWriter, req *http.Request) {
	mode := strings.ToLower(chi.URLParam(req, "mode"))
	r.state.mu.Lock()
	switch mode {
	case "press", "on", "activate":
		r.setPhysicalEStopLocked(true, "dashboard UI")
	case "release", "off", "deactivate":
		r.setPhysicalEStopLocked(false, "dashboard UI")
	default:
		r.state.mu.Unlock()
		http.Error(w, "unknown physical e-stop mode", http.StatusBadRequest)
		return
	}
	r.state.mu.Unlock()
	r.handleUISummary(w, req)
}

func (r *MockRobot) renderDashboardTemplate(w http.ResponseWriter, name string) {
	r.renderDashboardTemplateWithData(w, name, r.dashboardViewModel(""))
}

func (r *MockRobot) renderDashboardTemplateWithData(w http.ResponseWriter, name string, data dashboardViewModel) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := dashboardTemplates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *MockRobot) dashboardViewModel(poiSearch string) dashboardViewModel {
	r.state.mu.RLock()
	defer r.state.mu.RUnlock()

	poiSearch = strings.TrimSpace(poiSearch)
	view := dashboardViewModel{
		GeneratedAt:         time.Now().Format(time.RFC3339),
		Pose:                r.state.Pose,
		CurrentFloor:        r.state.CurrentFloor,
		Battery:             r.state.Battery,
		Health:              r.state.Health,
		DeliveryStage:       r.state.DeliveryStage,
		SoftBrakeActive:     r.state.SoftBrakeActive,
		PhysicalEStopActive: r.state.PhysicalEStopActive,
		CliffSafe:           r.state.CliffSafe,
		LidarOn:             r.state.LidarOn,
		FrontCamOn:          r.state.FrontCamOn,
		BackCamOn:           r.state.BackCamOn,
		FrontVisibleQR:      r.currentVisibleFrontQrLocked(),
		BackVisibleQR:       r.currentVisibleBackQrLocked(),
		Events:              append([]RobotEvent(nil), r.state.Events...),
		Pois:                filterPois(r.state.MultiFloorPois, poiSearch),
		PoiSearch:           poiSearch,
		PoiTotalCount:       len(r.state.MultiFloorPois),
	}

	if r.state.CurrentAction != nil {
		view.CurrentAction = *r.state.CurrentAction
		view.HasCurrentAction = true
	} else {
		view.CurrentAction = ActionInfo{
			ActionName: "idle",
			Stage:      "IDLE",
			State:      ActionState{Status: StatusDone, Result: ResultSuccess},
		}
	}

	if r.state.MovementTarget != nil {
		target := *r.state.MovementTarget
		view.MovementTarget = &target
	}

	return view
}

func filterPois(pois []MultiFloorPoi, search string) []MultiFloorPoi {
	if strings.TrimSpace(search) == "" {
		return append([]MultiFloorPoi(nil), pois...)
	}

	needle := strings.ToLower(search)
	filtered := make([]MultiFloorPoi, 0, len(pois))
	for _, poi := range pois {
		haystack := strings.ToLower(strings.Join([]string{
			poi.ID,
			poi.PoiName,
			poi.DisplayName,
			poi.Type,
			poi.Building,
			poi.Floor,
			poi.Group,
		}, " "))
		if strings.Contains(haystack, needle) {
			filtered = append(filtered, poi)
		}
	}
	return filtered
}
