// Package metrics provides MQTT publishing for CoolBot devices.
package metrics

import (
	"fmt"
	"log/slog"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

const (
	// TopicPrefix is the base topic for CoolBot sensors.
	TopicPrefix = "farm/sensors/coolbot"

	// DefaultQoS for MQTT messages (0 = at most once, 1 = at least once).
	DefaultQoS = 1

	// Retain messages so new subscribers get last known values.
	DefaultRetain = true
)

// Publisher publishes CoolBot metrics to MQTT.
type Publisher struct {
	client mqtt.Client
	logger *slog.Logger
	prefix string
	qos    byte
	retain bool
}

// Config holds MQTT connection configuration.
type Config struct {
	Broker   string // e.g., "tcp://minis.local:1883"
	ClientID string // Unique client identifier
	Username string // Optional
	Password string // Optional
	Prefix   string // Topic prefix (default: farm/sensors/coolbot)
}

// NewPublisher creates a new MQTT publisher.
func NewPublisher(cfg Config, logger *slog.Logger) (*Publisher, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID).
		SetAutoReconnect(true).
		SetConnectRetry(true).
		SetConnectRetryInterval(5 * time.Second).
		SetConnectionLostHandler(func(c mqtt.Client, err error) {
			logger.Warn("mqtt connection lost", "error", err)
		}).
		SetOnConnectHandler(func(c mqtt.Client) {
			logger.Info("mqtt connected")
		})

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username)
		opts.SetPassword(cfg.Password)
	}

	client := mqtt.NewClient(opts)
	token := client.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	prefix := cfg.Prefix
	if prefix == "" {
		prefix = TopicPrefix
	}

	return &Publisher{
		client: client,
		logger: logger,
		prefix: prefix,
		qos:    DefaultQoS,
		retain: DefaultRetain,
	}, nil
}

// Close disconnects from the MQTT broker.
func (p *Publisher) Close() {
	p.client.Disconnect(1000)
}

// DeviceMetrics holds metrics for a single device.
type DeviceMetrics struct {
	DeviceID        int
	DeviceName      string
	RoomTemp        float64
	FrostTemp       float64
	SetPoint        float64
	Humidity        float64
	CompressorState int
	RSSI            int
	Online          bool
	FirmwareVersion string
	Status          string
}

// PublishDeviceMetrics publishes all device metrics to MQTT.
func (p *Publisher) PublishDeviceMetrics(m DeviceMetrics) error {
	// Build topic suffix from device name if we have multiple devices
	// For now, publish directly under the prefix

	metrics := map[string]string{
		"room_temp":        fmt.Sprintf("%.2f", m.RoomTemp),
		"frost_temp":       fmt.Sprintf("%.2f", m.FrostTemp),
		"set_point":        fmt.Sprintf("%.2f", m.SetPoint),
		"humidity":         fmt.Sprintf("%.1f", m.Humidity),
		"compressor_state": fmt.Sprintf("%d", m.CompressorState),
		"rssi":             fmt.Sprintf("%d", m.RSSI),
		"online":           boolToStr(m.Online),
	}

	var lastErr error
	for attr, value := range metrics {
		topic := fmt.Sprintf("%s/%s", p.prefix, attr)
		token := p.client.Publish(topic, p.qos, p.retain, value)
		if token.Wait() && token.Error() != nil {
			p.logger.Error("mqtt publish failed", "topic", topic, "error", token.Error())
			lastErr = token.Error()
		} else {
			p.logger.Debug("published", "topic", topic, "value", value)
		}
	}

	return lastErr
}

// PublishSingle publishes a single metric value.
func (p *Publisher) PublishSingle(attribute string, value string) error {
	topic := fmt.Sprintf("%s/%s", p.prefix, attribute)
	token := p.client.Publish(topic, p.qos, p.retain, value)
	if token.Wait() && token.Error() != nil {
		return token.Error()
	}
	return nil
}

func boolToStr(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// IsConnected returns true if connected to the broker.
func (p *Publisher) IsConnected() bool {
	return p.client.IsConnected()
}
