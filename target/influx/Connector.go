package influx

import (
	"crypto/tls"
	"encoding/json"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

//Connector makes the basic connection to an influxdb.
type Connector struct {
	connectionHost string
	connectionArgs string
	dumpFile       string
	workers        []*Worker
	maxWorkers     int
	jobs           chan collector.Printable
	quit           chan bool
	log            *factorlog.FactorLog
	version        string
	isAlive        bool
	databaseExists bool
	databaseName   string
	httpClient     http.Client
}

var regexDatabaseName = regexp.MustCompile(`.*db=(.*)`)

//ConnectorFactory Constructor which will create some workers if the connection is established.
func ConnectorFactory(jobs chan collector.Printable, connectionHost, connectionArgs, dumpFile, version string, workerAmount, maxWorkers int, createDatabaseIfNotExists bool) *Connector {
	var databaseName string
	for _, argument := range strings.Split(connectionArgs, "&") {
		hits := regexDatabaseName.FindStringSubmatch(argument)
		if len(hits) > 1 {
			databaseName = hits[1]
		}
	}
	timeout := time.Duration(5 * time.Second)
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
	client := http.Client{Timeout: timeout, Transport: transport}
	s := &Connector{
		connectionHost, connectionArgs, dumpFile, make([]*Worker, workerAmount), maxWorkers,
		jobs, make(chan bool), logging.GetLogger(), version, false, false, databaseName, client,
	}

	gen := WorkerGenerator(jobs, connectionHost+"/write?"+connectionArgs, dumpFile, version, s, data.InfluxDB)
	s.TestIfIsAlive()
	if !s.isAlive {
		s.log.Info("Waiting for InfluxDB server")
	}
	for !s.isAlive {
		s.TestIfIsAlive()
		time.Sleep(time.Duration(5) * time.Second)
		s.log.Debugln("Waiting for InfluxDB server")
	}
	if s.isAlive {
		s.log.Debug("Influxdb is running")
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
		s.log.Panic("Database does not exists and Nagflux was not able to created")
	}

	for w := 0; w < workerAmount; w++ {
		s.workers[w] = gen(w)
	}
	go s.run()
	return s
}

//AddWorker creates a new worker
func (connector *Connector) AddWorker() {
	oldLength := connector.AmountWorkers()
	if oldLength < connector.maxWorkers {
		gen := WorkerGenerator(
			connector.jobs, connector.connectionHost+"/write?"+connector.connectionArgs,
			connector.dumpFile, connector.version, connector, data.InfluxDB,
		)
		connector.workers = append(connector.workers, gen(oldLength+2))
		connector.log.Infof("Starting Worker: %d -> %d", oldLength, connector.AmountWorkers())
	}
}

//RemoveWorker stops a worker
func (connector *Connector) RemoveWorker() {
	oldLength := connector.AmountWorkers()
	if oldLength > 1 {
		lastWorkerIndex := oldLength - 1
		connector.workers[lastWorkerIndex].Stop()
		connector.workers = connector.workers[:lastWorkerIndex]
		connector.log.Infof("Stopping Worker: %d -> %d", oldLength, connector.AmountWorkers())
	}
}

//AmountWorkers current amount of workers.
func (connector Connector) AmountWorkers() int {
	return len(connector.workers)
}

//IsAlive is the database system alive.
func (connector Connector) IsAlive() bool {
	return connector.isAlive
}

//DatabaseExists does the database exist.
func (connector Connector) DatabaseExists() bool {
	return connector.databaseExists
}

//Stop the connector and its workers.
func (connector *Connector) Stop() {
	connector.quit <- true
	<-connector.quit
	connector.log.Debug("InfluxConnectorFactory stopped")
}

//Waits just for the end.
func (connector *Connector) run() {
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

//TestIfIsAlive test active if the database system is alive.
func (connector *Connector) TestIfIsAlive() bool {
	result := helper.RequestedReturnCodeIsOK(connector.httpClient, connector.connectionHost+"/ping", "GET")
	connector.isAlive = result
	connector.log.Infof("Is InfluxDB running: %t", result)
	config.PauseNagflux.Store(!result)
	return result
}

//TestDatabaseExists test active if the database exists.
func (connector *Connector) TestDatabaseExists() bool {
	resp, err := connector.httpClient.Get(connector.connectionHost + "/query?q=show%20databases&" + connector.connectionArgs)
	if err != nil {
		return false
	}
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

//CreateDatabase creates the database.
func (connector *Connector) CreateDatabase() bool {
	resp, err := connector.httpClient.Get(connector.connectionHost + "/query?q=create%20database%20" + connector.databaseName)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if string(body) == `results":[{}]}` {
		return true
	}
	return false
}
