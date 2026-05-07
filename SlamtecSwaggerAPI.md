# Slamware RESTful API

**Version:** 1.1.1 | OAS3

This document is applicable to the general chassis and service robots of Slamtec.


---

## Table of Contents

- [system](#system) — System resource
- [slam](#slam) — Localization and mapping
- [artifact](#artifact) — Semantic elements
- [motion](#motion) — Robot motion control
- [firmware](#firmware) — Firmware upgrade
- [statistics](#statistics) — Statistics
- [sensors](#sensors) — Sensor control
- [application](#application) — Android application management(ARM only)
- [platform](#platform) — Robot chassis and platform
- [multi-floor](#multi-floor) — Map management and across floor movement
- [industry](#industry) — For industrial chassis
- [delivery](#delivery) — Delivery service (specific models are required, not supported on chassis)

---

## system

System resource

### `GET` `/api/core/system/v1/capabilities`

**Get robot capabilities**

This API is used to determine which functions the robot supports and whether initialization has been completed. Some interfaces in this document require specific capabilities to run. Required minimum firmware version 4.2.0

**Responses:**

- **200** — OK

  **`Capability`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `name` | `string` — *slamware. Agent. Core general functions of chassis core such as localization and navigation slamware. Agent. Platform general functions of platform such as log collection slamware. Agent. Multi_ Floor multi floor map management and cross floor mobile function slamware. Agent. Delivery distribution service function slamware. Agent. Mercury2 intelligent hotel distribution service function supporting cloud scheduling — Enum: `[ slamware.agent.core, slamware.agent.platform, slamware.agent.multi_floor, slamware.agent.delivery, slamware.agent.mercury2 ]` |
  | `version` | `string` |
  | `enabled` | `boolean` — If the plug-in fails to initialize, or the value is false when it is just started, the upper computer should continue to wait for a period of time until the value becomes true or timeout. |


---

### `GET` `/api/core/system/v1/power/status`

**Get power status**

**Responses:**

- **200** — OK

  **`PowerStatus`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `batteryPercentage` | `integer` — Battery power percentage，0 ~ 100 |
  | `dockingStatus` | `string` — Pile alignment status — Enum: `[ on_dock, not_on_dock ]` |
  | `isCharging` | `boolean` — Whether it is charging |
  | `isDCConnected` | `boolean` — Whether the external power is connected |
  | `powerStage` | `string` — Power status — Enum: `[ starting, running, restarting, shutingdown, error ]` |
  | `sleepMode` | `string` — sleepMode — Enum: `[ awake, waking_up, asleep ]` |


---

### `POST` `/api/core/system/v1/power/:shutdown`

**Shutdown or restart the robot**

**Request Body** (`application/json`):

> Robot will shutdown after shutdown_time_interval minutes and restart after restart_time_interval minutes. if restart_time_interval is 0, robot won't restart again.

| Field | Type / Description |
|-------|-------------------|
| `shutdown_time_interval` | `integer` — Shutdown time |
| `restart_time_interval` | `integer` — Restart time |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `POST` `/api/core/system/v1/power/:hibernate`

**Hibernate the robot**

Lidar will stop working when robot hibernate

**Responses:**

- **200** — OK

---

### `POST` `/api/core/system/v1/power/:wakeup`

**Wakeup the robot**

**Responses:**

- **200** — OK

---

### `POST` `/api/core/system/v1/power/:restartmodule`

**Restart module**

**Request Body** (`application/json`):

> restart the specified module

| Field | Type / Description |
|-------|-------------------|
| `mode` | `string` — RestartModeSoft restart software(slamwared) RestartModeHard restart operation system RestartModeBase restart base — Enum: `[ RestartModeSoft, RestartModeHard, RestartModeBase ]` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/system/v1/robot/info`

**Get device information**

**Responses:**

- **200** — OK

  **`DeviceInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `manufacturerId` | `integer` — Manufacturer ID |
  | `manufacturerName` | `string` — Manufacturer name |
  | `modelId` | `integer` — Equipment model ID |
  | `modelName` | `string` — Equipment model name |
  | `deviceID` | `string` — Serial number |
  | `hardwareVersion` | `string` — Hardware version number |
  | `softwareVersion` | `string` — Software version number |


---

### `GET` `/api/core/system/v1/robot/health`

**Get health information**

Get device health information

**Responses:**

- **200** — OK

  **`BaseHealthInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `hasWarning` | `boolean` — Whether there is alarm information |
  | `hasError` | `boolean` — Whether there is an error message |
  | `hasFatal` | `boolean` — Whether there is a fatal error |
  | `baseError` | `integer` — 0 User, 1 System, 2 Power, 3 Motion, 4 Sensor, 255 Unknown — Enum: `[ 0, 1, 2, 3, 4, 255 ]` |
  | `id` | `integer` |
  | `component` | `integer` — 0 User, 1 System, 2 Power, 3 Motion, 4 Sensor, 255 Unknown — Enum: `[ 0, 1, 2, 3, 4, 255 ]` |
  | `errorCode` | `integer` |
  | `level` | `integer` — 0 Healthy, 1 Warn, 2 Error, 4 Fatal, 255 Unknown — Enum: `[ 0, 1, 2, 4, 255 ]` |
  | `message` | `string` |


---

### `DELETE` `/api/core/system/v1/robot/health/{error_code}`

**Remove error status**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `error_code` *(required)* | `integer` | path | error code |

**Responses:**

- **200** — OK
- **400** — Invalid Error Code

---

### `GET` `/api/core/system/v1/laserscan`

**Get current laser scan**

Required minimum firmware version 4.2.2

**Responses:**

- **200** — OK

  **`LaserScan`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Pose is the robot pose when it observe the laser scan.*
  | `pose` | `number` — Pose in 3D space |
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |
  | `laser_points` | `number` |
  | `distance` | `number` |
  | `angle` | `number` |
  | `valid` | `boolean` |


---

### `GET` `/api/core/system/v1/parameter`

**Get system parameters**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `param` *(required)* | `string` | query | system parameter name: base.max_moving_speed - maximum linear speed base.max_angular_speed - maximum angular speed docking.docked_register_strategy - charging dock registration strategy，always: register every time back to the charging dock，when_not_exists: register when there is no charging dock registered in the map |

**Responses:**

- **200** — OK

  Returns: `string`

- **400** — Parameter is required

---

### `PUT` `/api/core/system/v1/parameter`

**Set system parameters**

**Request Body** (`application/json`):

> The set system parameters are only valid for this run, and will be restored to original values after robot restarted.

| Field | Type / Description |
|-------|-------------------|
| `param` | `string` — system parameter name: base.max_moving_speed - Maximum linear speed m/s base.max_angular_speed - Maximum angular speed rad/s base.emergency_stop - value of on means emergency stop is triggered, off means emergency stop is eliminated base.brake_release - value of on means the brake is released, off means the brake is restored — Enum: `[ base.max_moving_speed, base.max_angular_speed, base.emergency_stop, base.brake_release ]` |
| `value` | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`

- **400** — Bad Request

---

### `GET` `/api/core/system/v1/network/status`

**Get network status**

**Responses:**

- **200** — OK

  **`NetworkStatus`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `networkstatus` | `string` — Ethernet address — Enum: `[ STA, AP ]` |
  | `ethip1` | `string` — Ethernet address |
  | `ip` | `string` — IP address |
  | `mac` | `string` — MAC address |
  | `mode` | `string` — Enum: `[ STA, AP ]` |
  | `quality` | `integer` — network quality |
  | `ssid` | `string` |


---

### `PUT` `/api/core/system/v1/network/status`

**Set network status**

When the network is managed by the Android system, this API will return false

**Request Body** (`application/json`):

> Set NetworkMode.For example, when you set Station mode, you need to pass in ssid and password

| Field | Type / Description |
|-------|-------------------|
  *Set NetworkMode.For example, when you set Station mode, you need to pass in ssid and password*
| `networkmode` | `integer` — Network mode: 0 - Set to AP mode 1 - Set to Station mode 2 - Disable WIFI 3 - Disable DHCP 4 - Enable DHCP — Enum: `[ 0, 1, 2, 3, 4 ]` |
| `options` | `string` — ssid和password均为可选项，默认从配置文件设置热点名称 |
  *ssid和password均为可选项，默认从配置文件设置热点名称*
| `ssid` | `string` |
| `password` | `string` |
| `ssid` *(required)* | `string` |
| `password` *(required)* | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`

- **400** — Bad Request

---

### `GET` `/api/core/system/v1/network/route`

**Get routing information**

**Responses:**

- **200** — OK

  **`RouteStatus`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `priority` | `string` — Routing priority selection: WiFi - WIFI priority 4G - 4G priority — Enum: `[ wifi, 4g ]` |

- **500** — Failed to get route

---

### `PUT` `/api/core/system/v1/network/route`

**Set routing information**

Set routing priority. When both wifi and 4g are available, you can choose which one is prior.

**Request Body** (`application/json`):

> Routing priority selection: WiFi - WIFI priority 4G - 4G priority

| Field | Type / Description |
|-------|-------------------|
| `priority` | `string` — Routing priority selection: WiFi - WIFI priority 4G - 4G priority — Enum: `[ wifi, 4g ]` |

**Responses:**

- **200** — OK
- **400** — Invalid JSON data
- **500** — Failed to set route

---

### `GET` `/api/core/system/v1/network/apn`

**Get cmlink apn**

Required minimum firmware version 4.4.0

**Responses:**

- **200** — OK

  **`ApnStatus`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `apn` | `string` — 根据地区选择对应apn: 例如香港地区：cmhk |

- **500** — Failed to get cmlink apn

---

### `PUT` `/api/core/system/v1/network/apn`

**Set cmlink apn**

Set cmlink apn in different regions. Please refer to the official website of the operator for specific apn Required minimum firmware version 4.4.0

**Request Body** (`application/json`):

> 根据地区选择对应apn: 例如香港地区：cmhk

| Field | Type / Description |
|-------|-------------------|
| `apn` | `string` — 根据地区选择对应apn: 例如香港地区：cmhk |

**Responses:**

- **200** — OK
- **500** — Failed to set cmlink apn

---

### `PUT` `/api/core/system/v1/cube/config`

**Set Cube configuration**

Read a cube_cfg_dat as the Request Body. Please export the Cube configuration file with RoboStudio's Cube configuration tool or contact Slamtec technical support to obtain it. Required minimum firmware version 4.2.0

**Responses:**

- **200** — OK

---

### `POST` `/api/core/system/v1/light/control`

**set light control**

Set different channels, different parts, different types of led light color effects.

**Request Body** (`application/json`):

> led control channel: One, Two

| Field | Type / Description |
|-------|-------------------|
| `channel` | `string` — led control channel: One, Two — Enum: `[ One, Two ]` |
| `controlPart` | `string` — led control part: Left, Right — Enum: `[ Left, Right ]` |
| `mode` | `string` — led控 channel mode: AlwaysBright, Breathe, Blink, HorseLamp — Enum: `[ AlwaysBright, Breathe, Blink, HorseLamp ]` |
| `color` | `integer` — Bright, Blink, and Horse Lamp mode indicates the color of the set led, and breathing mode indicates the color of the led at the beginning of breathing (usually black) |
  *Bright, Blink, and Horse Lamp mode indicates the color of the set led, and breathing mode indicates the color of the led at the beginning of breathing (usually black)*
| `red` | `integer` |
| `green` | `integer` |
| `blue` | `integer` |
| `brightnessEndColor` | `integer` — The breath mode indicates the color of the led at the end of the breath (set the color that the breath wants to achieve), and the set value needs to be larger than the color value; The remaining patterns can be filled with arbitrary values |
  *The breath mode indicates the color of the led at the end of the breath (set the color that the breath wants to achieve), and the set value needs to be larger than the color value; The remaining patterns can be filled with arbitrary values*
| `red` | `integer` |
| `green` | `integer` |
| `blue` | `integer` |
| `brightMs` | `integer` — Bright mode can be filled with any value; The breathing mode is filled with the single change time of brightness (the single change indicates the time of each increase of color by 1); Blink mode fills in the duration of lighting; Horse Lmap mode indicates the time to light the next light |
| `offMs` | `integer` — Blink mode fills in the duration of extinction; Other modes can be filled with arbitrary values |

**Responses:**

- **200** — OK
- **400** — Invalid JSON data
- **500** — Failed to set light control

---

### `POST` `/api/core/system/v1/aeb/control`

**set aeb control**

Set AEB On/Off.

**Request Body** (`application/json`):


Type: `string` — Enum: `[ On, Off ]`

**Responses:**

- **200** — OK
- **400** — Invalid JSON data
- **500** — Failed to set aeb control

---

### `POST` `/api/core/system/v1/jack/status`

**set jack status**

set jack status.

**Request Body** (`application/json`):


Type: `string` — Enum: `[ Up, Down, Stop, ClearAlarm ]`

**Responses:**

- **200** — OK
- **400** — Invalid JSON data
- **500** — Failed to set aeb control

---

### `GET` `/api/core/system/v1/jack/status`

**get jack status**

get jack status.

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `actual_pos` | `integer` |
  | `alarm` | `integer` |
  | `drv_status` | `integer` |
  | `stage` | `integer` |


---

### `GET` `/api/core/system/v1/battery/pack`

**get battery pack current and temperature**

get battery pack current and temperature

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `current` | `number` — current value of battery pack, unit mA, positive means charging, negative means discharging |
  | `temp_count` | `integer` — count of sensors in battery pack |
  | `temp` | `integer` — temperature value of battery pack, unit 0.1 Celsius |


---

### `GET` `/api/core/system/v1/rawadcimu`

**get IMU ADC raw data**

get robot IMU ADC raw data

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `acc_x` | `integer` |
  | `acc_y` | `integer` |
  | `acc_z` | `integer` |
  | `gyro_x` | `integer` |
  | `gyro_y` | `integer` |
  | `gyro_z` | `integer` |
  | `comp_x` | `integer` |
  | `comp_y` | `integer` |
  | `comp_z` | `integer` |
  | `timestamp` | `integer` |


---

### `GET` `/api/core/system/v1/rawimu`

**get IMU raw data**

get robot IMU raw data

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `acc_x` | `number` |
  | `acc_y` | `number` |
  | `acc_z` | `number` |
  | `gyro_x` | `number` |
  | `gyro_y` | `number` |
  | `gyro_z` | `number` |
  | `comp_x` | `number` |
  | `comp_y` | `number` |
  | `comp_z` | `number` |
  | `timestamp` | `integer` |


---

## slam

Localization and mapping

### `GET` `/api/core/slam/v1/localization/pose`

**Get the robot pose**

**Responses:**

- **200** — OK

  **`Pose3D`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |


---

### `PUT` `/api/core/slam/v1/localization/pose`

**Set the robot pose**

Set the robot pose into the particular position in the map

**Request Body** (`application/json`):

> Pose in 3D space

| Field | Type / Description |
|-------|-------------------|
  *Pose in 3D space*
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `yaw` | `number` |
| `pitch` | `number` |
| `roll` | `number` |

**Responses:**

- **200** — OK
- **400** — Invalid Argument

---

### `GET` `/api/core/slam/v1/localization/odopose`

**Get the robot pose by odometry**

**Responses:**

- **200** — OK

  **`Pose3D`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |


---

### `GET` `/api/core/slam/v1/localization/quality`

**Get localization quality**

localization quality range 0 ~ 100

**Responses:**

- **200** — OK

  Returns: `integer`


---

### `GET` `/api/core/slam/v1/localization/:enable`

**Localization is enabled or not**

The return value true indicates that localization is supported by using multi-sensor fused way, and value false indicates localization is just using odometry

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/slam/v1/localization/:enable`

**Enable / Disable localization**

The return value true means the operation was successful

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

  Returns: `boolean`

- **400** — Bad Request

---

### `POST` `/api/core/slam/v1/localization/status/:reset`

**Reset localization status**

Reset localization status

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/slam/v1/mapping/:enable`

**Mapping is enabled or not**

The return value true represents the mapping mode, false represents the localization mode

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/slam/v1/mapping/:enable`

**Enable / Disable mapping**

The return value true means the operation was successful

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

  Returns: `boolean`

- **400** — Bad Request

---

### `GET` `/api/core/slam/v1/loopclosure/:enable`

**Loop closure is enabled or not**

Required minimum firmware version 4.2.2

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/slam/v1/loopclosure/:enable`

**Enable / Disable loop closure**

response true means operation succeededRequired minimum firmware version 4.2.2

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

  Returns: `boolean`

- **400** — Bad Request

---

### `GET` `/api/core/slam/v1/homepose`

**Get pose of current charging dock**

If there is no charging dock in the current map, a 404 error will be returned

**Responses:**

- **200** — OK

  **`Pose3D`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |

- **404** — Home dock not found

---

### `PUT` `/api/core/slam/v1/homepose`

**Set current charging dock pose**

Set the current charging dock position. When there are multiple charging docks in the map, the APP needs to set one of them as the currently used home dock.

**Request Body** (`application/json`):

> Pose in 3D space

| Field | Type / Description |
|-------|-------------------|
  *Pose in 3D space*
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `yaw` | `number` |
| `pitch` | `number` |
| `roll` | `number` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/slam/v1/homedocks`

**Get all charging docks**

Get all charging docks.Required minimum firmware version 4.3.2

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `PUT` `/api/core/slam/v1/homedocks`

**Set charging docks**

Set charging docks to robot.Required minimum firmware version 4.3.2

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `id` *(required)* | `string` |
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `metadata` | Different API require different metadata. |
  *Different API require different metadata.*

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `POST` `/api/core/slam/v1/homedocks`

**Add a charging dock**

Add a charging dock to robot, metadata should contains display_name.

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `id` *(required)* | `string` |
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `metadata` | Different API require different metadata. |
  *Different API require different metadata.*

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/slam/v1/homedocks`

**Clear all charging docks**

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `POST` `/api/core/slam/v1/homedocks/:register`

**Register a charging dock**

Register a charging dock on the map based on the robot’s current location.

**Request Body** (`application/json`):

> display_name is home dock name for display

| Field | Type / Description |
|-------|-------------------|
  *display_name is home dock name for display*
| `metadata` | `string` |
| `display_name` | `string` |

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `PUT` `/api/core/slam/v1/homedocks/{dock_id}`

**Edit charging dock**

Edit charging dock, pose and metadata can be modified.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `dock_id` *(required)* | `string($uuid)` | path |  |

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `metadata` |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/slam/v1/homedocks/{dock_id}`

**Remove a charging dock**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `dock_id` *(required)* | `string($uuid)` | path |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/slam/v1/imu`

**Get IMU**

Get Imu in robot coordinate.

**Responses:**

- **200** — OK

  **`ImuData`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *IMU data*
  | `acc` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `availibilityBitMap` | `integer` — 1 quaternion is valid 2 accelerometer is valid 4 gyro is valid 8 compass is valid 16 raw accelerometer is valid 32 raw gyro is valid 64 raw compass is valid 128 6D pose is valid 256 9D pose is valid 512 Euler angles is valid |
  | `compass` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `euler_angle` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `gyro` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `quaternion` | `number` |
  | `w` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `raw_acc` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `raw_compass` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `raw_gyro` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `timestamp` | `integer` — milliseconds since the chassis starts |


---

### `GET` `/api/core/slam/v1/knownarea`

**Get known area**

The known area is the range of the explored map.

**Responses:**

- **200** — OK

  **`Rectangle`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `x` | `number` |
  | `y` | `number` |
  | `width` | `number` |
  | `height` | `number` |


---

### `GET` `/api/core/slam/v1/maps/explore`

**Get explore map**

Get the grip map of laser exploration. You can specify the range of acquisition through min_x, min_y, max_x, max_y. The response data is a byte stream, and the first 32 bytes are metadata (little-endian)，followed by map data. IndexData TypeDescription0-3floatThe X coordinate of the map position4-7floatThe Y coordinate of the map position8-11uint32grid number of X dimension12-15uint32grid number of Y dimension16-19floatmap resolution, side length of each grid, in meters20-31byte[]Reserved32-35uint32byte number of following data，it should be equal to dimension_x * dimension_y36-Endbyte[]map data，each byte represents a grid

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `min_x` | `number` | query |  |
| `min_y` | `number` | query |  |
| `max_x` | `number` | query |  |
| `max_y` | `number` | query |  |

**Responses:**

- **200** — OK

  Returns: `string`


---

### `GET` `/api/core/slam/v1/maps/stcm`

**Get a composite map**

Composite map containing all map data The response message is a binary byte stream, which can be directly saved as an stcm file.

**Responses:**

- **200** — OK

  Returns: `string`


---

### `PUT` `/api/core/slam/v1/maps/stcm`

**Set up a composite map**

Set the map to the slamware system, and read the stcm file as the request body in binary mode. [Note] The map will not be saved persistently and will become invalid after restart.

**Responses:**

- **200** — OK

---

### `DELETE` `/api/core/slam/v1/maps`

**Clear the map**

**Responses:**

- **200** — OK

---

### `PUT` `/api/core/slam/v1/maps/origin`

**Move map origin**

Move the map origin and update it to the slamware system

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `new_origin` | `number` |
| `x` | `number` |
| `y` | `number` |

**Responses:**

- **200** — OK

---

## artifact

Semantic elements

### `GET` `/api/core/artifact/v1/lines/{usage}`

**Get virtual lines**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | tracks virtual track walls virtual wall |

**Responses:**

- **200** — OK

  **`Line`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `integer` — The id will be ignored when adding a line. |
  | `start` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `end` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `metadata` | `string` — If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata. |
    *If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata.*
    *metadata是key和value都是字符串的map*
    *描述贝塞尔曲线的metadata，control_point1和control_point2是两个控制点的坐标，再加上Line的起点和终点可以确定一个三阶贝塞尔曲线*
  | `control_point1` *(required)* | `string` |
  | `control_point2` *(required)* | `string` |


---

### `POST` `/api/core/artifact/v1/lines/{usage}`

**Add a virtual line**

id in request body is useless.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | tracks virtual track walls virtual wall |

**Request Body** (`application/json`):

> The id will be ignored when adding a line.

| Field | Type / Description |
|-------|-------------------|
| `id` | `integer` — The id will be ignored when adding a line. |
| `start` | `number` |
| `x` | `number` |
| `y` | `number` |
| `end` | `number` |
| `x` | `number` |
| `y` | `number` |
| `metadata` | `string` — If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata. |
  *If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata.*
  *metadata是key和value都是字符串的map*
  *描述贝塞尔曲线的metadata，control_point1和control_point2是两个控制点的坐标，再加上Line的起点和终点可以确定一个三阶贝塞尔曲线*
| `control_point1` *(required)* | `string` |
| `control_point2` *(required)* | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/artifact/v1/lines/{usage}`

**Modify virtual line**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | tracks virtual track walls virtual wall |

**Request Body** (`application/json`):

> The id will be ignored when adding a line.

| Field | Type / Description |
|-------|-------------------|
| `id` | `integer` — The id will be ignored when adding a line. |
| `start` | `number` |
| `x` | `number` |
| `y` | `number` |
| `end` | `number` |
| `x` | `number` |
| `y` | `number` |
| `metadata` | `string` — If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata. |
  *If it is a straight track, metadata is EmptyMetadata; if it is a Bezier curve track, metadata is BezierCurveMetadata.*
  *metadata是key和value都是字符串的map*
  *描述贝塞尔曲线的metadata，control_point1和control_point2是两个控制点的坐标，再加上Line的起点和终点可以确定一个三阶贝塞尔曲线*
| `control_point1` *(required)* | `string` |
| `control_point2` *(required)* | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/artifact/v1/lines/{usage}`

**Clear a certain type of virtual line**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | tracks virtual track walls virtual wall |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/artifact/v1/lines/{usage}/{id}`

**Delete a virtual line**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | tracks virtual track walls virtual wall |
| `id` *(required)* | `integer` | path |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/artifact/v1/rectangle-areas/{usage}`

**Get rectangular areas**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | Available values : forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area |

**Responses:**

- **200** — OK

  **`RectangleArea`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `integer` |
  | `usage` | `string` — forbidden_area robot can not enter this area elevator_area elevator area dangerous_area robot will automatically slows down after entering the area coverage_area for sweeping and disinfect maintenance_area Only the map in maintenance area can be updated sensor_disable_area Sensor data ignored when robot is in sensor disable area restricted_area For multi-robot collaborative behavior, it can limit the number of robots entering restricted area — Enum: `[ forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area ]` |
  | `area` | `number` |
  | `start` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `end` | `number` |
  | `x` | `number` |
  | `y` | `number` |
  | `half_width` | `number` |
  | `metadata` | `string` — Key/Value metadata，all the data shall be serialized in a string. — Enum: `Array [ 3 ]` |
    *Key/Value metadata，all the data shall be serialized in a string.*
    *metadata是key和value都是字符串的map*
  | `elevator_id` *(required)* | `string` — SN of elevator |
  | `elevator_sill_width` *(required)* | `string` — width of elevator sill |
  | `elevator_scheduling_point_dist` *(required)* | `string` — distance from scheduling point to elevator door |
  | `elevator_door_type` | `string` — elevator door direction, 0 front door，1 rear door，2 door on both sides — Enum: `Array [ 3 ]` |
  | `escape_distance` *(required)* | `string` — The size of escapable area. |
  | `dangerous_area_type` *(required)* | `string` — Type of dangerous, 0 Slope area，1 Narrow corridor — Enum: `Array [ 2 ]` |
  | `max_line_speed` | `string` — Max speed when robot is in this area in m/s. |
  | `sensor_type` | `string` — 0 bumper, 1 cliff, 2 sonar, 3 depthcam camera |
  | `restricted_scheduling_points` | `string` — scheduing point for robot waiting |
  | `restricted_robots_number_limit` | `string` — Number of robots allowed to enter this area simultaneously |


---

### `POST` `/api/core/artifact/v1/rectangle-areas/{usage}`

**Add a rectangular area**

Different rectangle area need different metadata，please refer to the requestBody

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | Available values : forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area |

**Request Body** (`application/json`):

> Key/Value metadata，all the data shall be serialized in a string.

| Field | Type / Description |
|-------|-------------------|
| `area` | `number` |
| `start` | `number` |
| `x` | `number` |
| `y` | `number` |
| `end` | `number` |
| `x` | `number` |
| `y` | `number` |
| `half_width` | `number` |
| `metadata` | `string` — Key/Value metadata，all the data shall be serialized in a string. — Enum: `[ 0, 1, 2 ]` |
  *Key/Value metadata，all the data shall be serialized in a string.*
  *metadata是key和value都是字符串的map*
| `elevator_id` *(required)* | `string` — SN of elevator |
| `elevator_sill_width` *(required)* | `string` — width of elevator sill |
| `elevator_scheduling_point_dist` *(required)* | `string` — distance from scheduling point to elevator door |
| `elevator_door_type` | `string` — elevator door direction, 0 front door，1 rear door，2 door on both sides — Enum: `[ 0, 1, 2 ]` |
| `escape_distance` *(required)* | `string` — The size of escapable area. |
| `dangerous_area_type` *(required)* | `string` — Type of dangerous, 0 Slope area，1 Narrow corridor — Enum: `[ 0, 1 ]` |
| `max_line_speed` | `string` — Max speed when robot is in this area in m/s. |
| `sensor_type` | `string` — 0 bumper, 1 cliff, 2 sonar, 3 depthcam camera |
| `restricted_scheduling_points` | `string` — scheduing point for robot waiting |
| `restricted_robots_number_limit` | `string` — Number of robots allowed to enter this area simultaneously |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/artifact/v1/rectangle-areas/{usage}`

**Clear a certain type of rectangular area**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | Available values : forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/artifact/v1/rectangle-areas/{usage}/{id}`

**Edit a rectangular area**

Modify location or metadata of specified rectangular area

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | Available values : forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area |
| `id` *(required)* | `integer` | path |  |

**Request Body** (`application/json`):

> Key/Value metadata，all the data shall be serialized in a string.

| Field | Type / Description |
|-------|-------------------|
| `area` | `number` |
| `start` | `number` |
| `x` | `number` |
| `y` | `number` |
| `end` | `number` |
| `x` | `number` |
| `y` | `number` |
| `half_width` | `number` |
| `metadata` | `string` — Key/Value metadata，all the data shall be serialized in a string. — Enum: `[ 0, 1, 2 ]` |
  *Key/Value metadata，all the data shall be serialized in a string.*
  *metadata是key和value都是字符串的map*
| `elevator_id` *(required)* | `string` — SN of elevator |
| `elevator_sill_width` *(required)* | `string` — width of elevator sill |
| `elevator_scheduling_point_dist` *(required)* | `string` — distance from scheduling point to elevator door |
| `elevator_door_type` | `string` — elevator door direction, 0 front door，1 rear door，2 door on both sides — Enum: `[ 0, 1, 2 ]` |
| `escape_distance` *(required)* | `string` — The size of escapable area. |
| `dangerous_area_type` *(required)* | `string` — Type of dangerous, 0 Slope area，1 Narrow corridor — Enum: `[ 0, 1 ]` |
| `max_line_speed` | `string` — Max speed when robot is in this area in m/s. |
| `sensor_type` | `string` — 0 bumper, 1 cliff, 2 sonar, 3 depthcam camera |
| `restricted_scheduling_points` | `string` — scheduing point for robot waiting |
| `restricted_robots_number_limit` | `string` — Number of robots allowed to enter this area simultaneously |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/artifact/v1/rectangle-areas/{usage}/{id}`

**Delete a rectangular area**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `usage` *(required)* | `string` | path | Available values : forbidden_area, elevator_area, dangerous_area, coverage_area, maintenance_area, sensor_disable_area, restricted_area |
| `id` *(required)* | `integer` | path |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/artifact/v1/pois`

**Get all POIs in the current map**

POI is short for Point of interest, it is use to mark a certain pose on the map with some metadata。

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `POST` `/api/core/artifact/v1/pois`

**add a POI**

The caller should generate a random UUID as the id. If Pose is not included, it means to create a POI at the current position of the robot. The display_name in metadata is used for GUI display, and type is used to distinguish POI types

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `id` *(required)* | `string` |
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `metadata` | Different API require different metadata. |
  *Different API require different metadata.*

**Responses:**

- **200** — OK

---

### `DELETE` `/api/core/artifact/v1/pois`

**Clear POI**

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `POST` `/api/core/artifact/v1/pois/:adjust`

**Optimize POI**

If POIs are added in mapping mode, they will adjust position after loop closure, calling this interface can reduce the error of position adjustment 。 【Note】Only useful after loop closure. Required minimum firmware version 4.2.4

**Responses:**

- **200** — OK

---

### `GET` `/api/core/artifact/v1/pois/{poi_id}`

**Find POI by ID**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `poi_id` *(required)* | `string($uuid)` | path |  |

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `PUT` `/api/core/artifact/v1/pois/{poi_id}`

**Modify POI**

If pose or metadata if not contained in request body, it will stay unchanged

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `poi_id` *(required)* | `string($uuid)` | path |  |

**Request Body** (`application/json`):

> Pose in 3D space

| Field | Type / Description |
|-------|-------------------|
| `pose` | `number` — Pose in 3D space |
  *Pose in 3D space*
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `yaw` | `number` |
| `pitch` | `number` |
| `roll` | `number` |
| `metadata` |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `DELETE` `/api/core/artifact/v1/pois/{poi_id}`

**Delete POI**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `poi_id` *(required)* | `string($uuid)` | path |  |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/artifact/v1/laser-landmarks`

**Get laser landmarks**

Laser landmark refers to the location of the reflector identified by the lidar. Required minimum firmware version 5.1.1

**Responses:**

- **200** — OK

  **`PoseEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` *(required)* | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |
  | `metadata` | Different API require different metadata. |
    *Different API require different metadata.*


---

### `DELETE` `/api/core/artifact/v1/laser-landmarks`

**Clear laser landmarks**

Clear all laser landmarks. Required minimum firmware version 5.1.1

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/artifact/v1/laser-landmarks`

**Set laser landmarks**

Set laser landmarks to SlamwareRequired minimum firmware version 5.1.1

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `id` *(required)* | `string` |
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `metadata` | Different API require different metadata. |
  *Different API require different metadata.*

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `GET` `/api/core/artifact/v1/laser-landmarks/:update`

**Get laser landmark update state**

Slamware is updating laser landmarks or notRequired minimum firmware version 5.1.1

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/artifact/v1/laser-landmarks/:update`

**Set laser landmark update state**

Set if enable slamware updating laser landmarksRequired minimum firmware version 5.1.1

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `POST` `/api/core/artifact/v1/laser-landmarks/:remove`

**Delete laser landmarks**

Delete some laser landmarks, request body is id array。Required minimum firmware version 5.1.1

**Request Body** (`application/json`):


Type: `integer`

**Responses:**

- **200** — OK

  Returns: `boolean`


---

## motion

Robot motion control

### `GET` `/api/core/motion/v1/action-factories`

**Get all supported actions**

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `action_name` | `string` |


---

### `GET` `/api/core/motion/v1/actions/:current`

**Get current action**

**Responses:**

- **200** — OK

  **`ActionInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `action_id` | `integer` |
  | `action_name` | `string` |
  | `stage` | `string` |
  | `state` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `status` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `result` | `integer` — 0:Success, -1: Failed, -2: Aborted — Enum: `[ 0, -1, -2 ]` |
  | `reason` | `string` |

- **404** — Action Not Found

---

### `DELETE` `/api/core/motion/v1/actions/:current`

**Abort current action**

**Responses:**

- **200** — OK

---

### `POST` `/api/core/motion/v1/actions`

**Create a new action**

**Request Body** (`application/json`):

> action_name is queried through the /core/motion/v1/action-factories interface, and the specific content of options depends on the action name

| Field | Type / Description |
|-------|-------------------|
| `action_name` | `string` — slamtec.agent.actions.MoveToAction Autonomous navigation and movement, required parameter is MoveToActionOptions. slamtec.agent.actions.SeriesMoveToAction Autonomous navigation movements containing multiple targets, required parameter is SeriesMoveToActionOptions. slamtec.agent.actions.MoveByAction Remote control movement, it should be called every 100 ms for continuous movement, required parameter is MoveByActionOptions. slamtec.agent.actions.GoHomeAction Back to charging dock, required parameter is GoHomeActionOptions. slamtec.agent.actions.RotateToAction Rotate to the specified yaw, required parameter is RotateToActionOptions. slamtec.agent.actions.RotateAction Rotate the specified angle, required parameter is RotateActionOptions. slamtec.agent.actions.MoveToTagAction Accurately docking to tag， required parameter is MoveToTagActionOptions. slamtec.agent.actions.BackOffFromTagAction Back off from tag， required parameter is BackOffFromTagActionOptions. slamtec.agent.actions.RecoverLocalizationAction Recover localization, required parameter is RecoverLocalizationActionOptions. slamtec.agent.actions.MultiFloorMoveAction Cross-floor movement，required parameter is MultiFloorMoveActionOptions. slamtec.agent.actions.MultiFloorBackHomeAction Cross-floor go back home, required parameter is GoHomeActionOptions. slamtec.agent.actions.ReturnToParkingAction Autonomous return to the standby point (POI type: PARKING): The robot autonomously returns to its parking location. It supports multi-robot scheduling and queuing (requires Lora module), required parameter is ReturnToParkingActionOptions. slamtec.agent.actions.FollowPathPointsAction Follow the existing path point movement, required parameter is FollowPathPointsActionOptions. slamtec.agent.actions.EnterElevatorAction Enter an elevator, required parameter is EnterElevatorActionOptions. *slamtec.agent.actions.LeaveElevatorAction Leave an elevator, required parameter is LeaveElevatorActionOptions. |
| `options` | `number` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `target` *(required)* | `number` |
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `move_options` | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `flags` |  |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |
| `targets` *(required)* |  |
| `move_options` | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `flags` |  |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |
  *Direction or theta is needed, the former specifies a direction, the latter specifies an angular speed to turn.*
| `direction` | `integer` — 0 Go forward 1 Back off 2 Turn right 3 Turn left — Enum: `[ 0, 1, 2, 3 ]` |
| `theta` | `number` |
| `duration` | `integer` — Duration of movement. Default for 500 ms when not specified. [Note] Move by action can not avoid obstacles, do not set a too long duration. |
| `gohome_options` | `string` — dock means go back to home dock, no_dock means go back to landing point. — Enum: `Array [ 2 ]` |
| `flags` | `string` — dock means go back to home dock, no_dock means go back to landing point. — Enum: `Array [ 2 ]` |
| `back_to_landing` | `boolean` — Go back to landing point if charging failed. |
| `charging_retry_count` | `integer` — Number of failure retry |
| `move_options` |  |
| `angle` *(required)* | `number` — Target yaw of robot. |
| `angle` *(required)* | `number` — The angular speed in rad/s, positive number indicates counterclockwise rotation and negative number indicates clockwise rotation |
  *Accurate docking to tag. target is the landing point from where robot start docking.*
| `target` *(required)* | `number` — Pose in 3D space |
  *Pose in 3D space*
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `yaw` | `number` |
| `pitch` | `number` |
| `roll` | `number` |
| `tag_type` *(required)* | `integer` — 0: Visual Tag， 1: Laser Tag, 2: Laser Reflector Tag, 3: shelf — Enum: `[ 0, 1, 2, 3 ]` |
| `target_relative_pose` | `number` — Relative pose when the robot stops in the tag coordinate. |
  *Relative pose when the robot stops in the tag coordinate.*
| `x` | `number` — Longitudinal distance from tag |
| `y` | `number` — Lateral distance from the center of tag |
| `backward_docking` | `boolean` — Backward docking or not |
| `turn_radian` | `number` — Turning radian after successful docking, If the robot is required to turn to the specified Angle after successful docking, please set this parameter. |
| `tag_ids` | `integer` — It's valid when tag_type is 0, April Tag ID |
| `reflect_tag_num` | `integer` — It's valid when tag_type is 2, number of reflectors |
| `dock_retry_count` | `integer` — number of retries if dock failed |
| `dock_allowance` | `number` — It's valid when tag_type is 3. By default, the robot's center aligns with the shelf's center during docking. dock_allowance specifies the length of the robot's body that remains outside the shelf. |
  *Back off from tag*
| `backup_mode` | `integer` — 0 Free back off 1 Back off straightly — Enum: `[ 0, 1 ]` |
| `tag_type` | `integer` — 0: Visual Tag， 1: Laser Tag, 2: Laser Reflector Tag — Enum: `[ 0, 1, 2 ]` |
| `backup_distance` | `number` — Distance of back off. |
| `backward_docking` | `boolean` — if backward docking is true, robot actually move forward when BackOffFromTagAction is invoked. |
  *If area is none or empty, it global relocalization, otherwise it is local relocalization*
| `area` | `number` |
| `x` | `number` |
| `y` | `number` |
| `width` | `number` |
| `height` | `number` |
| `relocalization_options` | `integer` — timeout of this action — Enum: `Array [ 2 ]` |
| `max_recover_time` | `integer` — timeout of this action |
| `recover_movement_type` | `string` — RotateOnly Rotate before start recover localization NoMove Keep stationary — Enum: `Array [ 2 ]` |
| `target` *(required)* |  |
| `move_options` | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `flags` |  |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |
  *Options for ReturnToParkingAction*
| `target` | `string` — Optional parameter, The robot will autonomously chooses an idle parking POI if this parameter is empty, otherwise it can only return to the specified POI. |
  *Optional parameter, The robot will autonomously chooses an idle parking POI if this parameter is empty, otherwise it can only return to the specified POI.*
| `poi_name` | `string` |
| `move_options` | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `flags` |  |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |
| `path_points` *(required)* |  |
| `move_options` | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `Array [ 3 ]` |
| `flags` |  |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |
| `elevator_id` *(required)* | `string` |
| `enter_elevator_options` | `string` — front_door Enter elevator from front door rear_door Enter elevator from rear door |
| `elevator_door_flag` | `string` — front_door Enter elevator from front door rear_door Enter elevator from rear door |
| `elevator_stopping_yaw` | `string` — face_to_front_door make robot toward the front door face_to_rear_door make robot toward the rear door |
| `timeout_in_ms` | `number` — Total time out of entering elevator |
| `use_conservative_mode` | `boolean` — true means conservative strategy and head to the center of the elevator. The false means robot will try to move into the inside |
| `elevator_id` *(required)* | `string` |
| `target` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |
| `leave_elevator_options` | `string` — front_door Leave elevator from front door rear_door Leave elevator from rear door |
| `elevator_door_flag` | `string` — front_door Leave elevator from front door rear_door Leave elevator from rear door |
| `timeout_in_ms` | `number` — Total time out of leaving elevator |
| `arrive_door_timeout_in_ms` | `number` — Timeout of moving to elevator sill |
| `search_path_timeout_in_ms` | `number` — Timeout of searching path when leaving elevator |
| `on_elevator_door_timeout_in_ms` | `number` — Timeout when robot blocked on elevator sill |
| `if_need_reach_milestone` | `boolean` — true means go to target after leaving elvator，false means go to scheduling point |
| `move_options` |  |

**Responses:**

- **200** — OK

  **`ActionInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `action_id` | `integer` |
  | `action_name` | `string` |
  | `stage` | `string` |
  | `state` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `status` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `result` | `integer` — 0:Success, -1: Failed, -2: Aborted — Enum: `[ 0, -1, -2 ]` |
  | `reason` | `string` |

- **400** — Can not create action

---

### `GET` `/api/core/motion/v1/actions/{action_id}`

**Query Action status**

The status of the last 20 actions can be queried. The state.status is 4, which means the action has ended. At this time, the result is used to determine whether the action is successful or not.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `action_id` *(required)* | `integer` | path |  |

**Responses:**

- **200** — OK

  **`ActionInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `action_id` | `integer` |
  | `action_name` | `string` |
  | `stage` | `string` |
  | `state` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `status` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `result` | `integer` — 0:Success, -1: Failed, -2: Aborted — Enum: `[ 0, -1, -2 ]` |
  | `reason` | `string` |
  | `action_id` | `integer` |
  | `state` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `status` | `integer` — 0:NewBorn, 1:Working, 3:Paused, 4:Done — Enum: `[ 0, 1, 3, 4 ]` |
  | `result` | `integer` — 0:Success, -1: Failed, -2: Aborted — Enum: `[ 0, -1, -2 ]` |
  | `reason` | `string` |

- **404** — Not Found

---

### `GET` `/api/core/motion/v1/path`

**Get remaining path**

get remaining path points of current action

**Responses:**

- **200** — OK

  **`PathPoints`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `path_points` | `number` — Each element in the path points is an array containing two float points, which means x and y coordinate values. |


---

### `GET` `/api/core/motion/v1/milestones`

**Get remaining milestones**

get remaining milestones of current action

**Responses:**

- **200** — OK

  **`PathPoints`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `path_points` | `number` — Each element in the path points is an array containing two float points, which means x and y coordinate values. |


---

### `GET` `/api/core/motion/v1/speed`

**Get speed**

Get current speed

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `vx` | `number` — linear speed of x axis, unit: m/s |
  | `vy` | `number` — linear speed of y axis, unit: m/s |
  | `omega` | `number` — angular speed, unit: rad/s |


---

### `GET` `/api/core/motion/v1/time`

**Get remaining time**

Get estimated remaining time to reach the target

**Responses:**

- **200** — OK

  Returns: `number`


---

### `POST` `/api/core/motion/v1/:search_path`

**Search path**

Search for the optimal path form robot to the target

**Request Body** (`application/json`):

> timeout for searching path

| Field | Type / Description |
|-------|-------------------|
| `target` | `number` |
| `x` | `number` |
| `y` | `number` |
| `timeout` | `integer` — timeout for searching path |

**Responses:**

- **200** — OK

  **`PathPoints`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `path_points` | `number` — Each element in the path points is an array containing two float points, which means x and y coordinate values. |


---

### `GET` `/api/core/motion/v1/strategies`

**Get all motion strategies**

Motion strategy is a series of slamware parameters that involve various aspects such as motion speed, obstacle avoidance behavior. Different strategies can be applied to different scenarios. In general, the default strategy is sufficient for most cases. Required minimum firmware version 4.2.4

**Responses:**

- **200** — OK

  Returns: `string` — Enum: `[ default, depot, inventory, delivery, low_speed ]`


---

### `GET` `/api/core/motion/v1/strategies/:current`

**Get current strategy**

**Responses:**

- **200** — OK

  Returns: `string`


---

### `PUT` `/api/core/motion/v1/strategies/:current`

**Set current strategy**

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `strategy` | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

## firmware

Firmware upgrade

### `GET` `/api/core/firmware/v1/newversion`

**Query new firmware**

Return the available new firmware information

**Responses:**

- **200** — OK

  **`FirmwareInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `manufacturer` | `string` |
  | `model` | `string` |
  | `firmware` | `string` |
  | `firmware_id` | `string` |


---

### `GET` `/api/core/firmware/v1/autoupdate/:enable`

**Auto update is enabled or not**

Required minimum firmware version 4.2.3

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/core/firmware/v1/autoupdate/:enable`

**Enable / Disable auto udpate**

Required minimum firmware version 4.2.3

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

---

### `POST` `/api/core/firmware/v1/autoupdate/:start`

**Start firmware upgrade from cloud**

Required minimum firmware version 4.2.3

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `POST` `/api/core/firmware/v1/update/:start`

**Upload firmware and start upgrade**

Upload firmware file as request body to robot, and start firmware upgrade.Required minimum firmware version 4.2.3

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `result` | `boolean` |
  | `msg` | `string` |
  | `data` |  |


---

### `GET` `/api/core/firmware/v1/progress`

**Get firmware upgrade progress**

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `current_step` | `integer` — 0:Preparing, 1:PrepareFinished, 2:Downloding, 3:DownloadFinished, 4:Updating, 5:UpdateFinished — Enum: `[ 0, 1, 2, 3, 4, 5 ]` |
  | `current_step_name` | `string` |
  | `current_step_progress` | `integer` — The progress of the current step, 0~100 |
  | `toptalSteps` | `integer` |
  | `status` | `integer` — 0:Success, 1:Error, 2:Init, 3:Upgrade — Enum: `[ 0, 1, 2, 3 ]` |
  | `error_code` | `integer` |


---

## statistics

Statistics

### `GET` `/api/core/statistics/v1/odometry`

**Get odometry**

Get total movement distance of the robot, unit m/s

**Responses:**

- **200** — OK

  Returns: `number`


---

### `GET` `/api/core/statistics/v1/runtime`

**Get run time**

Get total run time of the robot in seconds

**Responses:**

- **200** — OK

  Returns: `number`


---

## sensors

Sensor control

### `PUT` `/api/core/sensors/v1/depth/:enable`

**Enable / Disable depth camera data**

The return value true means the operation was successful

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` | `boolean` |

**Responses:**

- **200** — OK

---

### `GET` `/api/core/sensors/v1/masks`

**Get all disabled sensors mask data**

Get all disabled sensors mask data

**Responses:**

- **200** — OK

  **`DisabledSensorMaskData`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `integer` |
  | `isAlways` | `boolean` |


---

### `PUT` `/api/core/sensors/v1/masks`

**Enable / Disable base sensors**

set sensors mask.

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `id` | `number` |
| `isAlways` | `boolean` |
| `isEnabled` | `boolean` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

## application

Android application management(ARM only)

### `GET` `/api/core/application/v1/apps`

**Get all custom installed apps**

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `name` *(required)* | `string` |
  | `version` | `string` |


---

### `POST` `/api/core/application/v1/apps`

**Install APP**

**Responses:**

- **200** — OK
- **500** — Failed to install application

---

### `DELETE` `/api/core/application/v1/apps/{app_name}`

**Uninstall APP**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `app_name` *(required)* | `string` | path |  |

**Responses:**

- **200** — OK

---

## platform

Robot chassis and platform

### `GET` `/api/platform/v1/timestamp`

**Get system timestamp**

Get the number of milliseconds since the system was started, returning an integer in string format. Required minimum firmware version 4.2.4

**Responses:**

- **200** — OK

  Returns: `string`


---

### `GET` `/api/platform/v1/events`

**Get events**

Get events occurring on the robot.

**Responses:**

- **200** — OK

  **`RobotEvent`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Type will expand new definition in different scenarios. APP only needs to process the events that you care about. GeneralEventType for general events, ElevatorEventType for the elevator related events, DeliveryEventType for delviery related events.*
  | `type` | `string` — DEVICE_ERROR An Error or Fatal health information has occurred PATH_OCCUPIED Path occupied by obstacle ROBOT_BLOCKED Blocked at same place for long time(3 minutes). RESET_MAP_TO_DOCK Robot is pushed to homd dock and map reset START_CHARGING Start charging STOP_CHARGING Stop charging ON_DOCK Robot go back to home dock successfully OFF_DOCK Robot leave home dock UPGRADE Start firmware upgrading POWER_OFF Power off PASS_THE_NARROW_CORRIDOR Robot is passing narrow corridor MAP_LOOP_CLOSURE Loop closure occurs SET_MAP_DONE Set map operation completed MOVE_TO_LANDING_POINT_FAILED Failed to go back to home dock SEARCH_DOCK_FAILED Failed to find home dock CHARGING_BASE_FAILED Failed to charging base SYNC_MAP_FROM_CLOUD Download map from cloud DOCK_ID_NOT_FOUND Home dock with specified id not found BRAKE_RELEASED Brake released BUMPER_TRIGGERED Bumper sensor is triggered CURRENT_POSE_OCCUPIED Current robot pose occupied by obstacle CLIFF_DETECTED Cliff detected — Enum: `[ DEVICE_ERROR, PATH_OCCUPIED, ROBOT_BLOCKED, RESET_MAP_TO_DOCK, START_CHARGING, STOP_CHARGING, ON_DOCK, OFF_DOCK, UPGRADE, POWER_OFF, PASS_THE_NARROW_CORRIDOR, MAP_LOOP_CLOSURE, SET_MAP_DONE, MOVE_TO_LANDING_POINT_FAILED, SEARCH_DOCK_FAILED, CHARGING_BASE_FAILED, SYNC_MAP_FROM_CLOUD, DOCK_ID_NOT_FOUND, BRAKE_RELEASED, BUMPER_TRIGGERED, CURRENT_POSE_OCCUPIED, CLIFF_DETECTED ]` |
  | `timestamp` | `string` — Milliseconds since the system started |


---

## multi-floor

Map management and across floor movement

### `GET` `/api/multi-floor/status`

**Get map loading status**

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `is_in_mapping_mode` | `boolean` — in mapping mode or not |
  | `map_load_status` | `string` — NOT_LOADED there is no local map file. LOADING loading map. LOADED map loaded successfully。 LOADING_SKIPPED Skip the map loading step, when service is restart, it will be such status NEED_LOAD Receive sync map command from cloud when executing task. ERROR Error status — Enum: `[ NOT_LOADED, LOADING, LOADED, LOADING_SKIPPED, NEED_LOAD, ERROR ]` |
  | `is_managed_by_cloud` | `boolean` — Robot is managed by cloud or not |


---

### `GET` `/api/multi-floor/map/v1/floors`

**Get all floor information**

**Responses:**

- **200** — OK

  **`FloorInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `building` | `string` |
  | `floor` | `string` |
  | `order` | `integer` — If floors are displayed in UI, they should be ordered occording to this field. |
  | `is_default_floor` | `boolean` — Whether it is default floor. |


---

### `GET` `/api/multi-floor/map/v1/floors/:current`

**Get the current floor information of the robot**

**Responses:**

- **200** — OK

  **`CurrentFloorInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `building` | `string` |
  | `floor` | `string` |
  | `elevator` | `string` — If the field is not empty, it means that the robot is still in the elevator. |
  | `map_id` | `string` — The map ID of the current floor. |


---

### `PUT` `/api/multi-floor/map/v1/floors/:current`

**Set the floor information of the robot**

Under normal circumstances, the robot should automatically switch floors during the elevator ride. This interface is only for special situations (such as manual handling robots).

**Request Body** (`application/json`):

> Pose in 2D space

| Field | Type / Description |
|-------|-------------------|
| `building` | `string` |
| `floor` *(required)* | `string` |
| `pose` | `number` — Pose in 2D space |
  *Pose in 2D space*
| `x` | `number` |
| `y` | `number` |
| `yaw` | `number` |

**Responses:**

- **200** — OK

---

### `GET` `/api/multi-floor/map/v1/pois`

**Get POI information**

Get all POIs of the specified floor, you will get all pois of the map if you do not set the specifed floor

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `floor` | `string` | query | Floor name |
| `building` | `string` | query | Building name |

**Responses:**

- **200** — OK

  **`MultiFloorPoiInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `poi_name` | `string` |
  | `type` | `string` — Enum: `[ ROOM, REFILL, RECEPTION, TABLE, PARKING, RECYCLE, DISINFECT ]` |
  | `floor` | `string` |
  | `building` | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |

- **400** — Invalid floor or building

---

### `POST` `/api/multi-floor/map/v1/pois/:search_nearby`

**Look for the nearest POI**

Find the nearest POI. There are three special name, ON_DOCK robot on homedock, IN_ELEVATOR robot is in the elevator, UNKNOWN means no POI in the map. The other values represent the name of the POI added in the map.

**Responses:**

- **200** — OK

  **`NearbyPoiInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *The relative pose of POI to robot is in robot coordinate. The front of the robot is the X axis, and the left is the positive direction of the Y axis.*
  | `id` | `string` |
  | `name` *(required)* | `string` — ON_DOCK robot is on home dock IN_ELEVATOR robot is in elevator UNKNOWN No POI — Enum: `[ ON_DOCK, IN_ELEVATOR, UNKNOWN ]` |
  | `relative_pose` | `number` |
  | `x` | `number` |
  | `y` | `number` |


---

### `POST` `/api/multi-floor/map/v1/pois/:dispatch`

**Query the optimal traversal order of POI**

Given several POI names, return the adjusted POI sequence, making the shortest total path that the robot passes through these POI in turn.【Note】The cost time increases exponentially, do not pass a large number of POI. Required minimum firmware version 4.5.0

**Request Body** (`application/json`):


Type: `string`

**Responses:**

- **200** — OK

  Returns: `string`


---

### `GET` `/api/multi-floor/map/v1/homedocks`

**Get charging docks**

Get charging docks of designated floor, return all docks if no query parameter is assigned.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `floor` | `string` | query | floor name |
| `building` | `string` | query | building name |

**Responses:**

- **200** — OK

  **`MultiFloorDockInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `dock_name` | `string` |
  | `floor` | `string` |
  | `building` | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |


---

### `GET` `/api/multi-floor/map/v1/homedocks/:current`

**Get current charging dock**

Get charging dock information currently bound by robot. Result will be false if there is no bound dock.

**Responses:**

- **200** — OK

  **`MultiFloorDockInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `result` | `boolean` |
  | `msg` | `string` |
  | `data` | `string` — Pose in 2D space |
  | `id` | `string` |
  | `dock_name` | `string` |
  | `floor` | `string` |
  | `building` | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |


---

### `PUT` `/api/multi-floor/map/v1/homedocks/:current`

**Bind charging dock**

【Note】If the bound charging dock is not on the starting floor, you need to push the robot to the charging dock, and then call the API. At this case, the starting floor will be modified synchronously and the map will be reset.

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `dock_id` | `string` |

**Responses:**

- **200** — OK

---

### `POST` `/api/multi-floor/map/v1/homedocks/:search_nearby`

**Find the charging dock closest to the robot**

Find the charging dock closest to the robot.

**Responses:**

- **200** — OK

  **`MultiFloorDockInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `dock_name` | `string` |
  | `floor` | `string` |
  | `building` | `string` |
  | `pose` | `number` — Pose in 2D space |
    *Pose in 2D space*
  | `x` | `number` |
  | `y` | `number` |
  | `yaw` | `number` |


---

### `POST` `/api/multi-floor/map/v1/stcm`

**Upload the map to the robot**

The uploaded map will be saved persistently in the file system, but will not be loaded into Slamware. [Attention] When the robot is managed by the cloud, the map downloaded from the cloud will overwrite the local map.

**Responses:**

- **200** — OK

---

### `DELETE` `/api/multi-floor/map/v1/stcm`

**Delete saved map**

The current map in memory is not emptied, but the cached map in the file system is deleted

**Responses:**

- **204** — OK

---

### `POST` `/api/multi-floor/map/v1/stcm/:save`

**Save current map to disk**

Read the map from slamware and save it to a file. 【Note】this operation is prohibited in multi floor environment, otherwise the maps of other floors will lost.

**Responses:**

- **200** — OK

---

### `POST` `/api/multi-floor/map/v1/stcm/:reload`

**Reload the map**

Reload the map, try to download from the cloud, robot will download map from current file system if robot is not in the cloud environment. pose is an optional parameters. The robot pose is set to the front of charging dock by default 【Note】the map will be loaded automatically when the system is started. This API usually needs to be called only when the map changes during the deployment phase.

**Request Body** (`application/json`):

> Pose in 3D space

| Field | Type / Description |
|-------|-------------------|
| `pose` | `number` — Pose in 3D space |
  *Pose in 3D space*
| `x` | `number` |
| `y` | `number` |
| `z` | `number` |
| `yaw` | `number` |
| `pitch` | `number` |
| `roll` | `number` |

**Responses:**

- **200** — OK

---

### `POST` `/api/multi-floor/map/v1/stcm/:sync`

**Synchronize map**

Save current map and reload. Equivalent to a combination of save and reload.

**Responses:**

- **200** — OK

---

### `POST` `/api/multi-floor/map/v1/scene/unbind`

**Unbind robot with scene**

Unbind the robot from the cloud scenario and delete the local map, to be called when the robot needs to switch to a new deployment scenario.Required minimum firmware version 6.2.0

**Request Body** (`application/json`):

> Whether to retain the robot's local map, true means retain.

| Field | Type / Description |
|-------|-------------------|
| `keep_local_map` | `boolean` — Whether to retain the robot's local map, true means retain. |

**Responses:**

- **200** — OK

---

### `POST` `/api/multi-floor/map/v1/search_path_points`

**Search path points via virtual tracks**

In the graph formed by the virtual tracks, search for feasible paths from the starting point to the destination.

**Request Body** (`application/json`):

> If you do not specify the starting point, take the robot location as the starting point

| Field | Type / Description |
|-------|-------------------|
  *If you do not specify the starting point, take the robot location as the starting point*
| `building` | `string` — building name |
| `floor` | `string` — floor name |
| `start_point` | `number` |
| `x` | `number` |
| `y` | `number` |
| `end_point` *(required)* | `number` |
| `x` | `number` |
| `y` | `number` |
| `with_direction` | `boolean` — Is the graph directional?(default is false) |

**Responses:**

- **200** — OK

  **`PathPoints`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `path_points` | `number` — Each element in the path points is an array containing two float points, which means x and y coordinate values. |


---

### `PUT` `/api/multi-floor/localization/v1/pose`

**Set robot pose to POI**

Set the robot pose is to a specified POI, it is generally used for the recovery operation after an exception.Required minimum firmware version 4.5.3

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `poi_name` | `string` |

**Responses:**

- **200** — OK

  Returns: `boolean`


---

### `PUT` `/api/multi-floor/localization/v1/homedock`

**Reset pose by homedock**

Set the robot's position to the specified homedock, usually used to restore localization after a localization loss.Required minimum firmware version 6.2.0

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `dock_id` | `string` |

**Responses:**

- **200** — OK

---

### `GET` `/api/multi-floor/map/v1/elevators`

**Get all elevators info**

Get the elements within the elevator area, including elevator IDs and waiting points.

**Responses:**

- **200** — OK

  **`ElevatorInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Elevator area information*
  | `door_type` *(required)* | `string` — Enum: `[ front_door, rear_door, double_doors ]` |
  | `elevator_id` *(required)* | `string` |
  | `front_scheduling_poses` | `number` — Pose in 3D space |
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |
  | `rear_scheduling_poses` | `number` — Pose in 3D space |
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |


---

### `GET` `/api/multi-floor/map/v1/elevators/{elevator_id}`

**Get the specific elevator information**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `elevator_id` *(required)* | `string` | path |  |

**Responses:**

- **200** — OK

  **`ElevatorInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *Elevator area information*
  | `door_type` *(required)* | `string` — Enum: `[ front_door, rear_door, double_doors ]` |
  | `elevator_id` *(required)* | `string` |
  | `front_scheduling_poses` | `number` — Pose in 3D space |
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |
  | `rear_scheduling_poses` | `number` — Pose in 3D space |
    *Pose in 3D space*
  | `x` | `number` |
  | `y` | `number` |
  | `z` | `number` |
  | `yaw` | `number` |
  | `pitch` | `number` |
  | `roll` | `number` |

- **400** — Unknown Elevator ID

---

### `GET` `/api/multi-floor/map/v1/elevators/{elevator_id}/pose_relation`

**Get the spatial relationship between the robot and the elevator**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `elevator_id` *(required)* | `string` | path |  |

**Responses:**

- **200** — OK

  Returns: `string` — Enum: `[ in_elevator, close_to_elevator_sill, out_of_elevator ]`

- **400** — Unknown Elevator ID

---

## industry

For industrial chassis

### `POST` `/api/industry/v1/tasks/templates`

**Create a task template**

Create a caller task template.

**Request Body** (`application/json`):

> Task template ID

| Field | Type / Description |
|-------|-------------------|
| `key` | `string` — Task template ID |
| `name` | `string` — Task template name |
| `action_list` | `string` — list of task point and operation |
| `display_name` | `string` — isplay_name of target POI |
| `action` | `string` — The operation the robot needs to perform at the target point |
| `wait_time` | `integer` — The waiting time after the robot completes the operation at the target point, in seconds. |

**Responses:**

- **200** — OK

  **`TaskTemplate`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `task_template_key` | `string` — Task template ID |
  | `task_template_type` | `integer` — Task template type |
  | `name` | `string` — Task template name |
  | `scene_id` | `string` — The scene ID associated with the current task template. |
  | `device_id` | `string` — The device ID that created the task template. |
  | `action_list` | `string` — list of task point and operation |
  | `display_name` | `string` — isplay_name of target POI |
  | `action` | `string` — The operation the robot needs to perform at the target point |
  | `wait_time` | `integer` — The waiting time after the robot completes the operation at the target point, in seconds. |


---

### `GET` `/api/industry/v1/tasks/templates`

**Get task templates**

Retrieve all task templates in the current device's associated scenario.

**Responses:**

- **200** — OK

  **`TaskTemplate`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `task_template_key` | `string` — Task template ID |
  | `task_template_type` | `integer` — Task template type |
  | `name` | `string` — Task template name |
  | `scene_id` | `string` — The scene ID associated with the current task template. |
  | `device_id` | `string` — The device ID that created the task template. |
  | `action_list` | `string` — list of task point and operation |
  | `display_name` | `string` — isplay_name of target POI |
  | `action` | `string` — The operation the robot needs to perform at the target point |
  | `wait_time` | `integer` — The waiting time after the robot completes the operation at the target point, in seconds. |


---

### `DELETE` `/api/industry/v1/tasks/templates/{key_id}`

**Delete a task template**

Delete a task template.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `key_id` *(required)* | `string($uuid)` | path |  |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *If result is true, it indicates that the operation was successful. If result is false, msg provides the reason for the failure.*
  | `result` | `boolean` |
  | `msg` | `string` |
  | `data` |  |


---

### `GET` `/api/industry/v1/tasks`

**Query task information**

By default, it returns all types of tasks in the "ready" and "running" states. When the status is set to "all," it queries the most recent tasks, including those successfully completed and those that failed.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `type` | `string` | query | Available values : carry_calling_by_template, carry_calling, industry |
| `status` | `string` | query | Available values : ready, running, succeeded, failed, all |

**Responses:**

- **200** — OK

  **`IndustryTask`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `task` | `string` — Enum: `[ CARRY_CALLING, CARRY_CALLING_BY_TEMPLATE, INDUSTRY ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `[ CARRY_CALLING, CARRY_CALLING_BY_TEMPLATE, INDUSTRY ]` |
  | `order_id` | `string` |
  | `template_key` | `string` |
  | `start_time` | `string` |
  | `task_targets` |  |
  | `message` |  |
  | `status` | `string` — Enum: `[ READY, RUNNING, SUCCEEDED, FAILED, CANCELING, CANCELED ]` |
  | `result` | `string` — Stage of task — Enum: `[ GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, GOING_TO_TASK_POINT, ARRIVED_AT_TASK_POINT, GOING_TO_TARGET_POINT, ARRIVED_AT_TARGET_POINT, WAIT_OPERATION ]` |
  | `stage` | `string` — Stage of task — Enum: `[ GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, GOING_TO_TASK_POINT, ARRIVED_AT_TASK_POINT, GOING_TO_TARGET_POINT, ARRIVED_AT_TARGET_POINT, WAIT_OPERATION ]` |
  | `reason` | `string` |
  | `timestamp` | `string` |


---

### `POST` `/api/industry/v1/tasks/events`

**Post a task event**

When the APP executes a caller task, it pushes task events through this interface and updates the task status.

**Request Body** (`application/json`):

> task status

| Field | Type / Description |
|-------|-------------------|
| `task_id` | `string` |
| `status` | `string` — task status — Enum: `[ RUNNING, SUCCEEDED, FAILED ]` |

**Responses:**

- **200** — OK

---

## delivery

Delivery service (specific models are required, not supported on chassis)

### `GET` `/api/delivery/v1/admin/password`

**Get operation password**

Expires indicates the password expiration time. If this field is not included, it means that the password is permanently valid. Enable indicates you need password when you use the robot

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `delivery_admin_password` | `boolean` |
  | `enable` *(required)* | `boolean` |
  | `password` | `string` |
  | `expires` | `string` |


---

### `PUT` `/api/delivery/v1/admin/password`

**Set operation password**

If enable is false, the password is not needed when you use the robot

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable` *(required)* | `boolean` |
| `password` | `string` |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *When result is true, the operation is successful. If result is false, reason is the reason for failure*
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `GET` `/api/delivery/v1/admin/mode`

**Set work mode**

**Responses:**

- **200** — OK

  **`DeliveryWorkMode`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `work_mode` | `string` — DISPATCH 派送模式，本地应当禁止用户创建任务，只响应云端的呼叫任务 RECYCLE 回盘模式，本地除了回盘禁止创建其他任务，可响应云端的呼叫回盘任务 MANUAL 手动操作模式，本地人工创单进行配送或回盘 — Enum: `[ DISPATCH, RECYCLE, MANUAL ]` |


---

### `PUT` `/api/delivery/v1/admin/mode`

**Get work mode**

**Request Body** (`application/json`):

> DISPATCH 派送模式，本地应当禁止用户创建任务，只响应云端的呼叫任务 RECYCLE 回盘模式，本地除了回盘禁止创建其他任务，可响应云端的呼叫回盘任务 MANUAL 手动操作模式，本地人工创单进行配送或回盘

| Field | Type / Description |
|-------|-------------------|
| `work_mode` | `string` — DISPATCH 派送模式，本地应当禁止用户创建任务，只响应云端的呼叫任务 RECYCLE 回盘模式，本地除了回盘禁止创建其他任务，可响应云端的呼叫回盘任务 MANUAL 手动操作模式，本地人工创单进行配送或回盘 — Enum: `[ DISPATCH, RECYCLE, MANUAL ]` |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *When result is true, the operation is successful. If result is false, reason is the reason for failure*
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `GET` `/api/delivery/v1/admin/language`

**Get robot language**

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `language` | `string` |


---

### `PUT` `/api/delivery/v1/admin/language`

**Set robot language**

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `language` | `string` |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *When result is true, the operation is successful. If result is false, reason is the reason for failure*
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `GET` `/api/delivery/v1/admin/working_time`

**Get working time**

**Responses:**

- **200** — OK

  **`WorkingTime`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `hours` | `string` — Both start and end time format should be HH::MM::SS |
    *Both start and end time format should be HH::MM::SS*
  | `start_time` | `string` |
  | `end_time` | `string` |
  | `restdays` | `number` — Rest days, 0 means Sunday, 1~6 means Monday to Saturday. — Enum: `[ 0, 1, 2, 3, 4, 5, 6 ]` |


---

### `PUT` `/api/delivery/v1/admin/working_time`

**Set working time**

**Request Body** (`application/json`):

> Both start and end time format should be HH::MM::SS

| Field | Type / Description |
|-------|-------------------|
| `hours` | `string` — Both start and end time format should be HH::MM::SS |
  *Both start and end time format should be HH::MM::SS*
| `start_time` | `string` |
| `end_time` | `string` |
| `restdays` | `number` — Rest days, 0 means Sunday, 1~6 means Monday to Saturday. — Enum: `[ 0, 1, 2, 3, 4, 5, 6 ]` |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *When result is true, the operation is successful. If result is false, reason is the reason for failure*
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `GET` `/api/delivery/v1/admin/move_options`

**Get move options**

**Responses:**

- **200** — OK

  **`MoveOptions`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `[ 0, 1, 2 ]` |
  | `flags` | `string` — precise make the robot more accurate to reach the target point with_yaw make the robot turn to yaw after reaching target fail_retry_count set the failure retry count when search path failed find_path_ignoring_dynamic_obstacles Ignoring dynamic obstacles when searching path, it is suitable for crowded and narrow areas — Enum: `[ precise, with_yaw, fail_retry_count, find_path_ignoring_dynamic_obstacles ]` |
  | `yaw` | `number` — The orientation of the robot after reaching the target point |
  | `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
  | `fail_retry_count` | `integer` — Number of failure retry |
  | `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |


---

### `PUT` `/api/delivery/v1/admin/move_options`

**Set move options**

Set the motion options during delivery, such as free navigation or track mode. When the request message is empty, restore the default option.

**Request Body** (`application/json`):

> 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance)

| Field | Type / Description |
|-------|-------------------|
| `mode` *(required)* | `integer` — 0 Free navigation, 1Restrict virtual track mode, 2 Virtual track priority mode(support obstacle avoidance) — Enum: `[ 0, 1, 2 ]` |
| `flags` | `string` — precise make the robot more accurate to reach the target point with_yaw make the robot turn to yaw after reaching target fail_retry_count set the failure retry count when search path failed find_path_ignoring_dynamic_obstacles Ignoring dynamic obstacles when searching path, it is suitable for crowded and narrow areas — Enum: `[ precise, with_yaw, fail_retry_count, find_path_ignoring_dynamic_obstacles ]` |
| `yaw` | `number` — The orientation of the robot after reaching the target point |
| `acceptable_precision` | `number` — When the target point is occupied, the action considered as successful when distance between the robot and the target point is less than this value. |
| `fail_retry_count` | `integer` — Number of failure retry |
| `speed_ratio` | `number` — 【Required minimum firmware version 4.5.4】Used to constraint the maximum speed of this movement, the minimum value is 0.1. |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
    *When result is true, the operation is successful. If result is false, reason is the reason for failure*
  | `result` | `boolean` |
  | `reason` | `string` |


---

### `GET` `/api/delivery/v1/admin/line_speed`

**Get move speed**

Required minimum firmware version 4.5.3

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `delivery_speed` | `number` |
  | `return_speed` | `number` |


---

### `PUT` `/api/delivery/v1/admin/line_speed`

**Set move speed**

Required minimum firmware version 4.5.3

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `delivery_speed` | `number` |
| `return_speed` | `number` |

**Responses:**

- **200** — OK

---

### `GET` `/api/delivery/v1/configurations`

**Get configuration information**

**Responses:**

- **200** — OK

  **`Cargo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `device_sn` | `string` |
  | `firmware_version` | `string` |
  | `is_manage_by_cloud` | `boolean` |
  | `enable_recovery_on_parking` | `boolean` |
  | `cargos` | `string` — Enum: `[ FRONT, BACK, TOP ]` |
  | `id` | `string` |
  | `pos` | `integer` |
  | `orientation` | `string` — Enum: `[ FRONT, BACK, TOP ]` |
  | `layer` | `integer` |
  | `type` | `string` — Enum: `[ TAKEOUT, RETAIL ]` |
  | `errors` | `string` |
  | `boxes` |  |


---

### `GET` `/api/delivery/v1/settings`

**Get delivery settings information**

**Responses:**

- **200** — OK

  **`DeliverySettings`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `delivery_settings` | `integer` — The robot will automatically shut down when the battery level is reached |
  | `low_battery_level` | `integer` — The robot will automatically shut down when the battery level is reached |
  | `level1` | `integer` — The robot will automatically shut down when the battery level is reached |
  | `level2` | `integer` — When the battery level is reached, the robot cancels all tasks and returns to the charging station |
  | `level3` | `integer` — Reserved. When scheduling machines through the cloud, once the power is reached, new tasks should be prohibited |
  | `level4` | `integer` — When the battery level is reached, the robot cannot create a takeaway delivery order |
  | `timeout_settings` | `integer` — Waiting time for user to open cargo after robot arrived at destination. |
  | `takeout_pickup_timeout` | `integer` — Waiting time for user to open cargo after robot arrived at destination. |
  | `takeout_open_door_timeout` | `integer` — Waiting time for auto close after opening the door |
  | `collect_pickup_timeout` | `integer` — Waiting time when robot arrived at reception after delivery failure |
  | `brake_released_timeout` | `integer` — Waiting time when brake released or emergency stop pressed. |
  | `food_pickup_timeout` | `integer` — Waiting time for user to pickup when food is delivered |


---

### `PUT` `/api/delivery/v1/settings/timeout`

**Set timeout of task**

**Request Body** (`application/json`):

> Maximum waiting time after the robot reaches the destination

| Field | Type / Description |
|-------|-------------------|
| `food_pickup_timeout` | `integer` — Maximum waiting time after the robot reaches the destination |

**Responses:**

- **200** — OK

---

### `GET` `/api/delivery/v1/voice_resources`

**Get voice resources**

Get voice resources information from the cloudRequired minimum firmware version 4.3.2

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `result` | `boolean` |
  | `msg` | `string` |
  | `data` | `string` |
  | `version` | `string` |
  | `content` | `string` |
  | `interval_count` | `integer` |
  | `play_type` | `integer` |
  | `repeat_count` | `integer` |


---

### `GET` `/api/delivery/v1/cargos`

**Get all cargo information**

Only robot with cargos supports the cargos series interface

**Responses:**

- **200** — OK

  **`Cargo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `pos` | `integer` |
  | `orientation` | `string` — Enum: `[ FRONT, BACK, TOP ]` |
  | `layer` | `integer` |
  | `type` | `string` — Enum: `[ TAKEOUT, RETAIL ]` |
  | `errors` | `string` |
  | `boxes` | `integer` — Enum: `Array [ 5 ]` |
  | `id` | `integer` |
  | `door_status` | `string` — Enum: `Array [ 5 ]` |
  | `lock_status` | `string` — Enum: `Array [ 2 ]` |
  | `stock_status` | `string` — Enum: `Array [ 3 ]` |
  | `status` | `string` — Enum: `Array [ 3 ]` |
  | `errors` |  |


---

### `GET` `/api/delivery/v1/cargos/{cargo_id}/boxes`

**Get all box information of a cargo**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `cargo_id` *(required)* | `string($uuid)` | path |  |

**Responses:**

- **200** — OK

  **`Box`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `integer` |
  | `door_status` | `string` — Enum: `[ OPEN, OPENING, CLOSING, CLOSED, SEMIOPEN ]` |
  | `lock_status` | `string` — Enum: `[ LOCKED, UNLOCKED ]` |
  | `stock_status` | `string` — Enum: `[ EMPTY, SEMIFULL, FULL ]` |
  | `status` | `string` — Enum: `[ EMPTY, NOT_EMPTY, ERROR ]` |
  | `errors` | `string` |

- **404** — Cargo not found

---

### `GET` `/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}`

**Get box information**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `cargo_id` *(required)* | `string($uuid)` | path |  |
| `box_id` *(required)* | `integer` | path | Example : 0 |

**Responses:**

- **200** — OK

  **`Box`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `integer` |
  | `door_status` | `string` — Enum: `[ OPEN, OPENING, CLOSING, CLOSED, SEMIOPEN ]` |
  | `lock_status` | `string` — Enum: `[ LOCKED, UNLOCKED ]` |
  | `stock_status` | `string` — Enum: `[ EMPTY, SEMIFULL, FULL ]` |
  | `status` | `string` — Enum: `[ EMPTY, NOT_EMPTY, ERROR ]` |
  | `errors` | `string` |

- **400** — Invalid BoxId
- **404** — Box not found

---

### `PUT` `/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/{op}`

**operateBox**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `cargo_id` *(required)* | `string($uuid)` | path |  |
| `box_id` *(required)* | `integer` | path | Example : 0 |
| `op` *(required)* | `string` | path | Available values : :open, :close |

**Responses:**

- **200** — OK
- **400** — Invalid Operation

---

### `GET` `/api/delivery/v1/cargos/{cargo_id}/boxes/{box_id}/operation_result`

**Query box operation results**

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `cargo_id` *(required)* | `string($uuid)` | path |  |
| `box_id` *(required)* | `integer` | path | Example : 0 |

**Responses:**

- **200** — OK

  **`Box`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `type` | `string` — Type of the last cargo operation — Enum: `[ OPEN, CLOSE ]` |
  | `stage` | `string` — Operation progress — Enum: `[ IN_PROGRESS, DONE, FAILED ]` |
  | `reason` | `string` |
  | `cargo_id` | `string` |
  | `box` | `integer` — Enum: `[ OPEN, OPENING, CLOSING, CLOSED, SEMIOPEN ]` |
  | `id` | `integer` |
  | `door_status` | `string` — Enum: `[ OPEN, OPENING, CLOSING, CLOSED, SEMIOPEN ]` |
  | `lock_status` | `string` — Enum: `[ LOCKED, UNLOCKED ]` |
  | `stock_status` | `string` — Enum: `[ EMPTY, SEMIFULL, FULL ]` |
  | `status` | `string` — Enum: `[ EMPTY, NOT_EMPTY, ERROR ]` |
  | `errors` | `string` |


---

### `GET` `/api/delivery/v1/cargos/assigned`

**Get occupied cargos**

**Responses:**

- **200** — OK

  **`AssignedCargoEntry`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `target` | `string` |
  | `order_id` | `string` |
  | `cargo_id` | `string` |
  | `boxes` | `string` |


---

### `GET` `/api/delivery/v1/tasks`

**Query task information**

By default, it returns all types of tasks in the "ready" and "running" states. When the status is set to "all," it queries the most recent tasks, including those successfully completed and those that failed.

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `type` | `string` | query | Available values : takeout, retail, collect, refill, food_delivery, recycle, return, call |
| `status` | `string` | query | Available values : ready, running, succeeded, failed, all |

**Responses:**

- **200** — OK

  **`DeliveryTask`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `id` | `string` |
  | `task` | `string` — Enum: `[ REFILL, COLLECT, TAKEOUT, RETAIL, STATION_TAKEOUT, GUIDE, FOOD_DELIVERY, RETURN, RECYCLE ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `[ REFILL, COLLECT, TAKEOUT, RETAIL, STATION_TAKEOUT, GUIDE, FOOD_DELIVERY, RETURN, RECYCLE ]` |
  | `req_id` | `string` |
  | `order_id` | `string` |
  | `no_pickup_wait` | `boolean` |
  | `cargos` |  |
  | `failed_tasks` |  |
  | `station_id` | `string` |
  | `station_cargos` |  |
  | `message` |  |
  | `status` | `string` — Enum: `[ READY, RUNNING, SUCCEEDED, FAILED, CANCELING, CANCELED ]` |
  | `result` | `string` — Task execution stage — Enum: `[ GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, ON_DELIVERING, ARRIVED_AT_DELIVERY_POSE, WAIT_REFILL_TIMEOUT, USER_OPERATE_ROBOT, GOING_TO_CABINET, ROBOT_OPERATE_CABINET, GOING_TO_TASK_POINT, ARRIVED_AT_TASK_POINT ]` |
  | `stage` | `string` — Task execution stage — Enum: `[ GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, ON_DELIVERING, ARRIVED_AT_DELIVERY_POSE, WAIT_REFILL_TIMEOUT, USER_OPERATE_ROBOT, GOING_TO_CABINET, ROBOT_OPERATE_CABINET, GOING_TO_TASK_POINT, ARRIVED_AT_TASK_POINT ]` |
  | `reason` | `string` |
  | `pickup_result` |  |
  | `timestamp` | `string` |


---

### `POST` `/api/delivery/v1/tasks`

**Create a task**

**Request Body** (`application/json`):

> Destination of the task

| Field | Type / Description |
|-------|-------------------|
| `location` | `string` — Destination of the task |
| `poi_name` | `string` — Destination of the task |
| `task_points` | `string` — Task point, indicating the point that needs to be stop during the execution of the task. It can be empty if not needed |
| `type` | `string` — TAKEOUT 外卖配送任务（仅限有货仓的机型） GUIDE 引领任务，将人带到指定目的地 FOOD_DELIVERY 送餐任务 RETURN 快速返航，回到取餐点 RECYCLE 回收餐盘 TAKEOUT_DISTRIBUTE 外卖分发，打开所有舱门由用户自主取物 — Enum: `[ TAKEOUT, GUIDE, FOOD_DELIVERY, RETURN, RECYCLE, TAKEOUT_DISTRIBUTE ]` |
| `cargos` | `string` |
| `cargo_id` | `string` |
| `boxes` | `string` |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `result` *(required)* | `boolean` |
  | `order_id` | `string` |
  | `errors` |  |


---

### `DELETE` `/api/delivery/v1/tasks`

**Cancel all tasks**

**Responses:**

- **200** — OK

---

### `POST` `/api/delivery/v1/tasks/:batch`

**Create multiple tasks**

Create multiple tasks. Required minimum firmware version 4.3.0

**Request Body** (`application/json`):

> Destination of the task

| Field | Type / Description |
|-------|-------------------|
| `location` | `string` — Destination of the task |
| `poi_name` | `string` — Destination of the task |
| `task_points` | `string` — Task point, indicating the point that needs to be stop during the execution of the task. It can be empty if not needed |
| `type` | `string` — TAKEOUT 外卖配送任务（仅限有货仓的机型） GUIDE 引领任务，将人带到指定目的地 FOOD_DELIVERY 送餐任务 RETURN 快速返航，回到取餐点 RECYCLE 回收餐盘 TAKEOUT_DISTRIBUTE 外卖分发，打开所有舱门由用户自主取物 — Enum: `[ TAKEOUT, GUIDE, FOOD_DELIVERY, RETURN, RECYCLE, TAKEOUT_DISTRIBUTE ]` |
| `cargos` | `string` |
| `cargo_id` | `string` |
| `boxes` |  |

**Responses:**

- **200** — OK

  | Field | Type / Description |
  |-------|-------------------|
  | `result` *(required)* | `boolean` |
  | `order_ids` | `string` |
  | `errors` |  |


---

### `DELETE` `/api/delivery/v1/tasks/{task_id}`

**Cancel task according to task ID**

Some tasks are distributed through the cloud, and there may be no order number, so you need to cancel them through the task ID

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `task_id` *(required)* | `string` | path |  |

**Responses:**

- **200** — OK

---

### `DELETE` `/api/delivery/v1/tasks/orders/{order_id}`

**Cancel task according to order ID**

All tasks created on the robot side will contain the order number, so you can cancel the task through the order number

**Parameters:**

| Parameter | Type | In | Description |
|-----------|------|----|-------------|
| `order_id` *(required)* | `string` | path |  |

**Responses:**

- **200** — OK

---

### `GET` `/api/delivery/v1/stage`

**Get current task status**

DEVICE_ERROR Device repor error information, robot can not move, APP should display a fault page. GOING_TO_TASK_POINT Going to task point，some tasks (such as return) need to stop at some task points in the middle, and then go to the target point after completing the operation. ARRIVED_AT_TASK_POINT arrived at task point, robot will wait for operation. ON_DELIVERING going to target point ARRIVED_AT_TARGET arraive at target ON_RETURNING returning to parking point GOING_HOME going back to home dock IDLE robot stop at parking point or home dock.

**Responses:**

- **200** — OK

  **`TaskStage`**
  
  | Field | Type / Description |
  |-------|-------------------|
    *The return data is one of the following structures, first judging the stage field, and then parsing the remaining information*
    *corresponding stage is DEVICE_ERROR*
  | `stage` | `string` — Enum: `[ DEVICE_ERROR ]` |
  | `info` |  |
  | `errors` |  |
    *corresponding stage is GOING_TO_TASK_POINT*
  | `stage` | `string` — Enum: `[ GOING_TO_TASK_POINT ]` |
  | `current_floor` | `string` |
  | `target_floor` | `string` |
  | `info` | `string` — Enum: `Array [ 9 ]` |
  | `task` | `string` — Enum: `Array [ 9 ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `Array [ 9 ]` |
  | `req_id` | `string` |
  | `order_id` | `string` |
  | `no_pickup_wait` | `boolean` |
  | `cargos` |  |
  | `failed_tasks` |  |
  | `station_id` | `string` |
  | `station_cargos` |  |
  | `message` |  |
    *corresponding stage is ARRIVED_AT_TASK_POINT*
  | `stage` | `string` — Enum: `[ ARRIVED_AT_TASK_POINT ]` |
  | `info` | `string` — Enum: `Array [ 9 ]` |
  | `task` | `string` — Enum: `Array [ 9 ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `Array [ 9 ]` |
  | `req_id` | `string` |
  | `order_id` | `string` |
  | `no_pickup_wait` | `boolean` |
  | `cargos` |  |
  | `failed_tasks` |  |
  | `station_id` | `string` |
  | `station_cargos` |  |
  | `message` |  |
  | `location` | `string` — Enum: `Array [ 7 ]` |
  | `poi_name` | `string` |
  | `type` | `string` — Enum: `Array [ 7 ]` |
    *corresponding stage is ON_DELIVERING*
  | `stage` | `string` — Enum: `[ ON_DELIVERING ]` |
  | `milestone` | `string` — Task execution stage — Enum: `[ GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, ON_DELIVERING, ARRIVED_AT_DELIVERY_POSE, WAIT_REFILL_TIMEOUT, USER_OPERATE_ROBOT, GOING_TO_CABINET, ROBOT_OPERATE_CABINET, GOING_TO_TASK_POINT, ARRIVED_AT_TASK_POINT ]` |
  | `current_floor` | `string` |
  | `target_floor` | `string` |
  | `info` | `string` — Enum: `Array [ 9 ]` |
  | `task` | `string` — Enum: `Array [ 9 ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `Array [ 9 ]` |
  | `req_id` | `string` |
  | `order_id` | `string` |
  | `no_pickup_wait` | `boolean` |
  | `cargos` |  |
  | `failed_tasks` |  |
  | `station_id` | `string` |
  | `station_cargos` |  |
  | `message` |  |
  | `location` | `string` — Enum: `Array [ 7 ]` |
  | `poi_name` | `string` |
  | `type` | `string` — Enum: `Array [ 7 ]` |
    *corresponding stage is ARRIVED_AT_TARGET*
  | `stage` | `string` — Enum: `[ ARRIVED_AT_TARGET ]` |
  | `info` | `string` — Task execution stage — Enum: `Array [ 9 ]` |
  | `id` | `string` |
  | `task` | `string` — Enum: `Array [ 9 ]` |
  | `target` | `string` |
  | `type` | `string` — Enum: `Array [ 9 ]` |
  | `req_id` | `string` |
  | `order_id` | `string` |
  | `no_pickup_wait` | `boolean` |
  | `cargos` |  |
  | `failed_tasks` |  |
  | `station_id` | `string` |
  | `station_cargos` |  |
  | `message` |  |
  | `status` | `string` — Enum: `[ READY, RUNNING, SUCCEEDED, FAILED, CANCELING, CANCELED ]` |
  | `result` | `string` — Task execution stage — Enum: `Array [ 13 ]` |
  | `stage` | `string` — Task execution stage — Enum: `Array [ 13 ]` |
  | `reason` | `string` |
  | `pickup_result` |  |
  | `timestamp` | `string` |
  | `pickup` | `integer` |
  | `num_total` | `integer` |
  | `num_picked_up` | `integer` |
  | `result` |  |
    *corresponding stage is ON_RETURNING*
  | `stage` | `string` — Enum: `[ ON_RETURNING ]` |
  | `current_floor` | `string` |
  | `target_floor` | `string` |
    *corresponding stage is GOING_HOME*
  | `stage` | `string` — Enum: `[ GOING_HOME ]` |
  | `milestone` | `string` — Enum: `[ INITIALIZING, GOING_TO_ELEVATOR, WAIT_FOR_ELEVATOR, GOING_INTO_ELEVATOR, IN_ELEVATOR, GOING_OUT_OF_ELEVATOR, GOING_TO_LANDING_POINT, GOING_HOME ]` |
  | `current_floor` | `string` |
  | `target_floor` | `string` |
    *corresponding stage is IDLE*
  | `stage` | `string` — Enum: `[ IDLE ]` |
  | `current_floor` | `string` |


---

### `PUT` `/api/delivery/v1/tasks/:task_execution`

**Pause / Resume task**

When user operates the APP, set false to prohibit the robot movement, the robot will not run even receiving the task; When the user completes the operation, set it to true to allow the robot to move.

**Request Body** (`application/json`):


| Field | Type / Description |
|-------|-------------------|
| `enable_task_execution` | `boolean` |

**Responses:**

- **200** — OK

  **`TaskExecutionInfo`**
  
  | Field | Type / Description |
  |-------|-------------------|
  | `enable_task_execution` | `boolean` |


---

### `PUT` `/api/delivery/v1/tasks/:task_finish`

**Finish all tasks**

The difference from the Delete API is that this API ends all tasks in a successful state

**Responses:**

- **204** — No Content

---

### `PUT` `/api/delivery/v1/tasks/:start_pickup`

**Start picking up**

**Responses:**

- **204** — No Content

---

### `PUT` `/api/delivery/v1/tasks/:end_pickup`

**End picking up**

**Responses:**

- **204** — No Content

---

### `PUT` `/api/delivery/v1/tasks/:end_operation`

**Complete operation**

When the robot arrives at task point, the API is used to notify the robot that the operation has been completed and it can continue the task.

**Responses:**

- **204** — No Content

---
