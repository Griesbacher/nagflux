package livestatus

import (
	"testing"
	"github.com/griesbacher/nagflux/logging"
)

func TestSanitizeValuesDowntime(t *testing.T) {
	t.Parallel()
	down := DowntimeData{Data:Data{fieldSeperator:"&", hostName:"host 1", serviceDisplayName:"service 1", author:"philip"}, endTime:"123"}
	down.sanitizeValues()
	if down.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, down.Data.hostName)
	}
}

func TestPrintDowntime(t *testing.T) {
	t.Parallel()
	logging.InitTestLogger()
	down := DowntimeData{Data:Data{fieldSeperator:"&", hostName:"host 1", serviceDisplayName:"service 1", author:"philip"}, endTime:"123"}
	if !didThisPanic(down.Print, 0.8) {
		t.Errorf("This should panic, due to unsuported influxdb version")
	}

	result := down.Print(0.9)
	expected := `host\ 1&service\ 1&messages,type=downtime,author=philip value="Downtime start: <br>" 000
host\ 1&service\ 1&messages,type=downtime,author=philip value="Downtime end: <br>" 123000`
	if result != expected {
		t.Errorf("The result did not match the expected. Result: %s Expected %s", result, expected)
	}
}
