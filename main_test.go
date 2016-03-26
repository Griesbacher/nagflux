package main

import (
	"encoding/json"
	"fmt"
	"github.com/griesbacher/nagflux/target/influx"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"
)

const (
	filename = "config.gcfg"
	envInflux = "NAGFLUX_TEST_INFLUX"
	envLivestatus = "NAGFLUX_TEST_LIVESTATUS"
	envSave = "NAGFLUX_TEST_SAVE"
	databaseName = "NAGFLUX_CI_TEST"
	timeout = time.Duration(20) * time.Second
)

type testData struct {
	input  string
	output influx.SeriesValue
}

func testResult(t []testData, result []influx.SeriesValue) bool {
	hits := 0
	for _, testData := range t {
		for _, values := range result {
			if reflect.DeepEqual(values, testData.output) {
				hits++
				break
			}
		}
	}

	return hits == len(t)
}

var NagiosTestData = []testData{
	//Nasty
	{`DATATYPE::SERVICEPERFDATA	TIMET::1	HOSTNAME::h1	SERVICEDESC::s1	SERVICEPERFDATA::C: use=1;2;3;4;5	SERVICECHECKCOMMAND::usage
`,
		// [time command crit crit-fill host max min performanceLabel service value warn warn-fill]
		// [1000 usage 3 none h1 5 <nil> 4 C:\ use s1 1 2 none]
		[]interface{}{1000.0, "usage", 3.0, "none", "h1", 5.0, nil, 4.0, `C: use`, "s1", 1.0, 2.0, "none"}},
	{`DATATYPE::SERVICEPERFDATA	TIMET::1	HOSTNAME::h1	SERVICEDESC::s1	SERVICEPERFDATA::D:\ use=1;2;3;4;5	SERVICECHECKCOMMAND::usage
`,
		// [time command crit crit-fill host max min performanceLabel service value warn warn-fill]
		// [1000 usage 3 none h1 5 <nil> 4 C:\ use s1 1 2 none]
		[]interface{}{1000.0, "usage", 3.0, "none", "h1", 5.0, nil, 4.0, `D:\ use`, "s1", 1.0, 2.0, "none"}},
	//Normal
	{`DATATYPE::SERVICEPERFDATA	TIMET::2	HOSTNAME::h2	SERVICEDESC::s2	SERVICEPERFDATA::rta=2;3;4;5;6	SERVICECHECKCOMMAND::ping
`, //[2000 ping 4 none h2 6 <nil> 5 rta s2 2 3 none]
		[]interface{}{2000.0, "ping", 4.0, "none", "h2", 6.0, nil, 5.0, "rta", "s2", 2.0, 3.0, "none"}},
}

var NagfluxTestData1 = []testData{
	{`table&time&t_host&t_service&t_command&t_performanceLabel&f_value
metrics&10&nagflux&service1&command1&perf&20
`,
		//[10 command1 <nil> <nil> nagflux <nil> <nil> <nil> perf service1 20 <nil> <nil>]
		[]interface{}{10.0, "command1", nil, nil, "nagflux", nil, nil, nil, "perf", "service1", 20.0, nil, nil}},
	{`metrics&20&nagflux&service 1&command1&perf 1&30
`,
		//[10 command1 <nil> <nil> nagflux <nil> <nil> perf service1 20 <nil> <nil>]
		[]interface{}{20.0, "command1", nil, nil, "nagflux", nil,nil, nil, "perf 1", "service 1", 20.0, nil, nil}},
}
var NagfluxTestData2 = []testData{
	{`table&time&t_host&t_service&f_message
messages&100&nagflux&service1&"""Hallo World"""
`,
		//[100 <nil> <nil> <nil> nagflux <nil> Hallo World <nil> <nil> service1 <nil> <nil> <nil>]
		[]interface{}{100.0, nil, nil, nil, "nagflux", nil, "Hallo World", nil, nil, "service1", nil, nil, nil}},
}

var TestDataName = `metrics`
var TestDataColumns = []string{"time", "command", "crit", "crit-fill", "host", "max","message", "min", "performanceLabel", "service", "value", "warn", "warn-fill"}

var OldConfig string
var influxParam string
var livestatusParam string
var save bool
var finished chan bool

func init() {
	finished = make(chan bool)
	influxParam = os.Getenv(envInflux)
	if influxParam == "" {
		influxParam = "http://127.0.0.1:8086"
		fmt.Printf("%s is not set, using default: %s\n", envInflux, influxParam)
	}

	livestatusParam = os.Getenv(envLivestatus)
	if livestatusParam == "" {
		livestatusParam = "127.0.0.1:6557"
		fmt.Printf("%s is not set, using default: %s\n", envLivestatus, livestatusParam)
	}

	if os.Getenv(envSave) == "" {
		save = true
		fmt.Println("Will save the database")
	}
}

