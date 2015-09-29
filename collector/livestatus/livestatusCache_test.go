package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"reflect"
	"testing"
	"time"
)

func TestNewLivestatusCacheBuilder(t *testing.T) {
	connector := &LivestatusConnector{logging.GetLogger(), "localhost:6558", "tcp"}
	if NewLivestatusCacheBuilder(connector) == nil {
		t.Error("Constructor returned null pointer")
	}
}

func TestAddDowntime(t *testing.T) {
	cache := LivestatusCache{make(map[string]map[string]string)}
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
}

func TestServiceInDowntime(t *testing.T) {
	queries := map[string]string{}
	queries[QueryForServicesInDowntime] = "1,2;host1;service1\n"
	queries[QueryForHostsInDowntime] = "3,4;host1\n5;host2\n"
	queries[QueryForDowntimeid] = "1;0;1\n2;2;3\n3;0;1\n4;1;2\n5;2;1\n"
	livestatus := &MockLivestatus{"localhost:6558", "tcp", queries, true}
	go livestatus.StartMockLivestatus()
	connector := &LivestatusConnector{logging.GetLogger(), livestatus.LivestatusAddress, livestatus.ConnectionType}

	cacheBuilder := NewLivestatusCacheBuilder(connector)

	time.Sleep(time.Duration(2) * time.Second)

	cacheBuilder.Stop()
	livestatus.StopMockLivestatus()

	intern := map[string]map[string]string{"host1": map[string]string{"": "1", "service1": "1"}, "host2": map[string]string{"": "2"}}
	cacheBuilder.mutex.Lock()
	if !reflect.DeepEqual(cacheBuilder.downtimeCache.downtime, intern) {
		t.Errorf("Internall Cache does not fit.\nExpexted:%s\nResult:%s\n", intern, cacheBuilder.downtimeCache.downtime)
	}
	cacheBuilder.mutex.Unlock()
	if !cacheBuilder.IsServiceInDowntime("host1", "service1", "1") {
		t.Errorf(`"host1","service1","1" should be in downtime`)
	}
	if !cacheBuilder.IsServiceInDowntime("host1", "service1", "2") {
		t.Errorf(`"host1","service1","2" should be in downtime`)
	}
	if cacheBuilder.IsServiceInDowntime("host1", "service1", "0") {
		t.Errorf(`"host1","service1","0" should NOT be in downtime`)
	}
	if cacheBuilder.IsServiceInDowntime("host1", "", "0") {
		t.Errorf(`"host1","","0" should NOT be in downtime`)
	}
	if !cacheBuilder.IsServiceInDowntime("host1", "", "2") {
		t.Errorf(`"host1","","2" should be in downtime`)
	}

}
