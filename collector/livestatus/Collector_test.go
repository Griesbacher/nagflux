package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/logging"
	"testing"
	"time"
)

func TestNewLivestatusCollector(t *testing.T) {
	livestatus := &MockLivestatus{"localhost:6559", "tcp", map[string]string{}, true}
	go livestatus.StartMockLivestatus()
	connector := &Connector{logging.GetLogger(), "localhost:6559", "tcp"}
	collector := NewLivestatusCollector(make(collector.ResultQueues), connector, false)
	if collector == nil {
		t.Error("Constructor returned null pointer")
	}
	collector.Stop()
}

func TestAddTimestampToLivestatusQuery(t *testing.T) {
	if addTimestampToLivestatusQuery(QueryIcinga2ForNotifications) != fmt.Sprintf(QueryIcinga2ForNotifications, time.Now().Add(intervalToCheckLivestatus/100*-150).Unix()) {
		t.Error("addTimestampToLivestatusQuery has changed")
	}
}
