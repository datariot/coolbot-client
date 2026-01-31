// Package metrics provides Prometheus metrics for CoolBot devices.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Temperature metrics
	RoomTemperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "room_temperature_fahrenheit",
			Help:      "Current room temperature in Fahrenheit",
		},
		[]string{"device_id", "device_name"},
	)

	FrostTemperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "frost_temperature_fahrenheit",
			Help:      "Current frost/fin temperature in Fahrenheit",
		},
		[]string{"device_id", "device_name"},
	)

	SetPointTemperature = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "set_point_fahrenheit",
			Help:      "Target set point temperature in Fahrenheit",
		},
		[]string{"device_id", "device_name"},
	)

	// Environmental metrics
	Humidity = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "humidity_percent",
			Help:      "Current humidity percentage",
		},
		[]string{"device_id", "device_name"},
	)

	// Device status metrics
	CompressorState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "compressor_state",
			Help:      "Compressor state (0=off, 1=on)",
		},
		[]string{"device_id", "device_name"},
	)

	RSSI = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "wifi_rssi_dbm",
			Help:      "WiFi signal strength in dBm",
		},
		[]string{"device_id", "device_name"},
	)

	DeviceOnline = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "device_online",
			Help:      "Device online status (0=offline, 1=online)",
		},
		[]string{"device_id", "device_name"},
	)

	// Error metrics
	SensorError = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "sensor_error",
			Help:      "Sensor error flag (0=ok, 1=error)",
		},
		[]string{"device_id", "device_name"},
	)

	FrostSensorError = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "frost_sensor_error",
			Help:      "Frost sensor error flag (0=ok, 1=error)",
		},
		[]string{"device_id", "device_name"},
	)

	FrostAlarm = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "frost_alarm",
			Help:      "Frost alarm flag (0=ok, 1=alarm)",
		},
		[]string{"device_id", "device_name"},
	)

	// Connection metrics
	LastUpdateTimestamp = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "last_update_timestamp_seconds",
			Help:      "Unix timestamp of last update from device",
		},
		[]string{"device_id", "device_name"},
	)

	// Info metric for device metadata
	DeviceInfo = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "coolbot",
			Name:      "device_info",
			Help:      "Device information (always 1, labels contain metadata)",
		},
		[]string{"device_id", "device_name", "firmware_version", "status"},
	)
)
