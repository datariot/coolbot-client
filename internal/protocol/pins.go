package protocol

import "strconv"

// Virtual pin definitions for CoolBot Pro devices.
//
// The meanings come from the vendor web app's own pin table
// (cb.storeitcold.com, the CoolBot device class's pinAssignments), and the
// values a live controller holds confirm the ones with an obvious shape
// (firmware, MAC, hardware revision). The first version of this file
// guessed, and named the power switch "compressor_state" and the alerts
// flag "humidity" (KB-1163). The CoolBot has no compressor or humidity
// reading: it runs the air conditioner by warming the AC's own temperature
// sensor, so PinHeaterOn is its call for cooling.
//
// A pin the vendor table does not name is published as pin_<n> rather than
// under a guessed name.
const (
	PinRoomTemp         = 0   // roomTemp: room temperature (°F)
	PinFinsTemp         = 1   // finsTemp: evaporator-fin temperature (°F)
	PinStatus           = 2   // status: "OK", "Hot", "Cold", "Room"/"Fins"/"Heater"/"Hardware" faults, …
	PinSetPoint         = 4   // setTemp: set point (°F)
	PinFinsSetTemp      = 6   // finsSetTemp: fin-temperature setting (integer)
	PinHeaterDelay      = 8   // heaterDelay (integer)
	PinPowerOn          = 9   // powerOn: the app's power switch (0/1)
	PinNotifyOn         = 10  // notifyOn: alerts enabled (0/1)
	PinTooHotTemp       = 12  // tooHotTemp: too-hot alert threshold (°F)
	PinTooColdTemp      = 16  // tooColdTemp: too-cold alert threshold (°F)
	PinRSSI             = 18  // RSSI (dBm), the app's "RSSI" widget
	PinFirmwareVersion  = 20  // firmware version string
	PinNotifyEmailOn    = 22  // notifyEmailOn (0/1)
	PinNotifySmsOn      = 23  // notifySmsOn (0/1)
	PinMACAddress       = 25  // MAC address
	PinHeaterOn         = 26  // heaterOn: warming the AC's sensor = calling for cooling (0/1)
	PinHardwareRevision = 36  // hardware revision string
	PinNotifyOnSetPoint = 37  // notifyOnSetPointChange (0/1)
	PinOfflineFlag      = 100 // offlineFlag (0/1)
)

var pinNames = map[int]string{
	PinRoomTemp:         "room_temp",
	PinFinsTemp:         "frost_temp",
	PinStatus:           "status",
	PinSetPoint:         "set_point",
	PinFinsSetTemp:      "fins_set_temp",
	PinHeaterDelay:      "heater_delay",
	PinPowerOn:          "power_on",
	PinNotifyOn:         "notify_on",
	PinTooHotTemp:       "too_hot_temp",
	PinTooColdTemp:      "too_cold_temp",
	PinRSSI:             "rssi",
	PinFirmwareVersion:  "firmware_version",
	PinNotifyEmailOn:    "notify_email_on",
	PinNotifySmsOn:      "notify_sms_on",
	PinMACAddress:       "mac_address",
	PinHeaterOn:         "heater_on",
	PinHardwareRevision: "hardware_revision",
	PinNotifyOnSetPoint: "notify_on_set_point_change",
	PinOfflineFlag:      "offline_flag",
}

// PinName returns the attribute name a virtual pin is published under:
// the vendor's meaning where it has one, pin_<n> where it does not.
func PinName(pin int) string {
	if name, ok := pinNames[pin]; ok {
		return name
	}
	return "pin_" + strconv.Itoa(pin)
}
