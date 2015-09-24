package influx

import (
	"bytes"
	"errors"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/statistics"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

//Reads data from the queue and sends them to the influxdb.
type InfluxWorker struct {
	workerId     int
	quit         chan bool
	quitInternal chan bool
	jobs         chan interface{}
	connection   string
	dumpFile     string
	statistics   statistics.DataReceiver
	log          *factorlog.FactorLog
	version      float32
	connector    *InfluxConnector
	httpClient   http.Client
	IsRunning    bool
}

var errorInterrupted = errors.New("Got interrupted")
var errorBadRequest = errors.New("400 Bad Request")
var errorHttpClient = errors.New("Http Client got an error")
var errorFailedToSend = errors.New("Could not send data")
var error500 = errors.New("Error 500")

//Generates a new Worker and starts it.
func InfluxWorkerGenerator(jobs chan interface{}, connection, dumpFile string, version float32, connector *InfluxConnector) func(workerId int) *InfluxWorker {
	return func(workerId int) *InfluxWorker {
		worker := &InfluxWorker{
			workerId, make(chan bool),
			make(chan bool, 1), jobs,
			connection, dumpFile,
			statistics.NewCmdStatisticReceiver(),
			logging.GetLogger(), version,
			connector, http.Client{}, true}
		go worker.run()
		return worker
	}
}

//Stops the worker
func (worker *InfluxWorker) Stop() {
	worker.quitInternal <- true
	worker.quit <- true
	<-worker.quit
	worker.IsRunning = false
	worker.log.Debug("InfluxWorker stopped")
}

//Tries to send data all the time.
func (worker InfluxWorker) run() {
	var queries []interface{}
	var query interface{}
	for {
		if worker.connector.IsAlive() {
			if worker.connector.DatabaseExists() {
				select {
				case <-worker.quit:
					worker.log.Debug("InfluxWorker quitting...")
					worker.sendBuffer(queries)
					worker.quit <- true
					return
				case query = <-worker.jobs:
					queries = append(queries, query)
					if len(queries) == 300 {
						worker.sendBuffer(queries)
						queries = queries[:0]
					}
				case <-time.After(time.Duration(30) * time.Second):
					worker.sendBuffer(queries)
					queries = queries[:0]
				}
			} else {
				//Test Database
				worker.connector.TestDatabaseExists()
				if worker.waitForExternalQuit() {
					return
				}
			}
		} else {
			//Test Influxdb
			worker.connector.TestIfIsAlive()
			if worker.waitForExternalQuit() {
				return
			}
		}
	}
}

//Checks if a external quit signal arrives.
func (worker InfluxWorker) waitForExternalQuit() bool {
	select {
	case <-worker.quit:
		worker.quit <- true
		return true
	case <-time.After(time.Duration(30) * time.Second):
		return false
	}
}

//Sends the given queries to the influxdb.
func (worker InfluxWorker) sendBuffer(queries []interface{}) {
	if len(queries) == 0 {
		return
	}

	var lineQueries []string
	for _, query := range queries {
		cast, castErr := worker.castJobToString(query)
		if castErr == nil {
			lineQueries = append(lineQueries, cast)
		}
	}

	var dataToSend []byte
	for _, lineQuery := range lineQueries {
		dataToSend = append(dataToSend, []byte(lineQuery)...)
	}

	startTime := time.Now()
	sendErr := worker.sendData([]byte(dataToSend), true)
	if sendErr != nil {
		for i := 0; i < 3; i++ {
			switch sendErr {
			case errorBadRequest:
				//Maybe just a few queries are wrong, so send them one by one and find the bad one
				var badQueries []string
				for _, lineQuery := range lineQueries {
					queryErr := worker.sendData([]byte(lineQuery), false)
					if queryErr != nil {
						badQueries = append(badQueries, lineQuery)
					}
				}
				worker.dumpErrorQueries("\n\nOne of the values is not clean..\n", badQueries)
				sendErr = nil
			case nil:
				//Single point of exit
				break
			default:
				if err := worker.waitForQuitOrGoOn(); err != nil {
					//No error handling, because it's time to terminate
					worker.dumpRemainingQueries(lineQueries)
					sendErr = nil
				}
				//Resend Data
				sendErr = worker.sendData([]byte(dataToSend), true)
			}
		}
		if sendErr != nil {
			//if there is still an error dump the queries and go on
			worker.dumpErrorQueries("\n\n"+sendErr.Error()+"\n", lineQueries)
		}

	}
	worker.statistics.ReceiveQueries("send", statistics.QueriesPerTime{len(lineQueries), time.Since(startTime)})
}

//Writes the bad queries to a dumpfile.
func (worker InfluxWorker) dumpErrorQueries(messageForLog string, errorQueries []string) {
	errorFile := worker.dumpFile + "-errors"
	worker.log.Warnf("Dumping queries with errors to: %s", errorFile)
	errorQueries = append([]string{messageForLog}, errorQueries...)
	worker.dumpQueries(errorFile, errorQueries)
}

var mutex = &sync.Mutex{}

//Dumps the remaining queries if a quit signal arises.
func (worker InfluxWorker) dumpRemainingQueries(remainingQueries []string) {
	mutex.Lock()
	worker.log.Debugf("Global queue %d own queue %d", len(worker.jobs), len(remainingQueries))
	if len(worker.jobs) != 0 || len(remainingQueries) != 0 {
		worker.log.Debug("Saving queries to disk")

		remainingQueries = append(remainingQueries, worker.readQueriesFromQueue()...)

		worker.log.Debugf("dumping %d queries", len(remainingQueries))
		worker.dumpQueries(worker.dumpFile, remainingQueries)
	}
	mutex.Unlock()
}

//Reads the queries from the global queue and returns them as string.
func (worker InfluxWorker) readQueriesFromQueue() []string {
	var queries []string
	var query interface{}
	stop := false
	for !stop {
		select {
		case query = <-worker.jobs:
			cast, err := worker.castJobToString(query)
			if err == nil {
				queries = append(queries, cast)
			}
		case <-time.After(time.Duration(200) * time.Millisecond):
			stop = true
		}
	}
	return queries
}

//sends the raw data to influxdb and returns an err if given.
func (worker InfluxWorker) sendData(rawData []byte, log bool) error {
	req, err := http.NewRequest("POST", worker.connection, bytes.NewBuffer(rawData))
	if err != nil {
		worker.log.Warn(err)
	}
	resp, err := worker.httpClient.Do(req)
	if err != nil {
		worker.log.Warn(err)
		return errorHttpClient
	} else {
		defer resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			//OK
			return nil
		} else if resp.StatusCode == 500 {
			//Temporarily timeout
			if log {
				worker.logHttpResponse(resp)
			}
			return error500
		} else if resp.StatusCode == 400 {
			//Bad Request
			if log {
				worker.logHttpResponse(resp)
			}
			return errorBadRequest
		} else {
			//HTTP Error
			if log {
				worker.logHttpResponse(resp)
			}
			return errorFailedToSend
		}
	}
}

