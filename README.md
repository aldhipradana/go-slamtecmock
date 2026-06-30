# go-slamtecmock

A Go implementation of the Slamtec Slamware REST API mock server. It starts a local HTTP server on port **1448** that faithfully mimics a physical Slamtec robot, letting you develop and test the full agent stack **without any hardware**.

> This is a direct port of the Java `slamtecmock` module with identical behaviour and API surface.

---

## Table of Contents

1. [Prerequisites](#1-prerequisites)
2. [Running the Mock Server](#2-running-the-mock-server)
3. [Connecting a Client to the Mock](#3-connecting-a-client-to-the-mock)
4. [API Endpoints](#4-api-endpoints)
5. [Default Mock State](#5-default-mock-state)
6. [Simulation Behaviour](#6-simulation-behaviour)
7. [Action Lifecycle](#7-action-lifecycle)
8. [MoveToTag Two-Phase Docking](#8-movetotag-two-phase-docking)
9. [Jack (Lift) Simulation](#9-jack-lift-simulation)
10. [Pre-Seeded POIs](#10-pre-seeded-pois)
11. [Project Structure](#11-project-structure)
12. [Troubleshooting](#12-troubleshooting)

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
go run . -port 8080
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
2026/05/07 10:00:00.000000 Starting go-slamtecmock on port 1448
2026/05/07 10:00:00.000000 MockRobot starting on port 1448 — initial pose (1.00, 2.00), battery 85%
```

Stop with **Ctrl+C**.

---

## 3. Connecting a Client to the Mock

Point `SlamtecRobot` (or any HTTP client) at the mock:

```java
// Java client
SlamtecRobot robot = new SlamtecRobot("http://localhost", "1448");
```

```bash
# Quick curl check
curl http://localhost:1448/api/core/system/v1/power/status
```

The mock passes the initial reachability check (`GET /api/core/system/v1/power/status`) and all subsequent polling calls out of the box.

---

## 4. API Endpoints

### 4.1 System Resources

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

### 4.2 SLAM / Localisation

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/slam/v1/localization/pose` | Current robot pose `{x, y, yaw, pitch, roll, z}` |
| `GET` | `/api/core/slam/v1/homepose` | Registered home/charging dock pose |
| `PUT` | `/api/core/slam/v1/homepose` | Register a pose as the charging dock |
| `PUT` | `/api/multi-floor/localization/v1/pose` | Forcibly relocate the robot |
| `GET` | `/api/core/slam/v1/maps/explore` | Binary map header (32-byte little-endian) |
| `GET` | `/api/core/slam/v1/maps/stcm` | Binary composite map (same as explore in mock) |
| `PUT` | `/api/core/slam/v1/mapping/{enable}` | Enable / disable mapping mode |

### 4.3 Artifact POIs

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/core/artifact/v1/pois` | List POIs (optional `?type=` filter) |
| `POST` | `/api/core/artifact/v1/pois` | Create a new artifact POI |
| `DELETE` | `/api/core/artifact/v1/pois/{id}` | Delete a POI by ID |

### 4.4 Motion / Actions

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
| `slamtec.agent.actions.SchedulableMoveToTagAction` | Two-phase dock alignment (see §8) |
| `slamtec.agent.actions.MoveToTagAction` | Alias for SchedulableMoveToTagAction |

### 4.5 Multi-Floor

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/multi-floor/map/v1/floors` | All configured floors |
| `GET` | `/api/multi-floor/map/v1/floors/{floorId}` | Current floor info |
| `PUT` | `/api/multi-floor/map/v1/floors/{floorId}` | Update current floor (building + floor ID) |
| `GET` | `/api/multi-floor/map/v1/pois` | All multi-floor POIs; supports `?building=`, `?floor=`, `?type=`, `?group=` filters |
| `GET` | `/api/multi-floor/map/v1/elevators/{elevatorId}` | Elevator info (returns a default if unknown ID) |
| `POST` | `/api/multi-floor/motion/v1/movetoaction` | Create a multi-floor move action (same as `/api/core/motion/v1/actions`) |

### 4.6 Events

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/platform/v1/events` | Ring buffer of the last 50 events (newest first) |

### 4.7 Delivery

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

### 4.8 External Sensors

These endpoints are served on the same port as the Slamtec API (they are part of the mock sensor service, not the real robot's sensor daemon).

Front-camera QR visibility is pose-based and filtered by the mock's current building/floor. When a move targets a named POI, that floor context is updated automatically on arrival so cross-site flows can detect the expected QR instead of returning an empty value because of stale floor state.

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

---

## 5. Default Mock State

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

## 6. Simulation Behaviour

### Tick rate

The simulation runs at **5 ms per tick** (200 Hz). All state changes (movement, battery, jack) happen inside the tick loop protected by a single `sync.RWMutex`.

### Movement

- Speed: **15 m/s** (fast for simulation; mimics real-time testing without waiting).
- Each tick the robot advances toward `MovementTarget` by `speed × tickSec` metres.
- On arrival the action transitions to `status=DONE, result=SUCCESS`.
- For named-POI navigation (`target.poi_name`), arrival also syncs `CurrentFloor` to the POI's building/floor.

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

## 7. Action Lifecycle

```
POST /api/core/motion/v1/actions
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
curl http://localhost:1448/api/core/motion/v1/actions/current
```

Abort:

```bash
curl -X DELETE http://localhost:1448/api/core/motion/v1/actions/current
```

---

## 8. MoveToTag Two-Phase Docking

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

```json
{
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
}
```

---

## 9. Jack (Lift) Simulation

The jack animates gradually over ~3 seconds for a full stroke.

| Command | Behaviour |
|---|---|
| `"Up"` | Moves `actual_pos` from 31 → 26000001; `stage` transitions 3 → 5 when complete |
| `"Down"` | Moves `actual_pos` from 26000001 → 31; `stage` transitions 3 → 2 when complete |
| `"Stop"` | Halts movement immediately |
| `"ClearAlarm"` | Resets `alarm` to 0 |

```bash
# Raise jack
curl -X POST http://localhost:1448/api/core/system/v1/jack/status \
     -H "Content-Type: application/json" \
     -d '"Up"'

# Poll jack status
curl http://localhost:1448/api/core/system/v1/jack/status
```

---

## 10. Pre-Seeded POIs

The mock starts with a full POI dataset covering three buildings:

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
| `INSTITUTION_GROUND_LIFT_DOOR_QR_POINT` | Ground Lift Door QR Point | QR_POINT |
| `INSTITUTION_GROUND_LIFT_WALL_QR_POINT` | Ground Lift Wall QR Point | QR_POINT |

### Institution (floor 5 – Level 5)

| ID | Display name | Type |
|---|---|---|
| `INSTITUTION_TOP_LOADING_1..4` | Top Loading 1–4 | TROLLEY_POINT |
| `PORT_INSTITUTION_TOP_LOADING_1..4` | Port Top Loading 1–4 | INTERSECTION |
| `INSTITUTION_TOP_LIFT_1..4` | Top Lift 1–4 | LIFT_TROLLEY_POINT |
| `CARGO_LIFT_EXIT_LEVEL5` | Cargo Lift Exit Level 5 | LIFT_WAITING_POINT |
| `INSTITUTION_TOP_LIFT_DOOR_QR_POINT` | Top Lift Door QR Point | QR_POINT |
| `INSTITUTION_TOP_LIFT_WALL_QR_POINT` | Top Lift Wall QR Point | QR_POINT |
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
curl "http://localhost:1448/api/multi-floor/map/v1/pois?building=Kitchen&floor=1"
curl "http://localhost:1448/api/multi-floor/map/v1/pois?type=CHARGER_DOCK_POINT"
```

---

## 11. Project Structure

```
go-slamtecmock/
├── main.go       — Entry point; parses -port flag and blocks
├── robot.go      — MockRobot: simulation tick loop, action helpers, map header builder
├── handlers.go   — All HTTP route handlers wired via chi router
├── state.go      — RobotState struct, all POJOs, default seed data
├── go.mod        — Module: github.com/robotmanager/go-slamtecmock
└── README.md     — This file
```

**Key dependency:** [`github.com/go-chi/chi/v5`](https://github.com/go-chi/chi) — lightweight, idiomatic Go router.

---

## 12. Troubleshooting

### Port already in use

```
listen tcp :1448: bind: address already in use
```

Either kill the process on port 1448 or use a different port:

```bash
go run . -port 9000
```

### Action stuck in RUNNING

Check if a soft-brake is active:

```bash
curl "http://localhost:1448/api/core/system/v1/parameter?param=base.emergency_stop"
# → "on"

# Release it
curl -X PUT http://localhost:1448/api/core/system/v1/parameter \
     -H "Content-Type: application/json" \
     -d '{"param":"base.emergency_stop","value":"off"}'
```

### Battery critically low / emergency stop

When battery drops below 5 %, `has_system_emergency_stop` is set automatically and the robot freezes. Recharge by sending GoHome while battery is still above 5 %:

```bash
curl -X POST http://localhost:1448/api/core/motion/v1/actions \
     -H "Content-Type: application/json" \
     -d '{"action_name":"slamtec.agent.actions.GoHomeAction","options":{}}'
```

### POI not found → action FAILED

Verify the POI ID or name exists:

```bash
curl http://localhost:1448/api/multi-floor/map/v1/pois | jq '.[].id'
```

### Events

All state transitions emit events visible at:

```bash
curl http://localhost:1448/api/platform/v1/events
```

Common event types: `navigation.arrived`, `navigation.aborted`, `charging.docked`, `system.power_low`, `system.critical_low_battery`, `system.emergency_stop`, `docking.complete`, `jack.up`, `jack.down`, `floor.updated`, `delivery.start_pickup`.
