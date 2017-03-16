package config

import (
	"github.com/griesbacher/nagflux/data"
	"testing"
)

func TestStoreValue(t *testing.T) {
	pauseNagflux = PauseMap{}
	target := data.Target{Name: "foo", Datatype: data.InfluxDB}
	StoreValue(target, false)
	if len(pauseNagflux) != 1 {
		t.Error("Map size does not match")
	}
	if IsAnyTargetOnPause() {
		t.Error("No target should be at pause")
	}
	target2 := data.Target{Name: "bar", Datatype: data.InfluxDB}
	StoreValue(target2, true)
	if len(pauseNagflux) != 2 {
		t.Error("Map size does not match")
	}
	if !IsAnyTargetOnPause() {
		t.Error("One target should be at pause")
	}
}
