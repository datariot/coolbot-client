# CoolBot Pro - Blynk Protocol Specification

## Overview

CoolBot Pro uses a customized Blynk IoT platform for device communication. This document describes the WebSocket protocol for building third-party clients.

## Connection Details

| Property | Value |
|----------|-------|
| Protocol | Blynk WebSocket (Secure) |
| Transport | WebSocket over HTTPS |
| Server | `cb.storeitcold.com` |
| App Version | 1.22.2 |
| Build Number | 12220000 |

## Message Format

```
COMMAND_NAME @message_id: [param1·param2·param3·...]
```

- **Separator**: Middle dot `·` (U+00B7) - NOT underscore or dash
- **Message ID**: Incremental counter per session (starts at 1)
- **Encoding**: UTF-8 (GZIP for profile data)

## Authentication Flow

### 1. LOGIN Request

```
LOGIN @1: [email·password_hash·clientType·buildNumber·platform]
```

**Parameters:**
- `email`: User email address
- `password_hash`: Base64-encoded password hash
- `clientType`: "Other" (or device-specific identifier)
- `buildNumber`: Version identifier (12220000)
- `platform`: "Blynk" (or device platform)

**Example:**
```
LOGIN @1: [user@example.com·I8bWMbPeFIpL6+Zp5NsOJVYOALwGN9GgktVuAFhXcJE=·Other·12220000·Blynk]
```

### 2. LOGIN Response

```
RESPONSE @1 = Ok
```

## Core Commands

### LOAD_PROFILE_GZIPPED

Fetches the complete user profile including all devices and current values.

```
LOAD_PROFILE_GZIPPED @2: []
```

**Response**: GZIP-compressed JSON containing:
- Dashboard configurations
- Device list with tokens and status
- Widget definitions and pin mappings
- All virtual pin current values

### GET_PROJECT_BY_TOKEN

Gets device project details by token.

```
GET_PROJECT_BY_TOKEN @3: [device_token]
```

**Parameters:**
- `device_token`: 32-character hex device token

### HARDWARE (Real-time Updates)

Streaming updates use a special message ID `@7778`:

```
HARDWARE @7778: [dashboardId·vw·pin·value]
```

**Parameters:**
- `dashboardId`: Project dashboard ID
- `vw`: Data type (virtual write)
- `pin`: Virtual pin number (0-255)
- `value`: Current value

## Virtual Pin Mapping

From the vendor web app's own pin table (cb.storeitcold.com `main.js`, the
CoolBot device class's `pinAssignments`) and the app's widget labels. An
earlier version of this table guessed; v9 and v10 were wrong (KB-1163). The
CoolBot has **no compressor or humidity reading**: it runs the AC by warming
the AC's temperature sensor, so v26 `heaterOn` is its call for cooling.

| Pin | Vendor name | Meaning | Example |
|-----|-------------|---------|---------|
| v0 | roomTemp | Room temperature (°F) | `55.652` |
| v1 | finsTemp | Evaporator-fin temperature (°F) | `56.390` |
| v2 | status | Status word: `OK`, `Hot`, `Cold`, `Room`/`Fins`/`Heater`/`Hardware` faults | `Hot` |
| v4 | setTemp | Set point (°F) | `41.998` |
| v6 | finsSetTemp | Fin-temperature setting (integer) | `1` |
| v8 | heaterDelay | Heater delay (integer) | `1` |
| v9 | powerOn | Power switch in the app (0/1) | `1` |
| v10 | notifyOn | Alerts enabled (0/1) | `1` |
| v12 | tooHotTemp | Too-hot alert threshold (°F) | `45` |
| v16 | tooColdTemp | Too-cold alert threshold (°F) | `38` |
| v18 | (RSSI widget) | Wi-Fi signal (dBm) | `-70` |
| v20 | firmwareVersion | Firmware version | `7.9.6` |
| v22 | notifyEmailOn | Email alerts (0/1) | `0` |
| v23 | notifySmsOn | SMS alerts (0/1) | `1` |
| v25 | macAddress | MAC address | |
| v26 | heaterOn | Heater warming the AC sensor = calling for cooling (0/1) | `1` |
| v36 | hardwareVersion | Hardware revision | `7.A` |
| v37 | notifyOnSetPointChange | Alert on set-point change (0/1) | |
| v100 | offlineFlag | Offline flag (0/1) | `0` |

Other pins the controller writes (v7 `30 sec`, v13, v14, v17, v19, v21, v24,
v30–v35) have no name in the vendor table; the exporter publishes them as
`pin_<n>`.

## Device Information Structure

```json
{
  "id": 0,
  "name": "CoolBot",
  "boardType": "ESP8266",
  "token": "<device token>",
  "connectionType": "WI_FI",
  "status": "ONLINE",
  "disconnectTime": 1769824786398,
  "connectTime": 1769828523150,
  "lastLoggedIP": "67.165.209.53",
  "hardwareInfo": {
    "version": "0.4.8",
    "boardType": "Arduino",
    "build": "Oct 28 2024 17:45:49",
    "heartbeatInterval": 10
  }
}
```

## Connection Lifecycle

1. Establish WebSocket connection to `wss://cb.storeitcold.com/...`
2. Send `LOGIN` with credentials → wait for `RESPONSE`
3. Send `LOAD_PROFILE_GZIPPED` → receive profile data
4. Send `GET_PROJECT_BY_TOKEN` for each device (optional)
5. Listen for `HARDWARE` messages continuously (real-time updates)
6. Maintain heartbeat (10 second intervals)

## Implementation Notes

- Password appears to be pre-hashed before transmission (likely SHA256)
- Uses GZIP compression for large profile responses
- Message ordering is important (LOGIN before LOAD_PROFILE)
- Real-time updates use message ID @7778 (streaming format)
- Updates arrive at 10-15 second intervals per device
- Device tokens are 32-character hex strings
