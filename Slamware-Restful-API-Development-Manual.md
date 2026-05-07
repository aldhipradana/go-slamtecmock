# Slamware RESTful API Development Manual

## Table of Contents

1. [Overview](#1-overview)
2. [Interface Specifications](#2-interface-specifications)
   - [2.1 Naming Method](#21-naming-method)
   - [2.2 Type of Method](#22-type-of-method)
   - [2.3 API Parameters](#23-api-parameters)
   - [2.4 Return Status Code](#24-return-status-code)
   - [2.5 Return Value](#25-return-value)
   - [2.6 How to View Interfaces in API Documentation](#26-how-to-view-interfaces-in-api-documentation)
3. [Interface Classification](#3-interface-classification)
   - [3.1 System Resources](#31-system-resources)
   - [3.2 Functions Related to SLAM Positioning and Mapping](#32-functions-related-to-slam-positioning-and-mapping)
   - [3.3 Artifacts — Manually Mark Map Elements](#33-artifacts--manually-mark-map-elements)
   - [3.4 Motion Robot Operation Control](#34-motion-robot-operation-control)
   - [3.5 Firmware Upgrade](#35-firmware-upgrade)
   - [3.6 Statistics of Operation Data](#36-statistics-of-operation-data)
   - [3.7 Android Application Management (ARM Platform Only)](#37-android-application-management-arm-platform-only)
   - [3.8 Platform Universal Chassis and Platform-Related Functions](#38-platform-universal-chassis-and-platform-related-functions)
   - [3.9 Multi-Floor Map Management, Cross-Floor Movement](#39-multi-floor-map-management-cross-floor-movement)
   - [3.10 Delivery Service Related Interface](#310-delivery-service-related-interface)
4. [Example of Deployment Process](#4-example-of-deployment-process)
   - [4.1 Mapping](#41-mapping)
   - [4.2 Add POI](#42-add-poi)
   - [4.3 Add a Virtual Wall](#43-add-a-virtual-wall)
   - [4.4 Set Forbidden Areas](#44-set-forbidden-areas)
   - [4.5 Export Map](#45-export-map)
   - [4.6 Save the Map](#46-save-the-map)
5. [Examples of Business Processes](#5-examples-of-business-processes)
   - [5.1 Initialize the System](#51-initialize-the-system)
   - [5.2 Obtain System Resource Information](#52-obtain-system-resource-information)
   - [5.3 Obtain POI Information](#53-obtain-poi-information)
   - [5.4 Obtaining Event Information](#54-obtaining-event-information)
   - [5.5 The User Starts Operating the Robot](#55-the-user-starts-operating-the-robot)
   - [5.6 Autonomous Navigation Process Example](#56-autonomous-navigation-process-example)
   - [5.7 Delivery Business Process Example](#57-delivery-business-process-example)
   - [5.8 Automatic Recharge](#58-automatic-recharge)
   - [5.9 Robot Abnormal Recovery](#59-robot-abnormal-recovery)
   - [5.10 Motion Strategy](#510-motion-strategy)
   - [5.11 Frequently Asked Questions about Equipment Health Status](#511-frequently-asked-questions-about-equipment-health-status)
6. [QR Code Precise Docking](#6-qr-code-precise-docking)
   - [6.1 Obtain Deployed POI Information](#61-obtain-deployed-poi-information)
   - [6.2 Create Precise Docking Motion Behavior](#62-create-precise-docking-motion-behavior)
   - [6.3 After the Operation is Completed, Call the Backward Action First to Avoid Collision](#63-after-the-operation-is-completed-call-the-backward-action-first-to-avoid-collision)

---

## 1. Overview

Slamware firmware versions **4.0 and later** provide a RESTful API that is easier to use and richer than the C++ SDK. It is compatible with any client system and programming language.

The port of the service is **1448**. This article introduces the basic usage of the API. Please refer to the online document for the specific definition of the interface: [https://docs.slamtec.com/#/](https://docs.slamtec.com/#/)

If the firmware version is above **4.6.0**, you can debug the API online by entering `IP:1448` in the browser.

**Example:** `192.168.11.1:1448`

For example, the interface for connecting to the robot hotspot and obtaining the power status of the robot is as follows:

```
GET http://192.168.11.1:1448/api/core/system/v1/power/status
```

**Return content:**

```json
{
  "batteryPercentage": 90,
  "dockingStatus": "on_dock",
  "isCharging": true,
  "isDCConnected": false,
  "powerStage": "running",
  "sleepMode": "awake"
}
```

---

## 2. Interface Specifications

### 2.1 Naming Method

**API interface endpoint specification**

Most interfaces are organized in the following structure:

```
/api/{plugin}/{feature}/{version}/{resource…}
```

#### `plugin`

- **Core**: Agent core framework and general services
- **Platform**: Plug-in for universal chassis, providing basic functions such as reporting equipment events and uploading logs
- **multi_floor**: A plugin that provides multi-floor map management and cross-floor mobility capabilities, while being compatible with single-floor maps
- **Delivery**: A plugin that provides delivery services, which can be applied to restaurants, hotels and other scenarios

#### `feature`

- Robot Function Category

#### `version`

- Version number

---

### 2.2 Type of Method

Currently used **GET**, **PUT**, **POST**, **DELETE** four types of methods:

| Method | Description |
|--------|-------------|
| `GET` | Fetch resources (secure, idempotent) |
| `PUT` | Create/update resources (non-secure, idempotent) |
| `POST` | Creates a resource or performs an action (non-secure, non-idempotent) |
| `DELETE` | Delete a resource (non-secure, idempotent) |

---

### 2.3 API Parameters

#### Query Type

The query parameter is followed by a question mark with one or more pairs of `key=value`.

**Example** — obtain the POI of Building E, 2nd floor:

```
GET http://127.0.0.1:1448/api/multi-floor/map/v1/pois?floor=2F&building=E
```

#### Path Type

The path parameter is placed directly in the path.

**Example** — delete the virtual wall with id `199`:

```
DELETE http://127.0.0.1:1448/api/core/artifact/v1/lines/walls/199
```

#### Request Body

The `Content-Type` of the API request is `application/json`.

**Example:**

```bash
curl -X 'POST' \
  'http://127.0.0.1:1448/api/core/motion/v1/actions' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "action_name": "slamtec.agent.actions.MoveToAction",
    "options": {
      "target": {
        "x": 0,
        "y": 0,
        "z": 0
      },
      "move_options": {
        "mode": 0,
        "flags": [],
        "yaw": 0,
        "acceptable_precision": 0,
        "fail_retry_count": 0
      }
    }
  }'
```

---

### 2.4 Return Status Code

#### 2xx — Success

| Code | Description |
|------|-------------|
| `200 OK` | Any operation requested by the client was successfully performed |
| `204 No Content` | The server has successfully completed the request and there is no content to send in the response payload body |

#### 4xx — Client Error

| Code | Description |
|------|-------------|
| `400 Bad Request` | Generic client error state, used when there are no other 4xx error codes |
| `404 Not Found` | The URI resource requested by the REST API could not be found |

#### 5xx — Server Error

| Code | Description |
|------|-------------|
| `500 Internal Server Error` | Server internal error |

---

### 2.5 Return Value

When the interface returns a status code of `200`, `Content-Type` has the following types:

| Type | Description |
|------|-------------|
| `application/json` | Most interfaces return data in JSON format |
| `application/octet-stream` | Binary stream; obtaining the return value of explore map and stcm map is a binary stream |
| `text/plain` | The return value of some interfaces is a simple string |

---

### 2.6 How to View Interfaces in API Documentation

**Example**: Creating a new motion behavior (`/api/core/motion/v1/actions`)

#### Request/Response Overview

```
POST http://127.0.0.1:1448/api/core/motion/v1/actions
```

**Curl:**

```bash
curl -X 'POST' \
  'http://127.0.0.1:1448/api/core/motion/v1/actions' \
  -H 'accept: application/json' \
  -H 'Content-Type: application/json' \
  -d '{
    "action_name": "slamtec.agent.actions.MoveToAction",
    "options": {
      "target": {
        "x": 0.1,
        "y": 0.2
      },
      "move_options": {
        "mode": 0,
        "flags": ["with_yaw"],
        "yaw": 1
      }
    }
  }'
```

**Response Body:**

```json
{
  "action_id": 0,
  "action_name": "string",
  "stage": "GOING_TO_TARGET",
  "state": {
    "status": 0,
    "result": 0,
    "reason": ""
  }
}
```

#### Interface Request Body and Response Body Details

- Parameters marked **required** are mandatory fields.
- The **Schema** tab shows the generation rules of the JSON structure.
- Parameters with red `required` and `*` are required.
- Parameters in the schema represent field keys of the JSON structure.
- If there are ellipses in curly brackets `{item}`, you can click to view the details.
- `action_name` value example: `slamtec.agent.actions.MultiFloorMoveAction` → options in `oneOf` takes the corresponding `MultiFloorMoveActionOptions` structure.

---

## 3. Interface Classification

Classification of features described in Section 2.1.

### 3.1 System Resources

This type of interface can access the system-level resources of the robot, such as reading the power status, restarting the machine, setting system parameters, etc.

---

### 3.2 Functions Related to SLAM Positioning and Mapping

Obtain robot pose, obtain/register charging station, turn on/off mapping, obtain map data, etc.

After completing the mapping and adding the required POIs, you can call the **Get Composite Map** interface to export the map:

```
GET /api/core/slam/v1/maps/stcm
```

---

### 3.3 Artifacts — Manually Mark Map Elements

The following elements can be added to the map:

| Element | Description |
|---------|-------------|
| **Virtual Tracks** (`tracks`) | The robot can travel along a preset track through parameter control |
| **Virtual Walls** (`walls`) | Prohibit robots from entering certain areas |
| **Forbidden Area** (`forbidden_area`) | Similar to a virtual wall; supports automatic escape function. The robot will not enter the restricted area when outside; if pushed in, it can escape to the nearest edge. **Recommended over virtual walls.** |
| **Elevator Area** (`elevator_area`) | For multi-floor environments; requires elevator information and map merging via RS |
| **Dangerous Area** (`dangerous_area`) | Slopes, narrow roads, etc.; limits the maximum movement speed of the robot |
| **Coverage Area** (`coverage_area`) | The robot plans a path to cover the entire area, behaving like a sweeper |
| **Maintenance Area** (`maintenance_area`) | When the robot reopens map creation, it will only update the map in this area |

---

### 3.4 Motion Robot Operation Control

This type of interface provides:
- Obtaining all motion behaviors supported by the robot
- Obtaining/terminating/creating new motion behaviors
- Querying paths and target points
- Obtaining/setting motion strategies
- Enabling/querying manual relocation and other behavior controls

> **Note:** Robots need to create new motion behaviors to start moving.

- **Create a new motion behavior:** `POST /api/core/motion/v1/actions`
- **Query Action status:** `GET /api/core/motion/v1/actions/{action_id}`

---

### 3.5 Firmware Upgrade

This type of interface provides functions for upgrading robot firmware and querying related upgrade information.

---

### 3.6 Statistics of Operation Data

This type of interface mainly obtains the motion mileage and running time of the robot.

---

### 3.7 Android Application Management (ARM Platform Only)

This type of interface provides functions for robot installation/uninstallation of apps, as well as obtaining installed apps.

---

### 3.8 Platform Universal Chassis and Platform-Related Functions

This type of interface provides functions for obtaining robot system timestamps and obtaining robot event information.

> During the movement of the robot, it may encounter a series of situations such as obstacles, low battery, etc. The caller needs to constantly obtain event information to grasp the robot situation in real time.

- **Get event information:** `GET /api/platform/v1/events`

---

### 3.9 Multi-Floor Map Management, Cross-Floor Movement

Multi-floor map management, elevator and other functions, such as:
- Finding the nearest charging station to the robot
- Persistently saving the current map
- Reloading the map
- Synchronizing the map

---

### 3.10 Delivery Service Related Interface

> **Note:** The delivery-related interface is only available for the whole robot. The universal chassis is not supported by default. If you need to use it, please contact FAE.

Divided into three categories: **system configuration**, **cargo management**, and **task management**.

This type of interface provides well-integrated functions for delivery, guidance, and greeting scenarios. It mainly includes:
- Creating tasks
- Querying tasks
- Canceling tasks
- Pausing/continuing tasks
- Ending tasks
- Starting to pick up items
- Ending to pick up items
- Completing operations

#### Two Types of Map Operation Instructions

| Interface | Behavior |
|-----------|----------|
| `/api/core/slam` | Sets map to navigation system memory; not persistently saved; map can be exported |
| `/api/multi-floor/map` | Maps can be uploaded or retrieved from navigation memory and persisted to disk |

#### Two Types of POI Operation Instructions

| Interface | Behavior |
|-----------|----------|
| `/api/core/artifact` POI | Poi can be added, deleted, modified, and queried |
| `/api/multi-floor/map` POI | Only poi information can be found; `/api/multi-floor/map/v1/pois` provides additional `building`, `floor`, `poi_name`, `type` information |

---

## 4. Example of Deployment Process

This stage completes robot initialization operation, enables the robot, and makes it ready for use.

Mainly includes: turning on/off map creation, adding POIs, adding restricted areas, exporting maps.

> **Note:** If you use the RoboStudio deployment method, you can ignore this process.

### 4.1 Mapping

Turn on/off mapping:

```
PUT /api/core/slam/v1/mapping/:enable
```

If `enable` is set to `false` in the request body, it will close map construction. The return value `true` indicates that the operation was successful.

---

### 4.2 Add POI

```
POST /api/core/artifact/v1/pois
```

- The caller should randomly generate a **UUID** as `id`.
- `display_name` in `metadata` is used for interface display.
- `type` is used to distinguish POI types.

> When adding POI during the mapping process, it is recommended **not** to include `Pose`. The POI will be created with the current position of the robot, and the sensor observation information will be recorded. Pose adjustment will be performed after the loop is closed.

Please click on **Schema** in the interface documentation for detailed instructions.

---

### 4.3 Add a Virtual Wall

Add virtual line segment:

```
POST /api/core/artifact/v1/lines/{usage}
```

> `ID` is an invalid field when added and can be any value.

Please click on **Schema** in the interface documentation for detailed instructions.

---

### 4.4 Set Forbidden Areas

Add rectangular area:

```
POST /api/core/artifact/v1/rectangle-areas/{usage}
```

Different types of rectangular areas require different `metadata`. Please click on the interface document **Schema** to view detailed instructions.

---

### 4.5 Export Map

Obtain composite map:

```
GET /api/core/slam/v1/maps/stcm
```

The composite map contains all data. The response message is binary `ByteFlow` and can be directly saved as a `.stcm` file.

---

### 4.6 Save the Map

#### Option 1: Save via API

Follow the interface calling sequence to save the map:

1. **Upload the map to the robot**

   Uploaded maps will be persistently saved in the file system but will not be loaded into Slamware.

   > [Note] When the robot is managed by the cloud, maps downloaded from the cloud will overwrite local maps.

   ```
   POST /api/multi-floor/map/v1/stcm
   ```

2. **Reload the map**

   ```
   POST /api/multi-floor/map/v1/stcm/:reload
   ```

3. **Persistently save the current map**

   ```
   POST /api/multi-floor/map/v1/stcm/:save
   ```

#### Option 2: After Loading via RoboStudio

After loading the map into Slamware through RoboStudio, the caller only needs to call the **synchronize map** interface:

```
POST /api/multi-floor/map/v1/stcm/:sync
```

---

## 5. Examples of Business Processes

### 5.1 Initialize the System

When the robot starts, use the following polling interface to determine whether the system has completed initialization.

> Only when each component is **enabled** can the robot enter normal business logic.

```
GET /api/core/system/v1/capabilities
```

---

### 5.2 Obtain System Resource Information

Through the API, you can obtain key data such as device information, device health status, power status, etc., to ensure the normal operation and reasonable use of the robot.

| Description | Endpoint |
|-------------|----------|
| Get robot power status | `GET /api/core/system/v1/power/status` |
| Obtain device information | `GET /api/core/system/v1/robot/info` |
| Obtain device health status | `GET /api/core/system/v1/robot/health` |
| Obtain system parameters | `GET /api/core/system/v1/parameter` |
| Set system parameters | `PUT /api/core/system/v1/parameter` |
| Turn off or restart the robot | `POST /api/core/system/v1/power/:shutdown` |

---

### 5.3 Obtain POI Information

Get the list of POI information set in the map (i.e., target points that the robot can navigate to):

```
GET /api/multi-floor/map/v1/pois
```

Through parameters, you can get the POI list of specified floors, buildings, POI types, and POI groups. Get all POIs without parameters.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `floor` | string | No | Floor name |
| `building` | string | No | Building name |
| `type` | string | No | POI type |
| `group` | string | No | POI grouping |

---

### 5.4 Obtaining Event Information

When the robot is turned on, it obtains the current status, charging status, operating status, possible obstacles, elevator entry/exit, health status alarm, etc. It is recommended to poll this interface in real time.

```
GET /api/platform/v1/events
```

The robot notifies the calling of its own events. Enabling different plugins will expand different event types.

- **`GeneralEventType`**: General event
- **`ElevatorEventType`**: Event related to entering and leaving elevators
- **`DeliveryEventType`**: Delivery-related event

> The caller only needs to handle the events they care about.

Please click on **Schema** in the interface documentation for detailed instructions.

---

### 5.5 The User Starts Operating the Robot

When the user operates the robot:

1. Call the following interface with `enable_task_execution` set to `false` to **prohibit the robot from moving**.
2. When the user completes the operation, set `enable_task_execution` to `true` to **allow the robot to move**.

```
PUT /api/delivery/v1/tasks/:task_execution
```

**Request body example:**

```json
{
  "enable_task_execution": false
}
```

---

### 5.6 Autonomous Navigation Process Example

This example shows how a caller creates an action to control the robot to move to a specified POI.

#### 5.6.1 Obtain All Supported Actions

```
GET /api/core/motion/v1/action-factories
```

| Action Name | Description |
|-------------|-------------|
| `slamtec.agent.actions.MoveToAction` | Autonomous navigation move |
| `slamtec.agent.actions.MultiFloorMoveAction` | Autonomous navigation move; supports cross-floor, POI target points, multi-level scheduling |
| `slamtec.agent.actions.MultiFloorBackHomeAction` | Cross-floor autonomous return to charging station |
| `slamtec.agent.actions.SeriesMoveToAction` | Autonomous navigation move with multiple target points |
| `slamtec.agent.actions.MoveByAction` | Remote movement; needs to be called regularly to achieve continuous motion effect |
| `slamtec.agent.actions.GoHomeAction` | Autonomous return to charging station |
| `slamtec.agent.actions.RotateToAction` | Rotate in place to a specified angle |
| `slamtec.agent.actions.RotateAction` | Rotate in place to a specified angle |
| `slamtec.agent.actions.MoveToTagAction` | QR code precise docking |
| `slamtec.agent.actions.BackOffFromTagAction` | Back from QR code to prevent collision |
| `slamtec.agent.actions.RecoverLocalizationAction` | Relocation |
| `slamtec.agent.actions.ManualRelocalizationAction` | Manual relocation |
| `slamtec.agent.actions.SweepAction` | Coverage planning movement; suitable for cleaning/disinfection scenarios (requires firmware ≥ 4.4) |
| `slamtec.agent.actions.ReturnToParkingAction` | Autonomous return to standby point (POI type: PARKING); supports multi-machine obstacle avoidance and queuing (requires Lora module; firmware ≥ 4.5.5) |

---

#### 5.6.2 Creating New Motion Behaviors

```
POST /api/core/motion/v1/actions
```

Use `action_name`: `slamtec.agent.actions.MultiFloorMoveAction`.

**Request body example** — moves to A101 with precise-to-point and precise-to-angle (yaw value in radians: `1.0`):

```json
{
  "action_name": "slamtec.agent.actions.MultiFloorMoveAction",
  "options": {
    "target": {
      "poi_name": "A101"
    },
    "move_options": {
      "mode": 2,
      "flags": ["with_yaw", "precise"],
      "yaw": 1,
      "acceptable_precision": 0,
      "fail_retry_count": 0
    }
  }
}
```

**Parameter Details:**

1. **`mode`** (default: `0`)

   | Value | Description |
   |-------|-------------|
   | `0` | Free navigation |
   | `1` | Force following track mode (stop and wait in case of obstacles) |
   | `2` | Track priority mode (orbiting when encountering obstacles) |

2. **`flags`**

   | Flag | Description |
   |------|-------------|
   | `precise` | Precise-to-point mode; makes the robot more accurate to the point |
   | `with_yaw` | Precise-to-angle mode; the `yaw` field only takes effect if this flag is included |
   | `fail_retry_count` | Specify the number of retries after a search failure; uses default configuration if not specified |
   | `find_path_ignoring_dynamic_obstacles` | Ignore dynamic obstacles when searching for paths; suitable for crowded and narrow areas |

3. **`yaw`**: The orientation of the robot after reaching the target point (accurate to the angle).

4. **`acceptable_precision`**: The acceptable range to the target point. When the target point is occupied, the distance between the robot and the target point is considered successful within this range. Default: `0.1` meters or `0.18` meters. Does not affect the robot's accuracy to the point.

5. **`fail_retry_count`**: Number of failed retries.

---

#### 5.6.3 Query Action Status

When creating an action, an `action_id` will be returned. Query the current status of the action based on this ID. During the operation of the robot, it is necessary to view the action status in real time through polling this interface.

```
GET /api/core/motion/v1/actions/{action_id}
```

---

#### 5.6.4 Termination of Current Action

```
DELETE /api/core/motion/v1/actions/:current
```

---

### 5.7 Delivery Business Process Example

> **Note:** The delivery-related interface is only available for the whole robot. The universal chassis is not supported by default. If you need to use it, please contact FAE.

The interface endpoint prefix is `/api/delivery`.

Divided into three categories: **system configuration**, **cargo management**, and **task management**.

---

#### 5.7.1 User Starts Operating the Robot

When the user operates the robot:

1. Call the following interface with `enable_task_execution` set to `false` to prohibit the robot from moving.
2. Set `enable_task_execution` to `true` when the user completes the operation.

```
PUT /api/delivery/v1/tasks/:task_execution
```

```json
{
  "enable_task_execution": false
}
```

---

#### 5.7.2 Get Settings Related to Shipping

Users can obtain the low battery scenario configuration and timeout scenario configuration of the robot through the APP.

```
GET /api/delivery/v1/settings
```

Please click on **Schema** in the interface documentation for detailed instructions.

---

#### 5.7.3 Set the Timeout for the Task

```
PUT /api/delivery/v1/settings/timeout
```

**Request body example** — `food_pickup_timeout` represents the waiting time after reaching the target point, in seconds:

```json
{
  "food_pickup_timeout": 0
}
```

---

#### 5.7.4 Operate Cargo

If it is an H2 hotel delivery robot, the APP opens/closes the cabin door through the cargo interface. If there is no cargo, please ignore this operation.

```
PUT /api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}
```

---

#### 5.7.5 Create Tasks

After the user puts in the item and closes the cabin door, call the following interface to create a task:

```
POST /api/delivery/v1/tasks
```

Currently supported task types:

| Type | Description |
|------|-------------|
| `TAKEOUT` | Delivery |
| `GUIDE` | Guide |
| `FOOD_DELIVERY` | Food delivery |
| `RECYCLE` | Return disc |
| `RETURN` | Return |
| `DISINFECT` | Disinfection |

To create multiple tasks at once:

```
POST /api/delivery/v1/tasks/:batch
```

---

#### 5.7.6 Start the Mission

Set `enable_task_execution` to `true` to allow the robot to move. The robot will perform any pending task, or return to the charging station/PARKING POI if there is none.

```
PUT /api/delivery/v1/tasks/:task_execution
```

```json
{
  "enable_task_execution": true
}
```

---

#### 5.7.7 Suspend/Resume Task

```
PUT /api/delivery/v1/tasks/:task_execution
```

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `enable_task_execution` | boolean | Yes | `true`: Continue the task / `false`: Pause task |

**Request body example:**

```json
{
  "enable_task_execution": false
}
```

---

#### 5.7.8 Query Task Status

During the robot's task execution, the APP needs to regularly query the task status:

```
GET /api/delivery/v1/stage
```

| Stage | Description |
|-------|-------------|
| `DEVICE_ERROR` | Equipment failure; robot cannot move; caller should display the fault page |
| `GOING_TO_TASK_POINT` | On the way to the task point |
| `ARRIVED_AT_TASK_POINT` | Robot reached the task point; waits for operation to complete or timeout |
| `ON_DELIVERING` | Going to the target point (compatible name; not necessarily a delivery task) |
| `ARRIVED_AT_TARGET` | Reached the final target point |
| `ON_RETURNING` | Returning to the default docking point |
| `GOING_HOME` | Returning to the charging station |
| `IDLE` | Robot is at the default docking point or charging station |

For emergencies such as path being blocked or receiving new tasks, poll the event interface:

```
GET /api/platform/v1/events
```

---

#### 5.7.9 Start Taking Items

When the APP queries that the stage is `ARRIVED_AT_TARGET`, display the pick-up interface.

If the robot has Cargo, display an **"open cabin"** button. After the user clicks it, call:

```
PUT /api/delivery/v1/tasks/:start_pickup
```

Then call the cargo interface to open the cabin door.

---

#### 5.7.10 Complete the Retrieval

After the user completes the pick-up action, notify the robot:

```
PUT /api/delivery/v1/tasks/:end_pickup
```

---

#### 5.7.11 The Next Mission

Allow the robot to move autonomously. If there is a next task, it will continue; otherwise, it will return to the charging station or standby point (POI type: PARKING).

```
PUT /api/delivery/v1/tasks/:task_execution
```

---

### 5.8 Automatic Recharge

```
POST /api/core/motion/v1/actions
```

Use `action_name`: `slamtec.agent.actions.MultiFloorBackHomeAction`

**Request body example:**

```json
{
  "action_name": "slamtec.agent.actions.MultiFloorBackHomeAction",
  "options": {}
}
```

---

### 5.9 Robot Abnormal Recovery

Set the robot pose to the specified POI. Generally used for recovery operations after an abnormality occurs, such as restoring robot positioning information.

> **Note:** This interface is usually called after positioning is lost. If the robot is pushed back to the charging station, there is no need to call this interface — the system will automatically perform recovery actions.

```
PUT /api/multi-floor/localization/v1/pose
```

---

### 5.10 Motion Strategy

The motion strategy is a series of internal parameters of Slamware, involving motion speed and obstacle avoidance behavior. Different strategies can be applied to different scenarios. Generally, the default strategy can be used.

> Minimum firmware version required: **4.2.4**

| Description | Endpoint |
|-------------|----------|
| Get all supported motion strategies | `GET /api/core/motion/v1/strategies` |
| Get the current motion strategy | `GET /api/core/motion/v1/strategies/:current` |
| Set motion strategy | `PUT /api/core/motion/v1/strategies/:current` |

---

### 5.11 Frequently Asked Questions about Equipment Health Status

| Description | Trigger Condition | Chassis Action | Level | BigInt | Display Information | Remove Method |
|-------------|-------------------|---------------|-------|--------|---------------------|---------------|
| **Emergency stop switch** | Emergency stop switch trigger | Emergency stop; speed no longer responds | ERROR | `0x02010100` | "system emergency stop" | Release emergency stop switch recovery |
| **Low battery alarm** | Battery less than 15% | No action | WARNING | `0x01020100` | "power low" | Battery higher than 15% |
| **Low battery alarm** | Battery less than 5% | Automatic shutdown, all power off | — | — | "power low" | Battery higher than 5% |
| **Brake release** | Brake release button triggered | Motor loosens shaft; speed no longer responds | ERROR | `0x02010700` | "motor brake released" | Brake release button recovery |
| **Drive motor alarm** | Drive motor driver alarm | Shaft loose when overcurrent alarms occur | WARNING | `0x01030100` | "motor alarm" | Chassis firmware tries to clear the drive alarm automatically |
| **Motor odometer alarm** | Motor stopped (no speed or speed is 0), motor moves more than 500mm | No action | WARNING | `0x0103010x` | "motor[y]odometry alarm" | Speed, or brake release |
| **Platform watchdog trigger** | Platform firmware watchdog triggered; firmware restarted | Speed no longer responds after firmware restart | ERROR | `0x02030400` | "watchdog overflow" | Manual remove |
| **Magnetic sensor trigger** | Magnetic sensor trigger | Stop immediately; speed no longer responds | ERROR | `0x0204050x` | "magtape[x] triggered" | Emergency stop switch trigger, or manually clear |
| **Magnetic sensor communication error** | Magnetic sensor communication error | Stop immediately; speed no longer responds | FATAL | `0x0404060x` | "magnetic[x]:y." | Check connection cable and sensor; manually clear the error; restart chassis if necessary |
| **TOF cliff communication error** | TOF cliff communication error | Stop immediately; speed no longer responds | FATAL | `0x0404020x` | "cliff[x]:y." | Check connection cable and sensor; manually clear the error; restart chassis if necessary |
| **Collision sensor error** | Collision continuously triggers and walks forward 200mm | Stop immediately; speed no longer responds | ERROR | `0x0204010x` | "bumper sensor error" | Collision signal remove |
| **TOF cliff signal error** | TOF cliff continuously triggers and walks forward 200mm | Stop immediately; speed no longer responds | ERROR | `0x0204020x` | "cliff sensor error" | TOF cliff signal remove |
| **Low positioning error** | — | — | ERROR | `0x02010900` | "Low localization due to great environmental changes, because visual coarse poses received" or "Low localization quality" | — |
| **Relocation error** | — | — | ERROR | `0x02010800` | "Relocalization has failed last time, clear the error to move" | — |
| **Online SLAM reboot / Location abnormally** | Online SLAM restart | No action | ERROR | `0x02010600` | "slamware has rebooted, clear the error to move" | Push the robot back to the charging pile (if it appears multiple times in the same area, please rebuild the map) |

---

## 6. QR Code Precise Docking

Please refer to the **"QR code precise docking deployment manual"** for the QR code precise docking deployment process.

This section only explains how to use API calls.

### 6.1 Obtain Deployed POI Information

```
GET /api/core/artifact/v1/pois
```

Find the POI with `type: TAG`. Find the required POI according to the `display_name`, record the `pose` and `tag_ids`, and fill in the `MoveToTagAction` parameter.

> If `tag_ids` information is missing, please check the POI deployment process.

---

### 6.2 Create Precise Docking Motion Behavior

```
POST /api/core/motion/v1/actions
```

Use `action_name`: `slamtec.agent.actions.MoveToTagAction`

**Request body example:**

```json
{
  "action_name": "slamtec.agent.actions.MoveToTagAction",
  "options": {
    "target": {
      "x": 0.590,
      "y": 0.110,
      "yaw": -3.130
    },
    "tag_ids": [0, 50],
    "relative_pose_to_tag": {
      "x": 0.4,
      "y": 0.0
    }
  }
}
```

- `target` and `tag_ids` data are recorded from **6.1**.
- `relative_pose_to_tag` field can be left blank:
  - `x`: longitudinal distance from the center of the QR code
  - `y`: lateral deviation from the center of the QR code
- If not filled in, QR code docking defaults to `precise_move_to_tag` `safe_distance_to_tag` (7.5cm) as the default value of `x`; default value of `y` is `0`.

---

### 6.3 After the Operation is Completed, Call the Backward Action First to Avoid Collision

```
POST /api/core/motion/v1/actions
```

Use `action_name`: `slamtec.agent.actions.BackOffFromTagAction`

**Request body example:**

```json
{
  "action_name": "slamtec.agent.actions.BackOffFromTagAction",
  "options": {}
}
```
