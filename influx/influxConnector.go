package influx

import (
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"time"
)

type InfluxConnector struct {
	connection string
	dumpFile   string
	workers    []*InfluxWorker
	maxWorkers int
	jobs       chan interface{}
	quit       chan bool
	log        *factorlog.FactorLog
	version    float32
}

func InfluxConnectorFactory(jobs chan interface{}, connection, dumpFile string, workerAmount, maxWorkers int, version float32) *InfluxConnector {
	s := &InfluxConnector{connection, dumpFile, make([]*InfluxWorker, workerAmount), maxWorkers, jobs, make(chan bool), logging.GetLogger(), version}

	gen := InfluxWorkerGenerator(jobs, connection, dumpFile, version, s)

	for w := 0; w < workerAmount; w++ {
		s.workers[w] = gen(w)
	}

	go s.run()
	return s
}
func (factory *InfluxConnector) AddWorker() {
	oldLength := factory.AmountWorkers()
	if oldLength < factory.maxWorkers {
		gen := InfluxWorkerGenerator(factory.jobs, factory.connection, factory.dumpFile, factory.version, factory)
		factory.workers = append(factory.workers, gen(oldLength+2))
		factory.log.Debugf("Starting Worker: %d -> %d", oldLength, factory.AmountWorkers())
	}
}
func (factory *InfluxConnector) RemoveWorker() {
	oldLength := factory.AmountWorkers()
	if oldLength > 1 {
		lastWorkerIndex := oldLength - 1
		factory.workers[lastWorkerIndex].Stop()
		factory.workers = factory.workers[:lastWorkerIndex]
		factory.log.Debugf("Stopping Worker: %d -> %d", oldLength, factory.AmountWorkers())
	}
}

func (factory *InfluxConnector) AmountWorkers() int {
	return len(factory.workers)
}

func (factory *InfluxConnector) Stop() {
	factory.quit <- true
	<-factory.quit
	factory.log.Debug("InfluxConnectorFactory stopped")
}

func (factory *InfluxConnector) run() {
	for {
		select {
		case <-factory.quit:
			for _, worker := range factory.workers {
				go worker.Stop()
			}
			for len(factory.workers) > 0 {
				for factory.workers[0].IsRunning == true {
					time.Sleep(time.Duration(100) * time.Millisecond)
				}
				if len(factory.workers) > 1 {
					factory.workers = factory.workers[1:]
				} else {
					factory.workers = factory.workers[:0]
				}
			}
			factory.quit <- true
			return
		}
	}
}
