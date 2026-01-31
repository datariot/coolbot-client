# CoolBot Client

A Go client library and Prometheus exporter for [CoolBot Pro](https://www.storeitcold.com/product/coolbot-pro/) temperature controllers.

## Features

- WebSocket client for CoolBot Pro's Blynk-based protocol
- Real-time temperature and device status updates
- Prometheus metrics exporter for Grafana integration
- Support for multiple devices

## Installation

```bash
go install github.com/datariot/coolbot-client/cmd/coolbot-exporter@latest
```

Or build from source:

```bash
git clone https://github.com/datariot/coolbot-client.git
cd coolbot-client
go build -o coolbot-exporter ./cmd/coolbot-exporter
```

## Usage

### Prometheus Exporter

```bash
# Using flags
coolbot-exporter -email user@example.com -password yourpassword

# Using environment variables
export COOLBOT_EMAIL=user@example.com
export COOLBOT_PASSWORD=yourpassword
coolbot-exporter
```

Options:
- `-listen`: Address to listen on (default: `:9120`)
- `-email`: CoolBot account email
- `-password`: CoolBot account password
- `-server`: CoolBot server (default: `cb.storeitcold.com`)
- `-log-level`: Log level - debug, info, warn, error (default: `info`)

### Prometheus Configuration

Add to your `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'coolbot'
    static_configs:
      - targets: ['localhost:9120']
```

## Metrics

| Metric | Description | Labels |
|--------|-------------|--------|
| `coolbot_room_temperature_fahrenheit` | Current room temperature | device_id, device_name |
| `coolbot_frost_temperature_fahrenheit` | Frost/fin temperature | device_id, device_name |
| `coolbot_set_point_fahrenheit` | Target set point | device_id, device_name |
| `coolbot_humidity_percent` | Humidity percentage | device_id, device_name |
| `coolbot_compressor_state` | Compressor on/off | device_id, device_name |
| `coolbot_wifi_rssi_dbm` | WiFi signal strength | device_id, device_name |
| `coolbot_device_online` | Device online status | device_id, device_name |
| `coolbot_sensor_error` | Sensor error flag | device_id, device_name |
| `coolbot_frost_alarm` | Frost alarm flag | device_id, device_name |
| `coolbot_last_update_timestamp_seconds` | Last update time | device_id, device_name |
| `coolbot_device_info` | Device metadata | device_id, device_name, firmware_version, status |

## Grafana Dashboard

Example queries:

```promql
# Current room temperature
coolbot_room_temperature_fahrenheit{device_name="CoolBot"}

# Temperature vs set point
coolbot_room_temperature_fahrenheit - coolbot_set_point_fahrenheit

# Compressor duty cycle (over 1 hour)
avg_over_time(coolbot_compressor_state[1h]) * 100
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
