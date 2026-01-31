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

| Pin | Description | Type | Range/Values |
|-----|-------------|------|--------------|
| v0 | Room Temperature | Float | 0-110°F |
| v1 | Frost/Fin Temperature | Float | 0-110°F |
| v2 | Status | String | "OK", error codes |
| v4 | Set Point Temperature | Float | 36-45°F |
| v6 | Mode | Integer | Cooling modes |
| v7 | Cycle Time | String | "30 sec", etc. |
| v8 | Enable | Boolean | 1/0 |
| v9 | Compressor State | Integer | 0/1 |
| v10 | Humidity | Float | % |
| v12 | Max Set Point | Float | 42-45°F |
| v13 | Heat Mode | Boolean | 1/0 |
| v14 | RSSI (Device 0) | Integer | -100 to 0 dBm |
| v16 | Frost Minimum | Float | Alarm threshold |
| v17 | Frost Alarm | Boolean | 1/0 |
| v18 | RSSI | Integer | Signal strength |
| v19 | Compressor Run Count | Integer | Total cycles |
| v20 | Firmware Version | String | e.g., "7.9.5" |
| v22 | Night Mode Enable | Boolean | 1/0 |
| v23 | Power State | Boolean | 1/0 |
| v24 | Fan State | Boolean | 1/0 |
| v25 | MAC Address | String | HEX format |
| v26 | Status Flag | Boolean | 1/0 |
| v30 | Sensor Error | Boolean | 1/0 |
| v31 | Frost Sensor Error | Boolean | 1/0 |
| v32 | Heater Error | Boolean | 1/0 |
| v36 | Hardware Revision | String | e.g., "7.A" |
| v100 | Offline Flag | Boolean | 1/0 |

## Device Information Structure

```json
{
  "id": 0,
  "name": "CoolBot",
  "boardType": "ESP8266",
  "token": "cd47fa5be7f449459f1eb04abf2adba3",
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