//Logs a http response to warn.
func (worker InfluxWorker) logHttpResponse(resp *http.Response) {
	body, _ := ioutil.ReadAll(resp.Body)
	worker.log.Warnf("Influx status: %s - %s", resp.Status, string(body))
}

//Waits on an internal quit signal.
func (worker InfluxWorker) waitForQuitOrGoOn() error {
	select {
	//Got stop signal
	case <-worker.quitInternal:
		worker.log.Debug("Recived quit")
		worker.quitInternal <- true
		return errorInterrupted
	//Timeout and retry
	case <-time.After(time.Duration(10) * time.Second):
		return nil
	}
}

//Writes queries to a dumpfile.
func (worker InfluxWorker) dumpQueries(filename string, queries []string) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		if _, err := os.Create(filename); err != nil {
			worker.log.Critical(err)
		}
	}
	if f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY, 0600); err != nil {
		worker.log.Critical(err)
	} else {
		defer f.Close()
		for _, query := range queries {
			if _, err = f.WriteString(query); err != nil {
				worker.log.Critical(err)
			}
		}
	}
}

//Converts an interface{} to a string.
func (worker InfluxWorker) castJobToString(job interface{}) (string, error) {
	var result string
	var err error
	switch jobCast := job.(type) {
	case spoolfile.PerformanceData:
		if worker.version >= 0.9 {
			result = jobCast.String()
		} else {
			worker.log.Fatalf("This influxversion [%f] given in the config is not supportet", worker.version)
			err = errors.New("This influxversion given in the config is not supportet")
		}
	case string:
		result = jobCast
	case livestatus.Printable:
		result = jobCast.Print(worker.version)
	default:
		worker.log.Fatal("Could not cast object:", job)
		err = errors.New("Could not cast object")
	}
	if len(result) > 1 && result[len(result)-1:] != "\n" {
		result += "\n"
	}
	return result, err
}
