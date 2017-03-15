package helper

import (
	"fmt"
	"github.com/griesbacher/nagflux/config"
	"testing"
)

var CreateJSONFromStringMapData = []struct {
	input     map[string]string
	expected  string
	alternate string
}{
	{map[string]string{"a": "1"}, `,"a":1`, ""},
	{map[string]string{"a": "b"}, `,"a":"b"`, ""},
	{map[string]string{"a": "1", "2": "b"}, `,"a":1,2:"b"`, `,2:"b","a":1`},
}

func TestCreateJSONFromStringMap(t *testing.T) {
	t.Parallel()
	for _, data := range CreateJSONFromStringMapData {
		actual := CreateJSONFromStringMap(data.input)
		if !(actual == data.expected || actual == data.alternate) {
			t.Errorf("CreateJSONFromStringMap(%s): expected:%s or %s, actual:%s", data.input, data.expected, data.alternate, actual)
		}
	}
}

var SanitizeElasicInputData = []struct {
	input    string
	expected string
}{
	{"asdf", "asdf"},
	{"'asdf'", "asdf"},
	{"'as df'", "as df"},
	{`'as\ df'`, `as\\ df`},
	{`'as\" df'`, `as\\\" df`},
}

func TestSanitizeElasicInput(t *testing.T) {
	t.Parallel()
	for _, data := range SanitizeElasicInputData {
		actual := SanitizeElasicInput(data.input)
		if actual != data.expected {
			t.Errorf("SanitizeElasicInputData(%s): expected:%s, actual:%s", data.input, data.expected, actual)
		}
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
    BufferSize = 10000
    FileBufferSize = 65536

[Log]
    # leave empty for stdout
    LogFile = ""
    # List of Severities https://godoc.org/github.com/kdar/factorlog#Severity
    MinSeverity = "INFO"

[Monitoring]
    # leave empty to disable
    # PrometheusAddress = ":8080"
    PrometheusAddress = ":8080"

[Livestatus]
    # tcp or file
    Type = "tcp"
    # tcp: 127.0.0.1:6557 or file /var/run/live
    Address = "127.0.0.1:6557"
    # The amount to minutes to wait for livestatus to come up, if set to 0 the detection is disabled
    MinutesToWait = 2

[ModGearman "example"] #copy this block and rename it to add a second ModGearman queue
    Enabled = false
    Address = "127.0.0.1:4730"
    Queue = "perfdata"
    # Leave Secret and SecretFile empty to disable encryption
    # If both are filled the the Secret will be used
    # Secret to encrypt the gearman jobs
    Secret = ""
    # Path to a file which holds the secret to encrypt the gearman jobs
    SecretFile = "/etc/mod-gearman/secret.key"
    Worker = 1

[InfluxDBGlobal]
	CreateDatabaseIfNotExists = true
	NastyString = ""
	NastyStringToReplace = ""
	HostcheckAlias = "hostcheck"

[InfluxDB "nagflux"]
    Enabled = true
	Version = 1.0
	Address = "http://127.0.0.1:8086"
	Arguments = "precision=ms&u=root&p=root&db=nagflux"
	StopPullingDataIfDown = true

[InfluxDB "fast"]
    Enabled = true
	Version = 1.0
	Address = "http://127.0.0.1:8086"
	Arguments = "precision=ms&u=root&p=root&db=fast"
	StopPullingDataIfDown = false

[ElasticsearchGlobal]
    HostcheckAlias = "hostcheck"
    NumberOfShards = 1
    NumberOfReplicas = 1
    # Sorts the indices "monthly" or "yearly"
    IndexRotation = "%s"

[Elasticsearch "example"]
    Enabled = false
    Address = "http://localhost:9200"
    Index = "nagflux"
    Version = 2.1`

func TestGenIndex(t *testing.T) {
	config.InitConfigFromString(fmt.Sprintf(Config, "monthly"))
	//Do 24. Mär 15:00:44 CET 2016 == 1458828043
	result := GenIndex("index", "1458828043000")
	expected := "index-2016.03"
	if result != expected {
		t.Errorf(`GenIndex("index","1458828043000"): expected:%s, actual:%s`, expected, result)
	}
	config.InitConfigFromString(fmt.Sprintf(Config, "yearly"))
	result = GenIndex("index", "1458828043000")
	expected = "index-2016"
	if result != expected {
		t.Errorf(`GenIndex("index","1458828043000"): expected:%s, actual:%s`, expected, result)
	}
	config.InitConfigFromString(fmt.Sprintf(Config, "foo"))
	if !didThisPanic(GenIndex, "index", "1458828043000") {
		t.Error("The Config was invalid but did not panic!")
	}

}

func didThisPanic(f func(string, string) string, arg1, arg2 string) (result bool) {
	defer func() {
		if rec := recover(); rec != nil {
			result = true
		}
	}()
	f(arg1, arg2)
	return false
}
