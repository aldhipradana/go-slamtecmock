# go-slamtecmock

A Go multi-robot mock server that faithfully implements the Slamtec Slamware REST API. Run as many virtual robots as you need — each with fully independent state and simulation — on a **single port**, with a built-in **HTMX dashboard** to manage and observe them in real time. No physical hardware required.

> Drop-in replacement for a real Slamtec robot. Point the robot agent's `slamtec.baseUrl` and `sensor.baseUrl` at `/robot/{id}` and everything works as-is.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Running the Mock Server](#2-running-the-mock-server)
3. [Connecting the Robot Agent](#3-connecting-the-robot-agent)
4. [Dashboard (HTMX UI)](#4-dashboard-htmx-ui)
5. [Management REST API](#5-management-rest-api)
6. [Slamtec API Endpoints](#6-slamtec-api-endpoints)
7. [Default Mock State](#7-default-mock-state)
8. [Simulation Behaviour](#8-simulation-behaviour)
9. [Action Lifecycle](#9-action-lifecycle)
10. [MoveToTag Two-Phase Docking](#10-movetotag-two-phase-docking)
11. [Jack (Lift) Simulation](#11-jack-lift-simulation)
12. [Pre-Seeded POIs](#12-pre-seeded-pois)
13. [Project Structure](#13-project-structure)
14. [Troubleshooting](#14-troubleshooting)

---

## 1. Prerequisites

| Requirement | Version |
|---|---|
| Go | 1.22+ |
| Network | Port 1448 must be free (or use `-port`) |

No external services needed — all state is in-memory.

---

## 2. Running the Mock Server

### Default port (1448)

```bash
cd go-slamtecmock   # from the repo root: rm/go-slamtecmock
go run .
```

### Custom port

```bash
go run . -port 8089
```

### Build a binary

```bash
go build -o slamtecmock .
./slamtecmock -port 1448
```

### Docker (optional)

```bash
docker build -t go-slamtecmock .
docker run -p 1448:1448 go-slamtecmock
```

### Expected startup log

```
2026/05/07 10:00:00.000000 go-slamtecmock listening on :1448
2026/05/07 10:00:00.000000 Dashboard → http://localhost:1448
2026/05/07 10:00:00.000000 Robot #1 API → http://localhost:1448/robot/1/api/core/system/v1/power/status
2026/05/07 10:00:00.000000 Compat alias → http://localhost:1448/api/core/system/v1/power/status
2026/05/07 10:00:00.000000 Robot #1 started — pose (1.00, 2.00) battery 85%
2026/05/07 10:00:00.000000 RobotManager: created Robot #1
```

Stop with **Ctrl+C**.

---

## 3. Connecting the Robot Agent

### Using `config.properties`

The robot agent (e.g. `spsdelivery`, `vapp`) reads `config.properties` to locate the Slamtec robot and the sensor service. When using the mock, set `driver=slamtec` and point both URLs at the mock server's robot sub-path:

```properties
# Robot Driver — must be "slamtec" to use this mock
driver=slamtec

# Slamtec robot base URL — point at the mock's per-robot path
slamtec.baseUrl=http://127.0.0.1:1448/robot/1

# External sensor service URL — also served by the mock on the same sub-path
# Endpoints polled: GET /front_cam  /back_cam  /cliff_safe
sensor.baseUrl=http://127.0.0.1:1448/robot/1
sensor.timeout.ms=2000
```

> **Multiple robots:** To run a second robot agent against Robot #2, change both URLs to `.../robot/2`.  
> **Backward-compat:** The legacy `/api/...` prefix (without a robot sub-path) still routes to Robot #1, so existing configs that point at `http://127.0.0.1:1448` with no sub-path continue to work unchanged.

### Full example `config.properties` for local development

```properties
# SPS Delivery Robot Agent Configuration

robotId=70000000-0000-0000-0000-000000000001
layoutId=40000000-0000-0000-0000-000000000000

# Robot Manager Core API
httpBaseUrl=http://127.0.0.1:8080/
baseUrlApiVersion=robotapi/v1/

# MQTT
mqtt.brokerUrl=tcp://localhost:1883
mqtt.clientId=vapp_spsdelivery_robot
mqtt.username=
mqtt.password=
mqtt.prefix=ALDHI/

# HTTP dashboard port (robot agent's own UI)
dashboard.http.port=8090

# ── Mock robot configuration ──────────────────────────────
driver=slamtec
slamtec.baseUrl=http://127.0.0.1:1448/robot/1
sensor.baseUrl=http://127.0.0.1:1448/robot/1
sensor.timeout.ms=2000
```

### Quick curl check

```bash
# Robot #1 via new URL scheme
curl http://localhost:1448/robot/1/api/core/system/v1/power/status

# Backward-compat alias (Robot #1)
curl http://localhost:1448/api/core/system/v1/power/status
```

---

## 4. Dashboard (HTMX UI)

Open **http://localhost:1448** in any browser. The dashboard shows a live card for every robot, auto-refreshing every **3 seconds**.

```
┌──────────────────────────────────────────────────────────────┐
│  🤖 Slamtec Mock Manager                      [+ Add Robot]  │
├──────────────────────────────────────────────────────────────┤
│  ┌──────────────────────────────┐  ┌────────────────────────┐ │
│  │ Robot #1               [✕]  │  │ Robot #2          [✕]  │ │
│  │ Pose: (1.0, 2.0) 0.00°     │  │ Pose: (5.0, 10.0) 0.00°│ │
│  │ Battery: ████████░░ 85%    │  │ Battery: ██████░░░ 60%  │ │
│  │ Charging: No  Dock: No     │  │ Charging: No  Dock: No  │ │
│  │ Action: IDLE               │  │ Action: RUNNING         │ │
│  │ Floor: Kitchen / 1         │  │   MoveToAction          │ │
│  │ Jack: stage=2 pos=31       │  │   Stage: GOING_TO_TARGET│ │
│  │ Brake: off                 │  │ ...                     │ │
│  │ ─── Controls ────────────  │  │                         │ │
│  │ Move: x[  ] y[  ] yaw[  ] │  │                         │ │
│  │ [Move] [Go Home] [Abort]   │  │                         │ │
│  │ Rotate: angle[  ] [Rotate] │  │                         │ │
│  │ Jack: [Up] [Down] [Stop]   │  │                         │ │
│  │ Brake: [Enable Brake]      │  │                         │ │
│  │ Sensors: Lidar[On] …       │  │                         │ │
│  │ ─── Recent Events ───────  │  │                         │ │
│  │ navigation.arrived ...     │  │                         │ │
│  └──────────────────────────────┘  └────────────────────────┘ │
└──────────────────────────────────────────────────────────────┘
```

### Dashboard controls

| Control | What it does |
|---|---|
| **+ Add Robot** | Creates a new robot with default state; appends its card |
| **✕** | Deletes the robot and stops its simulation loop |
| **Move** | Sends a `MoveToAction` to the given x/y/yaw |
| **Go Home** | Sends a `GoHomeAction`; robot navigates to dock and starts charging |
| **Abort** | Cancels the current running action |
| **Rotate** | Sends a `RotateAction` by the given angle in radians |
| **Jack Up/Down/Stop/Clear Alarm** | Controls the jack (lift) actuator |
| **Enable / Release Brake** | Toggles `base.emergency_stop` on/off |
| **Sensor On/Off** | Toggles lidar, front camera, back camera |

---

## 5. Management REST API

These endpoints manage the robot pool itself. The Slamtec API (§6) is separate and per-robot.

### List all robots

```bash
GET /robots
```

```json
[
  { "id": 1, "pose_x": 1.0, "pose_y": 2.0, "battery_pct": 85, "is_charging": false, "action_status": "IDLE" },
  { "id": 2, "pose_x": 5.0, "pose_y": 10.0, "battery_pct": 60, "is_charging": false, "action_status": "RUNNING" }
]
```

### Create a robot

```bash
POST /robots
Content-Type: application/json

# Optional body — omit any field to use its default
{ "x": 5.0, "y": 10.0, "yaw": 0.0, "battery": 60 }
```

```json
{ "id": 3 }
```

### Delete a robot

```bash
DELETE /robots/{id}
# → 204 No Content
```

---

## 6. Slamtec API Endpoints

All Slamtec endpoints are available under **two URL schemes**:

| Scheme | Example | Notes |
|---|---|---|
| **Per-robot** | `/robot/1/api/core/system/v1/power/status` | Preferred; supports multiple robots |
| **Compat alias** | `/api/core/system/v1/power/status` | Always routes to Robot #1; for legacy clients |

The table below lists paths **without the prefix** (`/robot/{id}` or `/api`). Append to either scheme.

### 6.1 System Resources

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/system/v1/power/status` | Battery percentage, charging state, docking status |
| `GET` | `/api/core/system/v1/network/status` | Wi-Fi quality, SSID, IP, signal dB |
| `GET` | `/api/core/system/v1/robot/info` | Model, serial number, firmware version |
| `GET` | `/api/core/system/v1/robot/health` | Error flags (emergency stop, lidar, etc.) |
| `GET` | `/api/core/system/v1/capabilities` | List of enabled capabilities |
| `GET` | `/api/core/system/v1/parameter?param=<name>` | Read a system parameter |
| `PUT` | `/api/core/system/v1/parameter` | Write a system parameter |
| `POST` | `/api/core/system/v1/power/{cmd}` | Restart / shutdown (no-op in mock) |
| `GET` | `/api/core/system/v1/jack/status` | Jack position, stage, alarm, drv_status |
| `POST` | `/api/core/system/v1/jack/status` | Send jack command: `"Up"` / `"Down"` / `"Stop"` / `"ClearAlarm"` |

#### Special parameters

| Parameter | Values | Effect |
|---|---|---|
| `base.emergency_stop` | `"on"` / `"off"` | Activates / releases the soft-brake. Aborts the running action when set to `"on"`. |
| `base.brake_release` | `"on"` / `"off"` | Emits a brake-release event (no movement effect in mock). |

### 6.2 SLAM / Localisation

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/slam/v1/localization/pose` | Current robot pose `{x, y, yaw, pitch, roll, z}` |
| `GET` | `/api/core/slam/v1/homepose` | Registered home/charging dock pose |
| `PUT` | `/api/core/slam/v1/homepose` | Register a pose as the charging dock |
| `PUT` | `/api/multi-floor/localization/v1/pose` | Forcibly relocate the robot |
| `GET` | `/api/core/slam/v1/maps/explore` | Binary map header (32-byte little-endian) |
| `GET` | `/api/core/slam/v1/maps/stcm` | Binary composite map (same as explore in mock) |
| `PUT` | `/api/core/slam/v1/mapping/{enable}` | Enable / disable mapping mode |

### 6.3 Artifact POIs

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/artifact/v1/pois` | List POIs (optional `?type=` filter) |
| `POST` | `/api/core/artifact/v1/pois` | Create a new artifact POI |
| `DELETE` | `/api/core/artifact/v1/pois/{id}` | Delete a POI by ID |

### 6.4 Motion / Actions

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/motion/v1/action-factories` | Supported action type names |
| `GET` | `/api/core/motion/v1/strategies` | Available motion strategies |
| `GET` | `/api/core/motion/v1/strategies/{id}` | Current motion strategy |
| `PUT` | `/api/core/motion/v1/strategies/{id}` | Set current motion strategy |
| `POST` | `/api/core/motion/v1/actions` | **Create an action** (main dispatch endpoint) |
| `POST` | `/api/core/slam/v1/action/create` | Create action (legacy alias) |
| `GET` | `/api/core/motion/v1/actions/{actionId}` | Query action status; use `current` for the running action |
| `DELETE` | `/api/core/motion/v1/actions/{actionId}` | Abort the running action |

#### Supported action names

| Action name | Behaviour |
|---|---|
| `slamtec.agent.actions.MoveToAction` | Navigate to `{x, y}` target |
| `slamtec.agent.actions.MultiFloorMoveAction` | Navigate to target or named POI |
| `slamtec.agent.actions.GoHomeAction` | Navigate to the registered home/dock pose, dock and charge on arrival |
| `slamtec.agent.actions.MultiFloorBackHomeAction` | Alias for GoHome |
| `slamtec.agent.actions.RotateAction` | Rotate by `angle` radians (instant in mock) |
| `slamtec.agent.actions.EnterElevatorAction` | Simulate entering elevator (completes after 1.5 s) |
| `slamtec.agent.actions.LeaveElevatorAction` | Simulate leaving elevator (completes after 1.5 s) |
| `slamtec.agent.actions.SchedulableMoveToTagAction` | Two-phase dock alignment (see §10) |
| `slamtec.agent.actions.MoveToTagAction` | Alias for SchedulableMoveToTagAction |

### 6.5 Multi-Floor

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/multi-floor/map/v1/floors` | All configured floors |
| `GET` | `/api/multi-floor/map/v1/floors/{floorId}` | Current floor info |
| `PUT` | `/api/multi-floor/map/v1/floors/{floorId}` | Update current floor (building + floor ID) |
| `GET` | `/api/multi-floor/map/v1/pois` | All multi-floor POIs; supports `?building=`, `?floor=`, `?type=`, `?group=` filters |
| `GET` | `/api/multi-floor/map/v1/elevators/{elevatorId}` | Elevator info (returns a default if unknown ID) |
| `POST` | `/api/multi-floor/motion/v1/movetoaction` | Create a multi-floor move action (same as `/api/core/motion/v1/actions`) |

### 6.6 Events

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/platform/v1/events` | Ring buffer of the last 50 events (newest first) |

### 6.7 Delivery

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/delivery/v1/stage` | Current delivery stage string |
| `GET` | `/api/delivery/v1/settings` | Delivery settings map |
| `PUT` | `/api/delivery/v1/settings/timeout` | Update `food_pickup_timeout` |
| `PUT` | `/api/delivery/v1/tasks/task_execution` | Enable / disable task execution |
| `PUT` | `/api/delivery/v1/tasks/start_pickup` | Advance stage to `ARRIVED_AT_TARGET` |
| `PUT` | `/api/delivery/v1/tasks/end_pickup` | Advance stage to `ON_RETURNING` |
| `GET` | `/api/delivery/v1/cargos` | List cargos |
| `PUT` | `/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}` | Box door/lock operations: `open`, `close`, `lock`, `unlock` |
| `GET` | `/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/operation_result` | Last cargo operation result |

### 6.8 External Sensors

The mock serves the same sensor endpoints that the real robot's sensor daemon exposes. The robot agent reads these via `sensor.baseUrl` in `config.properties`.

| Method | Path | Response |
|---|---|---|
| `GET` | `/front_cam` | Plain-text QR tag ID detected by front camera, or empty |
| `GET` | `/back_cam` | Plain-text QR tag ID detected by back camera, or empty |
| `GET` | `/cliff_safe` | Plain-text `true` / `false` |
| `GET` | `/sensor/status` | Plain-text multi-line: `lidar=<bool>\nfront_cam=<bool>\nback_cam=<bool>` |
| `POST` | `/sensor/lidar/on` | Turn lidar on |
| `POST` | `/sensor/lidar/off` | Turn lidar off |
| `POST` | `/sensor/front_cam/on` | Turn front camera on |
| `POST` | `/sensor/front_cam/off` | Turn front camera off |
| `POST` | `/sensor/back_cam/on` | Turn back camera on |
| `POST` | `/sensor/back_cam/off` | Turn back camera off |

> **Note:** `sensor.baseUrl` in `config.properties` should point at the **same** per-robot sub-path as `slamtec.baseUrl` — e.g. `http://127.0.0.1:1448/robot/1`. The mock serves both the Slamtec API and the sensor endpoints under the same prefix.

---

## 7. Default Mock State

Each robot starts with the following state (shared default; customisable via `POST /robots` body):

| Property | Default value |
|---|---|
| Pose | `x=1.0, y=2.0, yaw=0.0` |
| Battery | 85 %, not charging, not on dock |
| Home / dock pose | `x=564, y=406, yaw=0` (KITCHEN_CHARGING_DOCK_POINT) |
| Network | quality=90, SSID="MockWifi", IP="192.168.1.100" |
| Robot model | MockRobot-4000, firmware 4.6.0 |
| Health | All flags `false`, no errors |
| Jack | Fully down (pos=31, stage=2) |
| Delivery stage | `IDLE` |
| Current floor | Building="Kitchen", Floor="1" |

---

## 8. Simulation Behaviour

### Tick rate (adaptive)

Each robot runs its own independent simulation goroutine with an **adaptive tick interval**:

| Robot state | Tick interval | CPU wakeups/s |
|---|---|---|
| Action RUNNING or jack moving | **5 ms** | 200 |
| Idle | **200 ms** | 5 |

The transition from idle → active happens within ≤200 ms of an action being created — imperceptible for mock purposes. This gives ~97% CPU reduction for idle robots.

### Movement

- Speed: **15 m/s** (fast for simulation; no waiting around).
- Each tick the robot advances toward `MovementTarget` by `speed × tickSec` metres.
- On arrival the action transitions to `status=DONE, result=SUCCESS`.

### Battery

| Condition | Rate |
|---|---|
| Moving (action RUNNING) | −0.05 %/s drain |
| Docked & charging | +0.5 %/s charge |
| Below 15 % | `system.power_low` event emitted (max once per 30 s) |
| Below 5 % | `system.critical_low_battery` event; `has_system_emergency_stop = true` |

### CORS

All origins are allowed (`Access-Control-Allow-Origin: *`).

---

## 9. Action Lifecycle

```
POST /robot/{id}/api/core/motion/v1/actions
        │
        ▼
  ActionInfo { status: RUNNING (1), result: SUCCESS (0) }
        │
  [tick loop moves robot toward target]
        │
        ▼
  ActionInfo { status: DONE (4), result: SUCCESS (0) }   ← arrived
         or   { status: DONE (4), result: ABORTED (-2) }  ← DELETE called
         or   { status: DONE (4), result: FAILED (-1) }   ← POI not found
```

**Status codes**

| Constant | Value |
|---|---|
| `STATUS_NEW` | 0 |
| `STATUS_RUNNING` | 1 |
| `STATUS_PAUSED` | 3 |
| `STATUS_DONE` | 4 |

**Result codes**

| Constant | Value |
|---|---|
| `RESULT_SUCCESS` | 0 |
| `RESULT_FAILED` | −1 |
| `RESULT_ABORTED` | −2 |

Query current action:

```bash
curl http://localhost:1448/robot/1/api/core/motion/v1/actions/current
```

Abort:

```bash
curl -X DELETE http://localhost:1448/robot/1/api/core/motion/v1/actions/current
```

---

## 10. MoveToTag Two-Phase Docking

`SchedulableMoveToTagAction` / `MoveToTagAction` simulates precision tag docking in two phases:

```
Phase 1 – APPROACH
  Robot navigates to the landing point (target x/y) at full speed.

Phase 2 – DOCKING
  Robot aligns with the tag:
    • backward_docking=true  → backs toward tag (−cos(yaw), −sin(yaw))
    • backward_docking=false → advances toward tag (+cos(yaw), +sin(yaw))
  Distance: dock_allowance metres (default 0.275 m), at 50% speed.
```

Events emitted:
- `docking.approach_complete` — transition from APPROACH to DOCKING
- `docking.complete` — docking finished successfully

Example request body:

```bash
curl -X POST http://localhost:1448/robot/1/api/core/motion/v1/actions \
  -H "Content-Type: application/json" \
  -d '{
    "action_name": "slamtec.agent.actions.SchedulableMoveToTagAction",
    "options": {
      "target": { "x": 10.0, "y": 5.0, "yaw": 1.57 },
      "move_to_tag_options": {
        "backward_docking": true,
        "dock_allowance": 0.275,
        "tag_type": 3,
        "reflect_tag_num": 2,
        "tag_ids": [1, 2]
      }
    }
  }'
```

---

## 11. Jack (Lift) Simulation

The jack animates gradually over ~3 seconds for a full stroke.

| Command | Behaviour |
|---|---|
| `"Up"` | Moves `actual_pos` from 31 → 26000001; `stage` transitions 3 → 5 when complete |
| `"Down"` | Moves `actual_pos` from 26000001 → 31; `stage` transitions 3 → 2 when complete |
| `"Stop"` | Halts movement immediately |
| `"ClearAlarm"` | Resets `alarm` to 0 |

```bash
# Raise jack on Robot #1
curl -X POST http://localhost:1448/robot/1/api/core/system/v1/jack/status \
     -H "Content-Type: application/json" \
     -d '"Up"'

# Poll jack status
curl http://localhost:1448/robot/1/api/core/system/v1/jack/status
```

---

## 12. Pre-Seeded POIs

Every robot starts with the same full POI dataset:

### Kitchen (floor 1)

| ID | Display name | Type |
|---|---|---|
| `KITCHEN_HOME_1` | Kitchen Home | INTERSECTION |
| `KITCHEN_HOME_2` | Kitchen Home 2 | INTERSECTION |
| `KITCHEN_VEHICLE_PORT_1` | Kitchen Vehicle Port 1 | INTERSECTION |
| `KITCHEN_LOADING_1..4` | Kitchen Loading 1–4 | TROLLEY_POINT |
| `KITCHEN_VEHICLE_1..4` | Kitchen Vehicle 1–4 | AV_LOADING_POINT |
| `KITCHEN_CHARGING_DOCK_POINT` | Kitchen Charging Dock | CHARGER_DOCK_POINT |

### Institution (floor 1 – Ground)

| ID | Display name | Type |
|---|---|---|
| `INSTITUTION_GROUND_HOME` | Ground Home | CHARGER_DOCK_POINT |
| `INSTITUTION_GROUND_LOADING_1..4` | Ground Loading 1–4 | TROLLEY_POINT |
| `PORT_INSTITUTION_GROUND_LOADING_1..4` | Port Ground Loading 1–4 | INTERSECTION |
| `INSTITUTION_GROUND_VEHICLE_1..4` | Ground Vehicle 1–4 | AV_LOADING_POINT |
| `INSTITUTION_GROUND_LIFT_1..4` | Ground Lift 1–4 | LIFT_TROLLEY_POINT |
| `CARGO_LIFT_ENTRY_GROUND` | Cargo Lift Entry Ground | LIFT_WAITING_POINT |

### Institution (floor 5 – Level 5)

| ID | Display name | Type |
|---|---|---|
| `INSTITUTION_TOP_LOADING_1..4` | Top Loading 1–4 | TROLLEY_POINT |
| `PORT_INSTITUTION_TOP_LOADING_1..4` | Port Top Loading 1–4 | INTERSECTION |
| `INSTITUTION_TOP_LIFT_1..4` | Top Lift 1–4 | LIFT_TROLLEY_POINT |
| `CARGO_LIFT_EXIT_LEVEL5` | Cargo Lift Exit Level 5 | LIFT_WAITING_POINT |
| `DESTINATION_AREA` | Destination Area | INTERSECTION |
| `LEVEL5_CHARGER` | Level 5 Robot Charger | CHARGER_DOCK_POINT |

### AVMap (floor 1)

| ID | Display name | Type |
|---|---|---|
| `AV_HOME` | AV Home | CHARGER_DOCK_POINT |
| `KITCHEN_ENTRANCE` | Kitchen Entrance | WAYPOINT |
| `KITCHEN_EXIT` | Kitchen Exit | WAYPOINT |
| `KITCHEN_DOCK` | Kitchen Dock | AV_LOADING_POINT |
| `INSTITUTION_DOCK` | Institution Dock | AV_LOADING_POINT |
| `PRISON_GATE` | Prison Gate | GATE |

Query by building/floor:

```bash
curl "http://localhost:1448/robot/1/api/multi-floor/map/v1/pois?building=Kitchen&floor=1"
curl "http://localhost:1448/robot/1/api/multi-floor/map/v1/pois?type=CHARGER_DOCK_POINT"
```

---

## 13. Project Structure

```
go-slamtecmock/
├── main.go       — Entry point; -port flag; starts RobotManager + HTTP server
├── manager.go    — RobotManager: robot CRUD, withRobot middleware, BuildRouter(),
│                   CORS, per-robot Slamtec sub-routes, /api compat alias,
│                   management REST handlers (GET/POST/DELETE /robots)
├── ui.go         — HTMX dashboard: view model, inline HTML templates, UI handlers
├── robot.go      — MockRobot: adaptive tick loop, movement, battery, jack,
│                   action creation and dispatch logic
├── handlers.go   — All 46 Slamtec HTTP handler methods on *MockRobot
├── state.go      — RobotState struct, all data types, default seed data + POIs
├── go.mod        — Module: github.com/aldhipradana/go-slamtecmock
├── go.sum
├── docs/
│   └── multi-robot-ui-plan.md  — Design plan for the multi-robot refactor
├── SlamtecSwaggerAPI.md        — Slamtec REST API reference
└── README.md     — This file
```

**Dependency:** [`github.com/go-chi/chi/v5`](https://github.com/go-chi/chi) — lightweight Go router. No other external dependencies; HTMX is loaded from CDN in the page template.

---

## 14. Troubleshooting

### Port already in use

```
listen tcp :1448: bind: address already in use
```

Use a different port and update `config.properties` accordingly:

```bash
go run . -port 8089
```

```properties
slamtec.baseUrl=http://127.0.0.1:8089/robot/1
sensor.baseUrl=http://127.0.0.1:8089/robot/1
```

### Robot agent can't connect

Verify the mock is reachable from the agent's host:

```bash
curl http://127.0.0.1:1448/robot/1/api/core/system/v1/power/status
# expected: {"batteryPercentage":85,"isCharging":false,...}
```

Check that `driver=slamtec` is set in `config.properties` — if `driver=dummy` the agent won't make any Slamtec HTTP calls.

### Action stuck in RUNNING

Check if a soft-brake is active:

```bash
curl "http://localhost:1448/robot/1/api/core/system/v1/parameter?param=base.emergency_stop"
# → "on"

# Release it
curl -X PUT http://localhost:1448/robot/1/api/core/system/v1/parameter \
     -H "Content-Type: application/json" \
     -d '{"param":"base.emergency_stop","value":"off"}'
```

Or use the **Release Brake** button on the dashboard.

### Battery critically low / emergency stop

When battery drops below 5 %, `has_system_emergency_stop` is set automatically and the robot freezes. Send GoHome before it gets that low, or use the dashboard to release the brake after acknowledging the condition:

```bash
curl -X POST http://localhost:1448/robot/1/api/core/motion/v1/actions \
     -H "Content-Type: application/json" \
     -d '{"action_name":"slamtec.agent.actions.GoHomeAction","options":{}}'
```

### POI not found → action FAILED

Verify the POI ID or name exists:

```bash
curl http://localhost:1448/robot/1/api/multi-floor/map/v1/pois | jq '.[].id'
```

### Sensor endpoints returning empty

The mock serves `GET /front_cam`, `GET /back_cam`, and `GET /cliff_safe` on the **same sub-path** as the Slamtec API. Confirm `sensor.baseUrl` in `config.properties` matches `slamtec.baseUrl` exactly:

```properties
# Correct — both point to the same robot sub-path
slamtec.baseUrl=http://127.0.0.1:1448/robot/1
sensor.baseUrl=http://127.0.0.1:1448/robot/1
```

### Events log

All state transitions emit events:

```bash
curl http://localhost:1448/robot/1/api/platform/v1/events
```

Common event types: `navigation.arrived`, `navigation.aborted`, `charging.docked`, `system.power_low`, `system.critical_low_battery`, `system.emergency_stop`, `system.emergency_stop_released`, `docking.complete`, `jack.up`, `jack.down`, `floor.updated`, `delivery.start_pickup`.

