# CoolBot Client

A Go client library for [CoolBot Pro](https://www.storeitcold.com/product/coolbot-pro/) temperature controllers that writes metrics to InfluxDB.

## Features

- WebSocket client for CoolBot Pro's Blynk-based protocol
- Real-time temperature and device status updates
- Direct InfluxDB integration for time-series storage
- Support for multiple devices

## Installation

```bash
go install github.com/datariot/coolbot-client/cmd/coolbot-exporter@latest
```

Or build from source:

```bash
git clone https://github.com/datariot/coolbot-client.git
cd coolbot-client
task build
```

## Usage

```bash
# Using flags
coolbot-exporter \
  -email user@example.com \
  -password yourpassword \
  -influx-url http://minis:8086 \
  -influx-token your-influxdb-token \
  -influx-org home \
  -influx-bucket coolbot

# Using environment variables
export COOLBOT_EMAIL=user@example.com
export COOLBOT_PASSWORD=yourpassword
export INFLUXDB_TOKEN=your-influxdb-token
coolbot-exporter
```

### Options

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `-email` | `COOLBOT_EMAIL` | - | CoolBot account email |
| `-password` | `COOLBOT_PASSWORD` | - | CoolBot account password |
| `-server` | - | `cb.storeitcold.com` | CoolBot server |
| `-influx-url` | `INFLUXDB_URL` | `http://minis:8086` | InfluxDB server URL |
| `-influx-token` | `INFLUXDB_TOKEN` | - | InfluxDB API token |
| `-influx-org` | - | `home` | InfluxDB organization |
| `-influx-bucket` | - | `coolbot` | InfluxDB bucket |
| `-interval` | - | `15s` | Metrics write interval |
| `-log-level` | - | `info` | Log level (debug, info, warn, error) |

## InfluxDB Setup

1. Create a bucket named `coolbot` in your InfluxDB instance
2. Generate an API token with write access to the bucket
3. Run the exporter with your credentials

## Metrics

Data is written to the `coolbot` measurement with the following fields:

| Field | Type | Description |
|-------|------|-------------|
| `room_temp_f` | float | Room temperature (°F) |
| `frost_temp_f` | float | Frost/fin temperature (°F) |
| `set_point_f` | float | Target set point (°F) |
| `humidity_pct` | float | Humidity percentage |
| `compressor_state` | int | Compressor on/off (0/1) |
| `rssi_dbm` | int | WiFi signal strength (dBm) |
| `online` | int | Device online status (0/1) |
| `firmware_version` | string | Firmware version |
| `status` | string | Device status |

Tags:
- `device_id` - Device identifier
- `device_name` - Device name

## Flux Queries

```flux
// Current room temperature
from(bucket: "coolbot")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "coolbot")
  |> filter(fn: (r) => r._field == "room_temp_f")

// Temperature vs set point
from(bucket: "coolbot")
  |> range(start: -24h)
  |> filter(fn: (r) => r._measurement == "coolbot")
  |> filter(fn: (r) => r._field == "room_temp_f" or r._field == "set_point_f")

// Compressor duty cycle
from(bucket: "coolbot")
  |> range(start: -1h)
  |> filter(fn: (r) => r._measurement == "coolbot")
  |> filter(fn: (r) => r._field == "compressor_state")
  |> mean()
  |> map(fn: (r) => ({r with _value: r._value * 100.0}))
```

## Library Usage

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/datariot/coolbot-client/internal/client"
)

func main() {
    cb := client.New("user@example.com", "password")

    if err := cb.Connect(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer cb.Close()

    // Register for updates
    cb.OnUpdate(func(update client.DeviceUpdate) {
        fmt.Printf("%s = %s\n", update.PinName, update.Value)
    })

    // Get current device state
    devices := cb.GetDevices()
    for _, dev := range devices {
        fmt.Printf("Device %s: %.1f°F\n", dev.Name, dev.RoomTemp)
    }

    // Listen for real-time updates
    cb.Listen(context.Background())
}
```

## Protocol Documentation

See [docs/PROTOCOL.md](docs/PROTOCOL.md) for details on the Blynk-based WebSocket protocol.

## License

MIT
