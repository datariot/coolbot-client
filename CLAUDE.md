# CoolBot Client

MQTT exporter for CoolBot Pro temperature controllers. Connects via WebSocket (Blynk protocol), publishes real-time metrics to MQTT for Telegraf/InfluxDB ingestion.

## Structure
```
cmd/coolbot-exporter/main.go    # CLI entry point
internal/
  client/client.go              # WebSocket connection, device state
  metrics/mqtt.go               # MQTT publisher
  protocol/                     # Blynk auth, message parsing, pin defs
docs/PROTOCOL.md                # Protocol reverse engineering notes
```

## Commands
```
task build          # → bin/coolbot-exporter
task run            # build + run
task dev            # run with debug logging
task test           # run tests
task docker         # build container image
```

## Config
Credentials via flags or env vars: `COOLBOT_EMAIL`, `COOLBOT_PASSWORD`
MQTT broker defaults to `tcp://minis.local:1883`, configurable topic prefix and publish interval.

## Dependencies
- `github.com/coder/websocket` — WebSocket client
- `github.com/eclipse/paho.mqtt.golang` — MQTT publisher
