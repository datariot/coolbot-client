# CoolBot Client

A Go client library for [CoolBot Pro](https://www.storeitcold.com/product/coolbot-pro/) temperature controllers that publishes metrics to MQTT.

## Features

- WebSocket client for CoolBot Pro's Blynk-based protocol
- Real-time temperature and device status updates
- MQTT publishing for integration with Telegraf/InfluxDB
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
  -mqtt-broker tcp://minis.local:1883

# Using environment variables
export COOLBOT_EMAIL=user@example.com
export COOLBOT_PASSWORD=yourpassword
coolbot-exporter
```

### Options

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `-email` | `COOLBOT_EMAIL` | - | CoolBot account email |
| `-password` | `COOLBOT_PASSWORD` | - | CoolBot account password |
| `-server` | - | `cb.storeitcold.com` | CoolBot server |
| `-mqtt-broker` | `MQTT_BROKER` | `tcp://minis.local:1883` | MQTT broker URL |
| `-mqtt-client-id` | - | `coolbot-exporter` | MQTT client ID |
| `-mqtt-username` | `MQTT_USERNAME` | - | MQTT username (optional) |
| `-mqtt-password` | `MQTT_PASSWORD` | - | MQTT password (optional) |
| `-mqtt-prefix` | - | `farm/sensors/coolbot` | MQTT topic prefix |
| `-interval` | - | `15s` | Publish interval |
| `-log-level` | - | `info` | Log level (debug, info, warn, error) |

## MQTT Topics

Data is published to `farm/sensors/coolbot/<attribute>` with plain float values:

| Topic | Description | Example |
|-------|-------------|---------|
| `farm/sensors/coolbot/room_temp` | Room temperature (°F) | `41.50` |
| `farm/sensors/coolbot/frost_temp` | Evaporator-fin temperature (°F) | `39.20` |
| `farm/sensors/coolbot/set_point` | Target set point (°F) | `42.00` |
| `farm/sensors/coolbot/heater_on` | Calling for cooling: heater warming the AC's sensor | `1` |
| `farm/sensors/coolbot/power_on` | Power switch in the app | `1` |
| `farm/sensors/coolbot/status` | Controller status word (`OK`, `Hot`, `Cold`, faults) | `OK` |
| `farm/sensors/coolbot/too_hot_temp` | Too-hot alert threshold (°F) | `45.00` |
| `farm/sensors/coolbot/too_cold_temp` | Too-cold alert threshold (°F) | `38.00` |
| `farm/sensors/coolbot/alarm_hot` | Too-hot alarm (> 0 = raised; the controller sends 255) | `255` |
| `farm/sensors/coolbot/alarm_cold` | Too-cold alarm (> 0 = raised) | `0` |
| `farm/sensors/coolbot/room_sensor_error` | Room sensor fault (> 0 = raised) | `0` |
| `farm/sensors/coolbot/fins_sensor_error` | Fins sensor fault (> 0 = raised) | `0` |
| `farm/sensors/coolbot/heater_error` | Heater fault (> 0 = raised) | `0` |
| `farm/sensors/coolbot/rssi` | WiFi signal strength (dBm) | `-71` |
| `farm/sensors/coolbot/online` | Cloud reports the controller connected | `1` |

Real-time pin changes also publish under the names in
[docs/PROTOCOL.md](docs/PROTOCOL.md#virtual-pin-mapping); a pin the vendor
table does not name goes out as `pin_<n>`. There is no compressor or humidity
topic: the CoolBot has neither reading.

## Telegraf Configuration

Add to your `telegraf.conf` to consume these topics:

```toml
[[inputs.mqtt_consumer]]
  servers = ["tcp://localhost:1883"]
  topics = ["farm/sensors/coolbot/#"]
  data_format = "value"
  data_type = "float"

  [[inputs.mqtt_consumer.topic_parsing]]
    topic = "farm/sensors/+/+"
    measurement = "measurement/_/_"
    tags = "_/device/field"
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

## Testing with mosquitto

```bash
# Subscribe to all coolbot topics
mosquitto_sub -h minis.local -t "farm/sensors/coolbot/#" -v

# You should see:
# farm/sensors/coolbot/room_temp 41.50
# farm/sensors/coolbot/frost_temp 39.20
# farm/sensors/coolbot/set_point 42.00
# ...
```

## Protocol Documentation

See [docs/PROTOCOL.md](docs/PROTOCOL.md) for details on the Blynk-based WebSocket protocol.

## License

MIT
