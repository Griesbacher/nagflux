package elasticsearch

import (
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"time"
)

//Connector makes the basic connection to an influxdb.
type Connector struct {
	connectionHost string
	index          string
	dumpFile       string
	workers        []*Worker
	maxWorkers     int
	jobs           chan collector.Printable
	quit           chan bool
	log            *factorlog.FactorLog
	version        float32
	isAlive        bool
	databaseExists bool
}

//ConnectorFactory Constructor which will create some workers if the connection is established.
func ConnectorFactory(jobs chan collector.Printable, connectionHost, index, dumpFile string, workerAmount, maxWorkers int, version float32, createDatabaseIfNotExists bool) *Connector {
	s := &Connector{connectionHost, index, dumpFile, make([]*Worker, workerAmount), maxWorkers, jobs, make(chan bool), logging.GetLogger(), version, true, true} //TODO: change boolean to false

	gen := WorkerGenerator(jobs, connectionHost+"/_bulk", index, dumpFile, version, s)

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

//AddWorker creates a new worker
func (connector *Connector) AddWorker() {
	oldLength := connector.AmountWorkers()
	if oldLength < connector.maxWorkers {
		gen := WorkerGenerator(connector.jobs, connector.connectionHost+"/_bulk", connector.index, connector.dumpFile, connector.version, connector)
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
	return true //TODO: fix
}

//TestDatabaseExists test active if the database exists.
func (connector *Connector) TestDatabaseExists() bool {
	return true //TODO: fix
}

//CreateDatabase creates the database.
func (connector *Connector) CreateDatabase() bool {
	return true //TODO: fix
}
