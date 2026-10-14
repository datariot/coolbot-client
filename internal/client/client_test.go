package client

import (
	"testing"

	"github.com/datariot/coolbot-client/internal/protocol"
)

// KB-1163: pin 2 is the controller's status word; it must never overwrite
// the cloud's connection status, or the publisher reports a live
// controller offline.
func TestStatusWordDoesNotClobberCloudStatus(t *testing.T) {
	c := New("e", "p")
	state := &DeviceState{CloudStatus: "ONLINE"}
	c.applyPinValue(state, protocol.PinStatus, "Hot")
	if state.CloudStatus != "ONLINE" {
		t.Fatalf("cloud status = %q, want ONLINE", state.CloudStatus)
	}
	if state.StatusWord != "Hot" {
		t.Fatalf("status word = %q, want Hot", state.StatusWord)
	}
}

func TestPinsLandInTheirVendorFields(t *testing.T) {
	c := New("e", "p")
	state := &DeviceState{}
	for pin, v := range map[int]string{
		protocol.PinFinsTemp:    "56.5",
		protocol.PinHeaterOn:    "1",
		protocol.PinPowerOn:     "1",
		protocol.PinTooHotTemp:  "45",
		protocol.PinTooColdTemp: "38",
		protocol.PinNotifyOn:    "1", // alerts flag: no field, must not land anywhere
	} {
		c.applyPinValue(state, pin, v)
	}
	if state.FrostTemp != 56.5 || state.HeaterOn != 1 || state.PowerOn != 1 ||
		state.TooHotTemp != 45 || state.TooColdTemp != 38 {
		t.Fatalf("state = %+v", *state)
	}
}

func TestPinNames(t *testing.T) {
	for pin, want := range map[int]string{
		protocol.PinPowerOn:  "power_on",
		protocol.PinHeaterOn: "heater_on",
		protocol.PinNotifyOn: "notify_on",
		protocol.PinFinsTemp: "frost_temp",
		33:                   "pin_33",
	} {
		if got := protocol.PinName(pin); got != want {
			t.Errorf("PinName(%d) = %q, want %q", pin, got, want)
		}
	}
}
