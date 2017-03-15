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
	filename      = "test.gcfg"
	envInflux     = "NAGFLUX_TEST_INFLUX"
	envLivestatus = "NAGFLUX_TEST_LIVESTATUS"
	envSave       = "NAGFLUX_TEST_SAVE"
	databaseName  = "NAGFLUX_CI_TEST"
	timeout       = time.Duration(20) * time.Second
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

var NagiosTestData1 = []testData{
	//Nasty
	{`DATATYPE::SERVICEPERFDATA	TIMET::1	HOSTNAME::h1	SERVICEDESC::s1	SERVICEPERFDATA::C: use=1;2;3;4;5	SERVICECHECKCOMMAND::usage
`,
		// [time command crit crit-fill host max min performanceLabel service value warn warn-fill]
		// [1000 usage 3 none h1 5 <nil> 4 C:\ use s1 1 2 none]
		[]interface{}{1000.0, "usage", 3.0, "none", "h1", 5.0, nil, 4.0, `C: use`, "s1", 1.0, 2.0, "none"}},
	{`DATATYPE::SERVICEPERFDATA	TIMET::3	HOSTNAME::h1	SERVICEDESC::s1	SERVICEPERFDATA::D:\ use=1;2;3;4;5	SERVICECHECKCOMMAND::usage
`,
		// [time command crit crit-fill host max min performanceLabel service value warn warn-fill]
		// [3000 usage 3 none h1 5 <nil> 4 D:\ use s1 1 2 none]
		[]interface{}{3000.0, "usage", 3.0, "none", "h1", 5.0, nil, 4.0, `D:\ use`, "s1", 1.0, 2.0, "none"}},
	//Normal
	{`DATATYPE::SERVICEPERFDATA	TIMET::2	HOSTNAME::h2	SERVICEDESC::s2	SERVICEPERFDATA::rta=2;3;4;5;6	SERVICECHECKCOMMAND::ping
`, //[2000 ping 4 none h2 6 <nil> 5 rta s2 2 3 none]
		[]interface{}{2000.0, "ping", 4.0, "none", "h2", 6.0, nil, 5.0, "rta", "s2", 2.0, 3.0, "none"}},
}
var NagiosTestData21 = []testData{
	//Database1
	{`DATATYPE::SERVICEPERFDATA	TIMET::4	HOSTNAME::h3	SERVICEDESC::s1	SERVICEPERFDATA::rta=2;3;4;5;6	SERVICECHECKCOMMAND::special	NAGFLUX:TARGET::` + database1 + `
`, //[2000 ping 4 none h2 6 <nil> 5 rta s2 2 3 none]
		[]interface{}{4000.0, "special", 4.0, "none", "h3", 6.0, nil, 5.0, "rta", "s1", 2.0, 3.0, "none"}},
}
var NagiosTestData22 = []testData{
	//Database2
	{`DATATYPE::SERVICEPERFDATA	TIMET::4	HOSTNAME::h3	SERVICEDESC::s2	SERVICEPERFDATA::rta=2;3;4;5;6	SERVICECHECKCOMMAND::special	NAGFLUX:TARGET::` + database2 + `
`, //[2000 ping 4 none h2 6 <nil> 5 rta s2 2 3 none]
		[]interface{}{4000.0, "special", 4.0, "none", "h3", 6.0, nil, 5.0, "rta", "s2", 2.0, 3.0, "none"}},
}

var NagfluxTestData1 = []testData{
	{`table&time&t_host&t_service&t_command&t_performanceLabel&f_value
metrics&10&nagflux&service1&command1&perf&20
`,
		//[10 command1 <nil> <nil> nagflux <nil> <nil> <nil> perf service1 20 <nil> <nil>]
		[]interface{}{10.0, "command1", nil, nil, "nagflux", nil, nil, nil, "perf", "service1", 20.0, nil, nil}},
	{`metrics&20&nagflux&service\ 1&command1&perf\ 1&30
`,
		//[20 command1 <nil> <nil> nagflux <nil> <nil> <nil> perf\ 1 service\ 1 30 <nil> <nil>]
		[]interface{}{20.0, "command1", nil, nil, "nagflux", nil, nil, nil, "perf 1", "service 1", 30.0, nil, nil}},
}

