package main

import (
	"testing"

	"github.com/datariot/coolbot-client/internal/client"
	"github.com/datariot/coolbot-client/internal/protocol"
)

// KB-1163: at connect the profile carries no room or fins reading; the
// first interval must not publish them as 0 °F.
func TestUnheardPinsAreNotPublished(t *testing.T) {
	dev := client.DeviceState{
		CloudStatus: "ONLINE",
		Pins:        map[int]string{protocol.PinSetPoint: "41.998", protocol.PinHeaterOn: "1"},
	}
	got := periodicReadings(dev)
	want := map[string]string{"online": "1", "set_point": "41.998", "heater_on": "1"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for k, v := range want {
		if got[k] != v {
			t.Fatalf("%s = %q, want %q (all: %v)", k, got[k], v, got)
		}
	}
}

func TestOfflineIsOnline0(t *testing.T) {
	got := periodicReadings(client.DeviceState{CloudStatus: "OFFLINE"})
	if got["online"] != "0" {
		t.Fatalf("online = %q", got["online"])
	}
}
