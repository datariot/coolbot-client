package protocol

// Virtual pin definitions for CoolBot Pro devices.
const (
	PinRoomTemp          = 0  // Room temperature (°F)
	PinFrostTemp         = 1  // Frost/fin temperature (°F)
	PinStatus            = 2  // Status string ("OK", error codes)
	PinSetPoint          = 4  // Set point temperature (°F)
	PinMode              = 6  // Cooling mode
	PinCycleTime         = 7  // Cycle time string
	PinEnable            = 8  // Enable flag
	PinCompressorState   = 9  // Compressor state
	PinHumidity          = 10 // Humidity (%)
	PinMaxSetPoint       = 12 // Max set point (°F)
	PinHeatMode          = 13 // Heat mode flag
	PinRSSI0             = 14 // RSSI for device 0
	PinFrostMinimum      = 16 // Frost minimum alarm threshold
	PinFrostAlarm        = 17 // Frost alarm flag
	PinRSSI              = 18 // RSSI signal strength
	PinCompressorCount   = 19 // Compressor run count
	PinFirmwareVersion   = 20 // Firmware version string
	PinNightModeEnable   = 22 // Night mode enable flag
	PinPowerState        = 23 // Power state
	PinFanState          = 24 // Fan state
	PinMACAddress        = 25 // MAC address (hex)
	PinStatusFlag        = 26 // Status flag
	PinSensorError       = 30 // Sensor error flag
	PinFrostSensorError  = 31 // Frost sensor error flag
	PinHeaterError       = 32 // Heater error flag
	PinHardwareRevision  = 36 // Hardware revision string
	PinOfflineFlag       = 100 // Offline notification flag
)

// PinName returns a human-readable name for a virtual pin.
func PinName(pin int) string {
	names := map[int]string{
		PinRoomTemp:         "room_temp",
		PinFrostTemp:        "frost_temp",
		PinStatus:           "status",
		PinSetPoint:         "set_point",
		PinMode:             "mode",
		PinCycleTime:        "cycle_time",
		PinEnable:           "enable",
		PinCompressorState:  "compressor_state",
		PinHumidity:         "humidity",
		PinMaxSetPoint:      "max_set_point",
		PinHeatMode:         "heat_mode",
		PinRSSI0:            "rssi_device0",
		PinFrostMinimum:     "frost_minimum",
		PinFrostAlarm:       "frost_alarm",
		PinRSSI:             "rssi",
		PinCompressorCount:  "compressor_count",
		PinFirmwareVersion:  "firmware_version",
		PinNightModeEnable:  "night_mode",
		PinPowerState:       "power_state",
		PinFanState:         "fan_state",
		PinMACAddress:       "mac_address",
		PinStatusFlag:       "status_flag",
		PinSensorError:      "sensor_error",
		PinFrostSensorError: "frost_sensor_error",
		PinHeaterError:      "heater_error",
		PinHardwareRevision: "hardware_revision",
		PinOfflineFlag:      "offline",
	}
	if name, ok := names[pin]; ok {
		return name
	}
	return "unknown"
}

// IsNumericPin returns true if the pin typically holds numeric values.
func IsNumericPin(pin int) bool {
	numeric := map[int]bool{
		PinRoomTemp:        true,
		PinFrostTemp:       true,
		PinSetPoint:        true,
		PinMode:            true,
		PinEnable:          true,
		PinCompressorState: true,
		PinHumidity:        true,
		PinMaxSetPoint:     true,
		PinHeatMode:        true,
		PinRSSI0:           true,
		PinFrostMinimum:    true,
		PinFrostAlarm:      true,
		PinRSSI:            true,
		PinCompressorCount: true,
		PinNightModeEnable: true,
		PinPowerState:      true,
		PinFanState:        true,
		PinStatusFlag:      true,
		PinSensorError:     true,
		PinFrostSensorError: true,
		PinHeaterError:     true,
		PinOfflineFlag:     true,
	}
	return numeric[pin]
}
