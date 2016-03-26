package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"testing"
	"github.com/griesbacher/nagflux/config"
	"fmt"
)

func TestSanitizeValuesNotification(t *testing.T) {
	t.Parallel()
	notification := NotificationData{Data: Data{hostName: "host 1"}, notificationType: "HOST NOTIFICATION", notificationLevel: "WARN"}
	notification.sanitizeValues()

	if notification.notificationType != `HOST\ NOTIFICATION` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `HOST\ NOTIFICATION`, notification.notificationType)
	}
}

func TestPrintNotification(t *testing.T) {
	logging.InitTestLogger()
	notification := NotificationData{Data: Data{hostName: "host 1", author: "philip"}, notificationType: "HOST NOTIFICATION", notificationLevel: "WARN"}
	if !didThisPanic(notification.PrintForInfluxDB, "0.8") {
		t.Error("Printed for unsuported influxdb version but got a response")
	}

	result := notification.PrintForInfluxDB("0.9")
	if result != `messages,host=host\ 1,service=hostcheck,type=host_notification,author=philip message="WARN:<br> " 000` {
		t.Errorf("Result does not match the expected. Result: %s", result)
	}

	notification2 := NotificationData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, notificationType: "SERVICE NOTIFICATION", notificationLevel: "WARN"}
	result2 := notification2.PrintForInfluxDB("0.9")
	if result2 != `messages,host=host\ 1,service=service\ 1,type=service_notification,author=philip message="WARN:<br> " 000` {
		t.Errorf("Result does not match the expected. Result: %s", result2)
	}

	notification3 := NotificationData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, notificationType: "NULL NOTIFICATION", notificationLevel: "WARN"}
	result3 := notification3.PrintForInfluxDB("0.9")
	if result3 != `messages,host=host\ 1,service=service\ 1,author=philip message="WARN:<br> " 000` {
		t.Errorf("Result does not match the expected. Result: %s", result3)
	}
}

const Config = `[main]
    NagiosSpoolfileFolder = "/var/spool/nagios"
    NagiosSpoolfileWorker = 1
    InfluxWorker = 2
    MaxInfluxWorker = 5
    DumpFile = "nagflux.dump"
    NagfluxSpoolfileFolder = "/var/spool/nagflux"
    FieldSeparator = "&"

[Log]
    # leave empty for stdout
    LogFile = ""
    # List of Severities https://godoc.org/github.com/kdar/factorlog#Severity
    MinSeverity = "INFO"

[Monitoring]
    # leave empty to disable
    # WebserverPort = ":7000"
    WebserverPort = ""

[Influx]
    Enabled = true
    Version = 0.9
    Address = "http://127.0.0.1:8086"
    Arguments = "precision=ms&u=root&p=root&db=nagflux"
    CreateDatabaseIfNotExists = true
    # leave empty to disable
    NastyString = ""
    NastyStringToReplace = ""
    HostcheckAlias = "hostcheck"

[Livestatus]
    # tcp or file
    Type = "tcp"
    # tcp: 127.0.0.1:6557 or file /var/run/live
    Address = "127.0.0.1:6557"

[Elasticsearch]
    Enabled = false
    Address = "http://localhost:9200"
    Index = "nagflux"
    Version = 2.1
    HostcheckAlias = "hostcheck"
    NumberOfShards = 1
    NumberOfReplicas = 1
    # Sorts the indices "monthly" or "yearly"
    IndexRotation = "%s"`

func TestPrintForElasticsearchNotification(t *testing.T) {
	logging.InitTestLogger()
	config.InitConfigFromString(fmt.Sprintf(Config, "monthly"))
	notification := NotificationData{Data: Data{hostName: "host 1", author: "philip", entryTime:"1458988932000"}, notificationType: "HOST NOTIFICATION", notificationLevel: "WARN"}
	if !didThatPanic(notification.PrintForElasticsearch, "1.0", "index") {
		t.Error("Printed for unsuported elasticsearch version but got a response")
	}

	result := notification.PrintForElasticsearch("2.0", "index")
	expected := `{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"WARN:<br> ","author":"philip","host":"host 1","service":"hostcheck","type":"host_notification"}
`
	if result != expected {
		t.Errorf("Result does not match the expected.\n%s%s", result, expected)
	}

	notification2 := NotificationData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", entryTime:"1458988932000"}, notificationType: "SERVICE NOTIFICATION", notificationLevel: "WARN"}
	result2 := notification2.PrintForElasticsearch("2.0", "index")
	expected2 := `{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"WARN:<br> ","author":"philip","host":"host 1","service":"service 1","type":"service_notification"}
`
	if result2 != expected2 {
		t.Errorf("Result does not match the expected.\n%s%s", result2, expected2)
	}

	notification3 := NotificationData{Data: Data{hostName: "host 1", serviceDisplayName: "service 1", author: "philip", entryTime:"1458988932000"}, notificationType: "NULL NOTIFICATION", notificationLevel: "WARN"}
	result3 := notification3.PrintForElasticsearch("2.0", "index")
	expected3 := `{"index":{"_index":"index-2016.03","_type":"messages"}}
{"timestamp":1458988932000000,"message":"WARN:<br> ","author":"philip","host":"host 1","service":"service 1","type":""}
`
	if result3 != expected3 {
		t.Errorf("Result does not match the expected.\n%s%s", result3, expected3)
	}
}

func didThisPanic(f func(string) string, arg string) (result bool) {
	defer func() {
		if rec := recover(); rec != nil {
			result = true
		}
	}()
	f(arg)
	return false
}

func didThatPanic(f func(string, string) string, arg1, arg2 string) (result bool) {
	defer func() {
		if rec := recover(); rec != nil {
			result = true
		}
	}()
	f(arg1, arg2)
	return false
}
