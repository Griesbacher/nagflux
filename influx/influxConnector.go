package influx

import (
	"encoding/json"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//Makes the basic connection to an influxdb.
type InfluxConnector struct {
	connectionHost string
	connectionArgs string
	dumpFile       string
	workers        []*InfluxWorker
	maxWorkers     int
	jobs           chan interface{}
	quit           chan bool
	log            *factorlog.FactorLog
	version        float32
	isAlive        bool
	databaseExists bool
	databaseName   string
}

//Constructor which will create some workers if the connection is established.
func InfluxConnectorFactory(jobs chan interface{}, connectionHost, connectionArgs, dumpFile string, workerAmount, maxWorkers int, version float32, createDatabaseIfNotExists bool) *InfluxConnector {

	regexDatabaseName, err := regexp.Compile(`.*db=(.*)`)
	if err != nil {
		logging.GetLogger().Error("Regex creation failed:", err)
	}
	var databaseName string
	for _, argument := range strings.Split(connectionArgs, "&") {
		hits := regexDatabaseName.FindStringSubmatch(argument)
		if len(hits) > 1 {
			databaseName = hits[1]
		}
	}
	s := &InfluxConnector{connectionHost, connectionArgs, dumpFile, make([]*InfluxWorker, workerAmount), maxWorkers, jobs, make(chan bool), logging.GetLogger(), version, false, false, databaseName}

	gen := InfluxWorkerGenerator(jobs, connectionHost+"/write?"+connectionArgs, dumpFile, version, s)

	s.TestIfIsAlive()
	for i := 0; i < 5 && !s.isAlive; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		s.TestIfIsAlive()
	}
	if !s.isAlive {
		s.log.Panic("Influxdb not running")
	}
	s.TestDatabaseExists()
	for i := 0; i < 5 && !s.databaseExists; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		if createDatabaseIfNotExists {
			s.CreateDatabase()
		}
		s.TestDatabaseExists()
	}
	if !s.databaseExists {
		s.log.Panic("Database does not exists and was not able to created")
	}

	for w := 0; w < workerAmount; w++ {
		s.workers[w] = gen(w)
	}
	go s.run()
	return s
}

//Creates a new worker
func (connector *InfluxConnector) AddWorker() {
	oldLength := connector.AmountWorkers()
	if oldLength < connector.maxWorkers {
		gen := InfluxWorkerGenerator(connector.jobs, connector.connectionHost+"/write?"+connector.connectionArgs, connector.dumpFile, connector.version, connector)
		connector.workers = append(connector.workers, gen(oldLength+2))
		connector.log.Debugf("Starting Worker: %d -> %d", oldLength, connector.AmountWorkers())
	}
}

//Stops a worker
func (connector *InfluxConnector) RemoveWorker() {
	oldLength := connector.AmountWorkers()
	if oldLength > 1 {
		lastWorkerIndex := oldLength - 1
		connector.workers[lastWorkerIndex].Stop()
		connector.workers = connector.workers[:lastWorkerIndex]
		connector.log.Debugf("Stopping Worker: %d -> %d", oldLength, connector.AmountWorkers())
	}
}

//Current amount of workers.
func (connector InfluxConnector) AmountWorkers() int {
	return len(connector.workers)
}

//Is the database system alive.
func (connector InfluxConnector) IsAlive() bool {
	return connector.isAlive
}

//Does the database exist.
func (connector InfluxConnector) DatabaseExists() bool {
	return connector.databaseExists
}

//Stop the connector and its workers.
func (connector *InfluxConnector) Stop() {
	connector.quit <- true
	<-connector.quit
	connector.log.Debug("InfluxConnectorFactory stopped")
}

//Waits just for the end.
func (connector *InfluxConnector) run() {
	for {
		select {
		case <-connector.quit:
			for _, worker := range connector.workers {
				go worker.Stop()
			}
			for len(connector.workers) > 0 {
				for connector.workers[0].IsRunning == true {
					time.Sleep(time.Duration(100) * time.Millisecond)
				}
				if len(connector.workers) > 1 {
					connector.workers = connector.workers[1:]
				} else {
					connector.workers = connector.workers[:0]
				}
			}
			connector.quit <- true
			return
		}
	}
}

//Test active if the database system is alive.
func (connector *InfluxConnector) TestIfIsAlive() bool {
	resp, err := http.Get(connector.connectionHost + "/ping")
	result := false
	if err != nil {
		return result
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result = true
	}

	connector.isAlive = result
	return result
}

//Represents the query result
type ShowSeriesResult struct {
	Results []struct {
		Series []struct {
			Columns []string
			Name    string
			Values  [][]string
		}
	}
}

//Test active if the database exists.
func (connector *InfluxConnector) TestDatabaseExists() bool {
	resp, _ := http.Get(connector.connectionHost + "/query?q=show%20databases")

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		var jsonResult ShowSeriesResult
		json.Unmarshal(body, &jsonResult)
		for _, tablename := range jsonResult.Results[0].Series[0].Values {
			if connector.databaseName == tablename[0] {
				connector.databaseExists = true
				return true
			}
		}
	}
	connector.databaseExists = false
	return false
}

//Creates the database.
func (connector *InfluxConnector) CreateDatabase() bool {
	resp, _ := http.Get(connector.connectionHost + "/query?q=create%20database%20" + connector.databaseName)

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == `results":[{}]}` {
		return true
	} else {
		return false
	}
}
