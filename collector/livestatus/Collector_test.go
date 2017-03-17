package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/logging"
	"testing"
	"time"
)

func TestNewLivestatusCollector(t *testing.T) {
	livestatus := &MockLivestatus{
		LivestatusAddress:"localhost:6559",
		ConnectionType:"tcp",
		Queries:map[string]string{},
		isRunning: true,
	}
	go livestatus.StartMockLivestatus()
	connector := &Connector{
		Log:logging.GetLogger(),
		LivestatusAddress:"localhost:6559",
		ConnectionType:"tcp",
	}
	collector := NewLivestatusCollector(make(collector.ResultQueues), connector, "")
	if collector == nil {
		t.Error("Constructor returned null pointer")
	}
	collector.Stop()
}

func TestAddTimestampToLivestatusQuery(t *testing.T) {
	if addTimestampToLivestatusQuery(QueryIcinga2ForNotifications) != fmt.Sprintf(QueryIcinga2ForNotifications, time.Now().Add(intervalToCheckLivestatus / 100 * -150).Unix()) {
		t.Error("addTimestampToLivestatusQuery has changed")
	}
}
