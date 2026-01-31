// Package metrics provides InfluxDB integration for CoolBot devices.
package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

// Writer writes CoolBot metrics to InfluxDB.
type Writer struct {
	client   influxdb2.Client
	writeAPI api.WriteAPIBlocking
	logger   *slog.Logger
	org      string
	bucket   string
}

// Config holds InfluxDB connection configuration.
type Config struct {
	URL    string // e.g., "http://minis:8086"
	Token  string // InfluxDB API token
	Org    string // Organization name
	Bucket string // Bucket name
}

// NewWriter creates a new InfluxDB metrics writer.
func NewWriter(cfg Config, logger *slog.Logger) *Writer {
	client := influxdb2.NewClient(cfg.URL, cfg.Token)
	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)

	return &Writer{
		client:   client,
		writeAPI: writeAPI,
		logger:   logger,
		org:      cfg.Org,
		bucket:   cfg.Bucket,
	}
}

// Close closes the InfluxDB client connection.
func (w *Writer) Close() {
	w.client.Close()
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

// WriteDeviceMetrics writes a complete set of device metrics to InfluxDB.
func (w *Writer) WriteDeviceMetrics(ctx context.Context, m DeviceMetrics) error {
	tags := map[string]string{
		"device_id":   fmt.Sprintf("%d", m.DeviceID),
		"device_name": m.DeviceName,
	}

	online := 0
	if m.Online {
		online = 1
	}

	fields := map[string]interface{}{
		"room_temp_f":       m.RoomTemp,
		"frost_temp_f":      m.FrostTemp,
		"set_point_f":       m.SetPoint,
		"humidity_pct":      m.Humidity,
		"compressor_state":  m.CompressorState,
		"rssi_dbm":          m.RSSI,
		"online":            online,
		"firmware_version":  m.FirmwareVersion,
		"status":            m.Status,
	}

	point := influxdb2.NewPoint(
		"coolbot",
		tags,
		fields,
		time.Now(),
	)

	if err := w.writeAPI.WritePoint(ctx, point); err != nil {
		w.logger.Error("failed to write metrics", "error", err, "device", m.DeviceName)
		return err
	}

	w.logger.Debug("wrote metrics", "device", m.DeviceName, "room_temp", m.RoomTemp)
	return nil
}

// WritePinUpdate writes a single pin update to InfluxDB.
func (w *Writer) WritePinUpdate(ctx context.Context, deviceID int, deviceName string, pinName string, value float64) error {
	tags := map[string]string{
		"device_id":   fmt.Sprintf("%d", deviceID),
		"device_name": deviceName,
		"pin":         pinName,
	}

	fields := map[string]interface{}{
		"value": value,
	}

	point := influxdb2.NewPoint(
		"coolbot_pins",
		tags,
		fields,
		time.Now(),
	)

	return w.writeAPI.WritePoint(ctx, point)
}

// Ping tests the InfluxDB connection.
func (w *Writer) Ping(ctx context.Context) error {
	health, err := w.client.Health(ctx)
	if err != nil {
		return err
	}
	if health.Status != "pass" {
		return fmt.Errorf("influxdb health check failed: %s", health.Status)
	}
	w.logger.Info("influxdb connection ok", "status", health.Status)
	return nil
}
