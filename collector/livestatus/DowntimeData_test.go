package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"testing"
)

func TestSanitizeValuesDowntime(t *testing.T) {
	t.Parallel()
	down := DowntimeData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, endTime: "123"}
	down.sanitizeValues()
	if down.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, down.Data.hostName)
	}
}

func TestPrintDowntime(t *testing.T) {
	t.Parallel()
	logging.InitTestLogger()
	down := DowntimeData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, endTime: "123"}
	if !didThisPanic(down.PrintForInfluxDB, "0.8") {
		t.Errorf("This should panic, due to unsuported influxdb version")
	}

	result := down.PrintForInfluxDB("0.9")
	expected := `messages,host=host\ 1,service=service\ 1,type=downtime,author=philip value="Downtime start: <br>" 000
messages,host=host\ 1,service=service\ 1,type=downtime,author=philip value="Downtime end: <br>" 123000`
	if result != expected {
		t.Errorf("The result did not match the expected. Result: %s Expected %s", result, expected)
	}
}