func TestEverything(t *testing.T) {
	go createTestData("test/nagios/", "1.txt", NagiosTestData)
	go createTestData("test/nagflux/", "1.txt", NagfluxTestData1)
	go createTestData("test/nagflux/", "2.txt", NagfluxTestData2)
	createConfig()
	dropDatabase()
	go main()
	time.Sleep(time.Duration(1) * time.Second)
	restoreConfig()
	go checkDatabase()
	select {
	case <-finished:
	case <-time.After(timeout):
		result, err := getEverything()
		t.Errorf("Expected data was not found in the influxdb within the timerange: %s\nError: %+v\nDatabase:%+v", timeout, err, result)
	}
	quit <- true
	if !save {
		dropDatabase()
	}
}

func createTestData(folder, file string, data []testData) {
	if err := os.MkdirAll(folder, 0700); err != nil {
		panic(err)
	}
	fileData := []byte{}
	for _, data := range data {
		fileData = append(fileData, []byte(data.input)...)
	}
	if err := ioutil.WriteFile(folder + file, fileData, 0644); err != nil {
		panic(err)
	}
	fmt.Println(string(fileData))
}

func checkDatabase() {
	nagiosResult := false
	nagfluxResult1 := false
	nagfluxResult2 := false
	for {
		time.Sleep(time.Duration(500) * time.Millisecond)
		query, _ := getEverything()
		if len((*query).Results) == 0 {
			continue
		}
		result := (*query).Results[0]
		if len(result.Series) != 2 || TestDataName != result.Series[1].Name || !reflect.DeepEqual(TestDataColumns, result.Series[1].Columns) {
			continue
		}
		fmt.Println(result.Series[0].Values)
		fmt.Println(result.Series[1].Values)
		if !nagiosResult {
			nagiosResult = testResult(NagiosTestData, result.Series[1].Values)
		}
		if !nagfluxResult1 {
			nagfluxResult1 = testResult(NagfluxTestData1, result.Series[1].Values)
		}
		if !nagfluxResult2 {
			nagfluxResult2 = testResult(NagfluxTestData2, result.Series[0].Values)
		}
		fmt.Println(nagiosResult, nagfluxResult1, nagfluxResult2)
		if nagiosResult && nagfluxResult1 && nagfluxResult2 {
			finished <- true
			return
		}
	}
}

func getEverything() (*influx.ShowSeriesResult, error) {
	resp, err := http.Get(influxParam + "/query?db=" + url.QueryEscape(databaseName) + "&q=select%20*%20from%20/.*/&epoch=ms")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var jsonResult influx.ShowSeriesResult
		json.Unmarshal(body, &jsonResult)
		return &jsonResult, nil
	}
	return nil, fmt.Errorf("Database query(%s) returned: %s", resp.Request.URL, resp.Status)
}

func dropDatabase() {
	http.Get(influxParam + "/query?q=drop%20database%20" + url.QueryEscape(databaseName))
}

func createConfig() {
	old, _ := ioutil.ReadFile(filename)
	OldConfig = string(old)
	config := []byte(fmt.Sprintf(`
[main]
	NagiosSpoolfileFolder = "test/nagios"
	NagiosSpoolfileWorker = 1
	InfluxWorker = 2
	MaxInfluxWorker = 5
	DumpFile = "nagflux.dump"
	NagfluxSpoolfileFolder = "test/nagflux"
	FieldSeparator = "&"

[Log]
	LogFile = ""
	MinSeverity = "WARN"

[Monitoring]
	WebserverPort = ""

[Influx]
    	Enabled = true
	Version = 0.9
	Address = "%s"
	Arguments = "precision=ms&db=%s"
	CreateDatabaseIfNotExists = true
	NastyString = ""
	NastyStringToReplace = ""

[Livestatus]
	Type = "tcp"
	Address = "%s"

[Elasticsearch]
	    Enabled = false
	    Address = "http://localhost:9200"
	    Index = "nagflux"
	    Version = 2.1
	`, influxParam, databaseName, livestatusParam))
	if err := ioutil.WriteFile(filename, config, 0644); err != nil {
		panic(err)
	}
}

func restoreConfig() {
	if err := ioutil.WriteFile(filename, []byte(OldConfig), 0644); err != nil {
		panic(err)
	}
}
