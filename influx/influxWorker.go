package influx

import (
	"bytes"
	"errors"
	"github.com/kdar/factorlog"
	"griesbacher.org/nagflux/collector"
	"griesbacher.org/nagflux/logging"
	"griesbacher.org/nagflux/statistics"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"
)

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
var errorFailedToSend = errors.New("Could not send data")

func InfluxWorkerGenerator(jobs chan interface{}, connection, dumpFile string, version float32, connector *InfluxConnector) func(workerId int) *InfluxWorker {
	return func(workerId int) *InfluxWorker {
		worker := &InfluxWorker{workerId, make(chan bool), make(chan bool, 1), jobs, connection, dumpFile, statistics.NewCmdStatisticReceiver(), logging.GetLogger(), version, connector, http.Client{}, true}
		go worker.run()
		return worker
	}
}

func (worker *InfluxWorker) Stop() {
	worker.quitInternal <- true
	worker.quit <- true
	<-worker.quit
	worker.IsRunning = false
	worker.log.Debug("InfluxWorker stopped")
}

func (worker InfluxWorker) run() {
	var queries []interface{}
	var query interface{}
	for {
		select {
		case <-worker.quit:
			worker.log.Debug("InfluxWorker quitting...")
			worker.sendBuffer(queries)
			worker.quit <- true
			return
		case query = <-worker.jobs:
			queries = append(queries, query)
			if len(queries) == 200 {
				worker.sendBuffer(queries)
				queries = queries[:0]
			}
		case <-time.After(time.Duration(10) * time.Second):
			worker.sendBuffer(queries)
			queries = queries[:0]
		}
	}
}

var mutex = &sync.Mutex{}

func (worker *InfluxWorker) sendBuffer(queries []interface{}) {
	if len(queries) > 0 {
		var err error
		var lineQueries []string
		for _, query := range queries {
			cast, err := worker.castJobToString(query)
			if err == nil {
				lineQueries = append(lineQueries, cast)
			}
		}

		var dataToSend []byte
		for _, lineQuery := range lineQueries {
			dataToSend = append(dataToSend, []byte(lineQuery)...)
		}

		startTime := time.Now()
		err = worker.sendData([]byte(dataToSend))
		worker.statistics.ReceiveQueries("send", statistics.QueriesPerTime{len(lineQueries), time.Now().Sub(startTime)})

		//This error says the data can't be sent and quit was signaled to stop
		if err == errorInterrupted {
			mutex.Lock()
			worker.log.Debugf("Global queue %d own queue %d", len(worker.jobs), len(lineQueries))
			if len(worker.jobs) != 0 || len(lineQueries) != 0 {
				worker.log.Debug("Saving queries to disk")

				lineQueries = append(lineQueries, worker.readQueriesFromQueue()...)

				worker.log.Debugf("dumping %d queries", len(lineQueries))
				worker.dumpQueries(worker.dumpFile, lineQueries)
			}
			mutex.Unlock()
		} else if err == errorFailedToSend {
			errorFile := worker.dumpFile + "-errors"
			worker.log.Warnf("Dumping queries with errors to: %s", errorFile)
			worker.dumpQueries(errorFile, lineQueries)
		}
	}
}

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
		case <-time.After(time.Duration(500) * time.Millisecond):
			stop = true
		}
	}
	return queries
}

func (worker InfluxWorker) sendData(rawData []byte) error {
	for {
		req, err := http.NewRequest("POST", worker.connection, bytes.NewBuffer(rawData))
		if err != nil {
			worker.log.Warn(err)
		}

		resp, err := worker.httpClient.Do(req)
		if err != nil {
			worker.log.Warn(err)
			worker.log.Warn("failed to send data, retrying...")
			if err := worker.waitForQuit(errorInterrupted); err != nil {
				return err
			}
		} else {
			defer resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode > 300 {
				worker.logHttpResponse(resp)
				if err := worker.waitForQuit(errorFailedToSend); err != nil {
					return err
				}
			} else if resp.StatusCode == 500 {
				worker.logHttpResponse(resp)
			} else {
				select {
				case <-worker.quitInternal:
					worker.quitInternal <- true
					return errorInterrupted
				}
				return nil
			}
		}
	}
}
func (worker InfluxWorker) logHttpResponse(resp *http.Response) {
	body, _ := ioutil.ReadAll(resp.Body)
	worker.log.Warnf("Influx status:%s message: %s", resp.Status, string(body))
}

func (worker InfluxWorker) waitForQuit(err error) error {
	select {
	//Got stop signal
	case <-worker.quitInternal:
		worker.log.Debug("Recived quit")
		worker.quitInternal <- true
		return err
	//Timeout and retry
	case <-time.After(time.Duration(10) * time.Second):
		return nil
	}
}

func (worker InfluxWorker) dumpQueries(filename string, queries []string) {
	worker.log.Debug("Dumping Queries...")
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

func (worker InfluxWorker) castJobToString(job interface{}) (string, error) {
	var result string
	var err error
	switch jobCast := job.(type) {
		case collector.PerformanceData:
		if worker.version >= 0.9 {
			result = jobCast.String()
		} else {
			worker.log.Fatalf("This influxversion [%f] given in the config is not supportet", worker.version)
			err = errors.New("This influxversion given in the config is not supportet")
		}
		case string:
		result = jobCast
		default:
		worker.log.Fatal("Could not cast object:", job)
		err = errors.New("Could not cast object")
	}
	if result[len(result)-1:] != "\n" {
		result += "\n"
	}
	return result, err
}