var NagfluxTestData2 = []testData{
	{`table&time&t_host&t_service&f_message
messages&100&nagflux&service1&"""Hallo World"""
`,
		//[100 <nil> <nil> <nil> nagflux <nil> Hallo World <nil> <nil> service1 <nil> <nil> <nil>]
		[]interface{}{100.0, nil, nil, nil, "nagflux", nil, "Hallo World", nil, nil, "service1", nil, nil, nil}},
	{`messages&300&nagflux&service1&"""Hallo \\"""
`,
		//[300 <nil> <nil> <nil> nagflux <nil> Hallo \ <nil> <nil> service1 <nil> <nil> <nil>]
		[]interface{}{300.0, nil, nil, nil, "nagflux", nil, `Hallo \`, nil, nil, "service1", nil, nil, nil}},
}

var TestDataName = `metrics`
var TestDataColumns = []string{"time", "command", "crit", "crit-fill", "host", "max", "message", "min", "performanceLabel", "service", "value", "warn", "warn-fill"}
var influxParam string
var livestatusParam string
var save bool
var finished chan bool

var database1 = databaseName + "_1"
var database2 = databaseName + "_2"

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
		save = false
		fmt.Println("Will NOT save the database")
	} else {
		save = true
		fmt.Println("Will save the database")
	}
}

func TestEverything(t *testing.T) {
	os.Args = append(os.Args, "-configPath="+filename)
	go createTestData("test/nagios/", "1.txt", NagiosTestData1)
	go createTestData("test/nagios/", "2.txt", NagiosTestData21)
	go createTestData("test/nagios/", "3.txt", NagiosTestData22)
	go createTestData("test/nagflux/", "1.txt", NagfluxTestData1)
	go createTestData("test/nagflux/", "2.txt", NagfluxTestData2)
	createConfig()
	dropDatabase()
	go main()
	time.Sleep(time.Duration(1) * time.Second)
	os.Remove(filename)
	go checkDatabase()
	select {
	case <-finished:
	case <-time.After(timeout):
		result1, err1 := getEverything(database1)
		result2, err2 := getEverything(database2)
		dropDatabase()
		t.Errorf(
			"Expected data was not found in the influxdb within the timerange: %s\nError: %+v\nDatabase:%+v\nError: %+v\nDatabase:%+v",
			timeout, err1, result1, err2, result2,
		)
	}
	select {
	case quit <- true:
	case <-time.After(time.Duration(2) * time.Second):
		fmt.Println("Killing the test due to timeout")
	}
	dropDatabase()
}

func createTestData(folder, file string, data []testData) {
	if err := os.MkdirAll(folder, 0700); err != nil {
		panic(err)
	}
	fileData := []byte{}
	for _, data := range data {
		fileData = append(fileData, []byte(data.input)...)
	}
	if err := ioutil.WriteFile(folder+file, fileData, 0644); err != nil {
		panic(err)
	}
	fmt.Println(string(fileData))
}

func checkDatabase() {
	result := map[string]bool{
		"nagiosResult1":  false,
		"nagiosResult21": false,
		"nagiosResult22": false,
		"nagfluxResult1": false,
		"nagfluxResult2": false,
	}
	for {
		time.Sleep(time.Duration(500) * time.Millisecond)

		query, _ := getEverything(database1)
		if query == nil || len((*query).Results) == 0 {
			continue
		}
		result1 := (*query).Results[0]
		if len(result1.Series) != 2 || TestDataName != result1.Series[1].Name || !reflect.DeepEqual(TestDataColumns, result1.Series[1].Columns) {
			continue
		}

		query2, _ := getEverything(database2)
		if query2 == nil || len((*query2).Results) == 0 {
			continue
		}
		result2 := (*query2).Results[0]
		if len(result2.Series) != 2 || TestDataName != result2.Series[1].Name || !reflect.DeepEqual(TestDataColumns, result2.Series[1].Columns) {
			continue
		}
		if !result["nagiosResult1"] {
			result["nagiosResult1"] = testResult(NagiosTestData1, result1.Series[1].Values)
		}
		if !result["nagiosResult21"] {
			result["nagiosResult21"] = testResult(NagiosTestData21, result1.Series[1].Values)
		}
		if !result["nagiosResult22"] {
			result["nagiosResult22"] = testResult(NagiosTestData22, result2.Series[1].Values)
		}
		if !result["nagfluxResult1"] {
			result["nagfluxResult1"] = testResult(NagfluxTestData1, result1.Series[1].Values)
		}
		if !result["nagfluxResult2"] {
			result["nagfluxResult2"] = testResult(NagfluxTestData2, result1.Series[0].Values)
		}

		completed := true
		for k, v := range result {
			if !v {
				fmt.Println(k + "not forfilled")
				completed = false
			}
		}
		if completed {
			finished <- true
			return
		}
	}
}

func getEverything(database string) (*influx.ShowSeriesResult, error) {
	resp, err := http.Get(influxParam + "/query?db=" + url.QueryEscape(database) + "&q=select%20*%20from%20/.*/&epoch=ms&u=omdadmin&p=omd")
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
	if !save {
		http.Get(influxParam + "/query?u=omdadmin&p=omd&q=drop%20database%20" + url.QueryEscape(database1))
		http.Get(influxParam + "/query?u=omdadmin&p=omd&q=drop%20database%20" + url.QueryEscape(database2))
	}
}

func createConfig() {
	config := []byte(fmt.Sprintf(`
[main]
	NagiosSpoolfileFolder = "test/nagios"
	NagiosSpoolfileWorker = 1
	InfluxWorker = 2
	MaxInfluxWorker = 5
	DumpFile = "test/nagflux.dump"
	NagfluxSpoolfileFolder = "test/nagflux"
	FieldSeparator = "&"
	FileBufferSize = 65536
	BufferSize = 10000

[Log]
	LogFile = ""
	MinSeverity = "DEBUG"

[Monitoring]
	PrometheusAddress = ""

[InfluxDBGlobal]
	CreateDatabaseIfNotExists = true
	NastyString = ""
	NastyStringToReplace = ""
	HostcheckAlias = "hostcheck"

[InfluxDB "%s"]
	Enabled = true
	Version = 1.0
	Address = "%s"
	Arguments = "precision=ms&db=%s&u=omdadmin&p=omd"
	StopPullingDataIfDown = true

[InfluxDB "%s"]
	Enabled = true
	Version = 1.0
	Address = "%s"
	Arguments = "precision=ms&db=%s&u=omdadmin&p=omd"
	StopPullingDataIfDown = true

[InfluxDB "broken"]
	Enabled = true
	Version = 1.0
	Address = "http://127.0.0.1:1"
	Arguments = "precision=ms&db=broken&u=omdadmin&p=123456"
	StopPullingDataIfDown = false

[Livestatus]
	Type = "tcp"
	Address = "%s"
	MinutesToWait = 0

`, database1, influxParam, database1, database2, influxParam, database2, livestatusParam))
	if err := ioutil.WriteFile(filename, config, 0644); err != nil {
		panic(err)
	}
}
