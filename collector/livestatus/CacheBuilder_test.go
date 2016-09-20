package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"reflect"
	"testing"
	"time"
)

func TestNewCacheBuilder(t *testing.T) {
	logging.InitTestLogger()
	connector := &Connector{logging.GetLogger(), "localhost:6558", "tcp"}
	builder := NewLivestatusCacheBuilder(connector)
	if builder == nil {
		t.Error("Constructor returned null pointer")
	}
}

func DisabledTestServiceInDowntime(t *testing.T) {
	logging.InitTestLogger()
	queries := map[string]string{}
	queries[QueryForServicesInDowntime] = "1,2;host1;service1\n"
	queries[QueryForHostsInDowntime] = "3,4;host1\n5;host2\n"
	queries[QueryForDowntimeid] = "1;0;1\n2;2;3\n3;0;1\n4;1;2\n5;2;1\n"
	livestatus := &MockLivestatus{"localhost:6558", "tcp", queries, true}
	go livestatus.StartMockLivestatus()
	connector := &Connector{logging.GetLogger(), livestatus.LivestatusAddress, livestatus.ConnectionType}

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
