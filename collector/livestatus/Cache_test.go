package livestatus

import (
	"reflect"
	"testing"
)

func TestAddDowntime(t *testing.T) {
	cache := Cache{make(map[string]map[string]string)}
	if !reflect.DeepEqual(cache.downtime, make(map[string]map[string]string)) {
		t.Error("Cache should be empty at the beginning.")
	}

	cache.addDowntime("hostname", "servicename", "123")
	intern := map[string]map[string]string{"hostname": map[string]string{"servicename": "123"}}
	if !reflect.DeepEqual(cache.downtime, intern) {
		t.Error("Added element is missing.")
	}

	cache.addDowntime("hostname2", "", "123")
	intern = map[string]map[string]string{"hostname": map[string]string{"servicename": "123"}, "hostname2": map[string]string{"": "123"}}
	if !reflect.DeepEqual(cache.downtime, intern) {
		t.Error("Added element is missing.")
	}

	cache.addDowntime("hostname2", "", "1")
	intern = map[string]map[string]string{"hostname": map[string]string{"servicename": "123"}, "hostname2": map[string]string{"": "123"}}
	if reflect.DeepEqual(cache.downtime, intern) {
		t.Error("Added element is missing.")
	}
}
