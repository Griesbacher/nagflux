package elasticsearch

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"net/http"
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
	httpClient     http.Client
}

//ConnectorFactory Constructor which will create some workers if the connection is established.
func ConnectorFactory(jobs chan collector.Printable, connectionHost, index, dumpFile string, workerAmount, maxWorkers int, version float32, createDatabaseIfNotExists bool) *Connector {
	if connectionHost[len(connectionHost)-1] != '/' {
		connectionHost += "/"
	}
	s := &Connector{connectionHost, index, dumpFile, make([]*Worker, workerAmount), maxWorkers,
		jobs, make(chan bool), logging.GetLogger(), version,
		false, false, http.Client{Timeout: time.Duration(5 * time.Second)},
	}

	gen := WorkerGenerator(jobs, connectionHost+"_bulk", index, dumpFile, version, s)

	s.TestIfIsAlive()
	for i := 0; i < 5 && !s.isAlive; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		s.TestIfIsAlive()
	}
	if !s.isAlive {
		s.log.Panic("Elasticsearch not running")
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
	connector.log.Debug("ElasticsearchConnectorFactory stopped")
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
	result := helper.RequestedReturnCodeIsOK(connector.httpClient, connector.connectionHost, "HEAD")
	connector.isAlive = result
	return result
}

//TestDatabaseExists test active if the database exists.
func (connector *Connector) TestDatabaseExists() bool {
	result := helper.RequestedReturnCodeIsOK(connector.httpClient, connector.connectionHost+connector.index, "HEAD")
	connector.databaseExists = result
	return result
}

//CreateDatabase creates the database.
func (connector *Connector) CreateDatabase() bool {
	mapping := fmt.Sprintf(MappingIndex, config.GetConfig().Elasticsearch.NumberOfShards, config.GetConfig().Elasticsearch.NumberOfReplicas)
	createIndex, _ := helper.SentReturnCodeIsOK(connector.httpClient, connector.connectionHost+connector.index, "PUT", mapping)
	if !createIndex {
		return false
	}
	createMessages, _ := helper.SentReturnCodeIsOK(connector.httpClient, connector.connectionHost+connector.index+"/messages/_mapping", "PUT", MappingMessages)
	if !createMessages {
		return false
	}
	createPerfdata, _ := helper.SentReturnCodeIsOK(connector.httpClient, connector.connectionHost+connector.index+"/metrics/_mapping", "PUT", MappingMetrics)
	if !createPerfdata {
		return false
	}
	return true
}

//MappingIndex is the mapping for the nagflux index
const MappingIndex = `{
  "settings": {
    "index": {
      "number_of_shards": "%d",
      "number_of_replicas": "%d",
	  "refresh_interval": "60s"
    }
  },
  "mappings": {
    "_default_": {
	  "dynamic_templates": [
        {
          "strings": {
            "match": "*",
            "match_mapping_type": "string",
            "mapping":   { "type": "string", "index": "not_analyzed" }
          }
        }
      ],
      "_source": {
        "enabled": false
      },
      "_all": {
        "enabled": false
      }
    }
  }
}`

//MappingMessages is the mapping used to store messages
const MappingMessages = `{
  "messages": {
    "properties": {
        "host": {
        "index": "not_analyzed",
        "type": "string"
      },
      "service": {
        "index": "not_analyzed",
        "type": "string"
      },
      "timestamp": {
        "format": "strict_date_optional_time||epoch_millis",
        "type": "date"
      },
      "author": {
        "index": "not_analyzed",
        "type": "string"
      },
      "type": {
        "index": "not_analyzed",
        "type": "string"
      },
      "message": {
        "index": "not_analyzed",
        "type": "string"
      }
    }
  }
}`

//MappingPerfdata is the mapping used to store performancedata
const MappingMetrics = `{
  "perfdata": {
    "properties": {
      "timestamp": {
        "format": "strict_date_optional_time||epoch_millis",
        "type": "date"
      },
      "host": {
        "index": "not_analyzed",
        "type": "string"
      },
      "service": {
        "index": "not_analyzed",
        "type": "string"
      },
      "command": {
        "index": "not_analyzed",
        "type": "string"
      },
      "performanceLabel": {
        "index": "not_analyzed",
        "type": "string"
      },
      "value": {
        "index": "no",
        "type": "float"
      },
      "warn": {
        "index": "no",
        "type": "float"
      },
      "warn-min": {
        "index": "no",
        "type": "float"
      },
      "warn-max": {
        "index": "no",
        "type": "float"
      },
      "warn-fill": {
        "index": "no",
        "type": "string"
      },
      "crit": {
        "index": "no",
        "type": "float"
      },
      "crit-min": {
        "index": "no",
        "type": "float"
      },
      "crit-max": {
        "index": "no",
        "type": "float"
      },
      "crit-fill": {
        "index": "no",
        "type": "string"
      },
      "min": {
        "index": "no",
        "type": "float"
      },
      "max": {
        "index": "no",
        "type": "float"
      }
    }
  }
}`
