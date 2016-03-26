package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"testing"
	"github.com/griesbacher/nagflux/config"
	"fmt"
)

func TestSanitizeValuesDowntime(t *testing.T) {
	t.Parallel()
	down := DowntimeData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, endTime: "123"}
	down.sanitizeValues()
	if down.Data.hostName != `host\ 1` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `host\ 1`, down.Data.hostName)
	}
}

func TestPrintInfluxdbDowntime(t *testing.T) {
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

func TestPrintElasticsearchDowntime(t *testing.T) {
	logging.InitTestLogger()
	config.InitConfigFromString(fmt.Sprintf(Config, "monthly"))
	down := DowntimeData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", entryTime:"1458988932000"}, endTime: "123"}
	if !didThatPanic(down.PrintForElasticsearch, "1.0", "index") {
		t.Errorf("This should panic, due to unsuported elasticsearch version")
	}

	result := down.PrintForElasticsearch("2.0", "index")
	expected := `{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"Downtime start: <br>","author":"philip","host":"host 1","service":"service 1","type":"downtime"}

{"index":{"_index":"index-1970.01","_type":"messages"}}
{"timestamp":123000,"message":"Downtime end: <br>","author":"philip","host":"host 1","service":"service 1","type":"downtime"}
`
	if result != expected {
		t.Errorf("The result did not match the expected. Result: %sExpected: %s", result, expected)
	}
}