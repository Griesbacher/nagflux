package elasticsearch

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"net/http"
	"strings"
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
	templateExists bool
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
	s.TestTemplateExists()
	for i := 0; i < 5 && !s.templateExists; i++ {
		time.Sleep(time.Duration(2) * time.Second)
		if createDatabaseIfNotExists {
			s.createTemplate()
		}
		s.TestTemplateExists()
	}
	if !s.templateExists {
		s.log.Panic("Template does not exists and was not able to created")
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
	return connector.templateExists
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

//TestTemplateExists test active if the template exists.
func (connector *Connector) TestTemplateExists() bool {
	result, body := helper.SentReturnCodeIsOK(connector.httpClient, connector.connectionHost+"_template", "GET", "")
	if result && strings.Contains(body, fmt.Sprintf(`"%s":`, connector.index)) {
		connector.templateExists = true
	} else {
		connector.templateExists = false
	}
	return connector.templateExists
}

//createTemplate creates the nagflux template.
func (connector *Connector) createTemplate() bool {
	mapping := fmt.Sprintf(NagfluxTemplate, connector.index, config.GetConfig().Elasticsearch.NumberOfShards, config.GetConfig().Elasticsearch.NumberOfReplicas)
	createIndex, _ := helper.SentReturnCodeIsOK(connector.httpClient, connector.connectionHost+"_template/"+connector.index, "PUT", mapping)
	if !createIndex {
		return false
	}
	return true
}

//NagfluxTemplate creates a template for settings and mapping for nagflux indices.
const NagfluxTemplate = `{
  "template": "%s-*",
  "settings": {
    "index": {
      "number_of_shards": "%d",
      "number_of_replicas": "%d",
      "refresh_interval": "300s"
    }
  },
  "mappings": {
    "messages": {
      "properties": {
        "service": {
          "index": "not_analyzed",
          "type": "string"
        },
        "author": {
          "index": "not_analyzed",
          "type": "string"
        },
        "host": {
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
        },
        "timestamp": {
          "format": "strict_date_optional_time||epoch_millis",
          "type": "date"
        }
      }
    },
    "metrics": {
      "properties": {
        "max": {
          "index": "no",
          "type": "float"
        },
        "performanceLabel": {
          "index": "not_analyzed",
          "type": "string"
        },
        "warn-max": {
          "index": "no",
          "type": "float"
        },
        "warn-fill": {
          "index": "no",
          "type": "string"
        },
        "command": {
          "index": "not_analyzed",
          "type": "string"
        },
        "warn": {
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
        "crit": {
          "index": "no",
          "type": "float"
        },
        "service": {
          "index": "not_analyzed",
          "type": "string"
        },
        "host": {
          "index": "not_analyzed",
          "type": "string"
        },
        "value": {
          "index": "no",
          "type": "float"
        },
        "timestamp": {
          "format": "strict_date_optional_time||epoch_millis",
          "type": "date"
        },
        "warn-min": {
          "index": "no",
          "type": "float"
        },
        "crit-min": {
          "index": "no",
          "type": "float"
        }
      }
    },
    "_default_": {
      "_source": {
        "enabled": false
      },
      "dynamic_templates": [
        {
          "strings": {
            "mapping": {
              "index": "not_analyzed",
              "type": "string"
            },
            "match_mapping_type": "string",
            "match": "*"
          }
        }
      ],
      "_all": {
        "enabled": false
      }
    }
  },
  "aliases": {}
}`
