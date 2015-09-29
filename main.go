package main

import (
	"flag"
	"fmt"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/collector/nagflux"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/influx"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/monitoring"
	"github.com/griesbacher/nagflux/statistics"
	"github.com/kdar/factorlog"
	"gopkg.in/gcfg.v1"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Stoppable interface {
	Stop()
}

//Interval of the main loop, in which the amount of workers are calculated.
const updateRate = 120

//Buffer size.
const resultQueueLength = 1000.0

var log *factorlog.FactorLog

func main() {
	//Parse Args
	var configPath string
	flag.Usage = func() {
		fmt.Println(`Nagflux by Philip Griesbacher @ 2015
Commandline Parameter:
-configPath Path to the config file. If no file path is given the default is ./config.gcfg.
		`)
	}
	flag.StringVar(&configPath, "configPath", "config.gcfg", "path to the config file")
	flag.Parse()

	//Load config
	var cfg config.Config
	err := gcfg.ReadFileInto(&cfg, configPath)
	if err != nil {
		panic(err)
	}

	//Create Logger
	logging.InitLogger(cfg.Log.LogFile, cfg.Log.MinSeverity)
	log = logging.GetLogger()

	resultQueue := make(chan interface{}, int(resultQueueLength))
	influx := influx.InfluxConnectorFactory(resultQueue, cfg.Influx.Address, cfg.Influx.Arguments, cfg.Main.DumpFile, cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, cfg.Influx.Version, cfg.Influx.CreateDatabaseIfNotExists)

	dumpFileCollector := nagflux.NewDumpfileCollector(resultQueue, cfg.Main.DumpFile)
	//Some time for the dumpfile to fill the queue
	time.Sleep(time.Duration(100) * time.Millisecond)

	liveconnector := &livestatus.LivestatusConnector{log, cfg.Livestatus.Address, cfg.Livestatus.Type}
	livestatusCollector := livestatus.NewLivestatusCollector(resultQueue, liveconnector, cfg.Grafana.FieldSeperator)
	livestatusCache := livestatus.NewLivestatusCacheBuilder(liveconnector)

	log.Info("Nagios Spoolfile Folder: ", cfg.Main.NagiosSpoolfileFolder)
	nagiosCollector := spoolfile.NagiosSpoolfileCollectorFactory(cfg.Main.NagiosSpoolfileFolder, cfg.Main.NagiosSpoolfileWorker, resultQueue, cfg.Grafana.FieldSeperator, livestatusCache)

	log.Info("Nagflux Spoolfile Folder: ", cfg.Main.NagfluxSpoolfileFolder)
	nagfluxCollector := nagflux.NewNagfluxFileCollector(resultQueue, cfg.Main.NagfluxSpoolfileFolder)

	statisticUser := statistics.NewSimpleStatisticsUser()
	statisticUser.SetDataReceiver(statistics.NewCmdStatisticReceiver())

	if cfg.Monitoring.WebserverPort != "" {
		monitoring.StartMonitoringServer(cfg.Monitoring.WebserverPort)
	}

	//Listen for Interrupts
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT)
	signal.Notify(interruptChannel, syscall.SIGTERM)
	go func() {
		<-interruptChannel
		log.Warn("Got Interrupted")
		cleanUp([]Stoppable{livestatusCollector, livestatusCache, nagiosCollector, dumpFileCollector, nagfluxCollector, influx}, resultQueue)
		os.Exit(1)
	}()

	//Main loop
	for {
		select {
		case <-time.After(time.Duration(updateRate) * time.Second):
			queriesSend, measureTime, err := statisticUser.GetData("send")
			if err != nil {
				continue
			}
			idleTime := (measureTime.Seconds() - queriesSend.Time.Seconds()/float64(influx.AmountWorkers())) / updateRate
			log.Debugf("Buffer len: %d - Idletime in percent: %0.2f ", len(resultQueue), idleTime*100)

			if idleTime > 0.25 {
				influx.RemoveWorker()
			} else if idleTime < 0.1 && float64(len(resultQueue)) > resultQueueLength*0.8 {
				influx.AddWorker()
			}
		}
	}

	cleanUp([]Stoppable{livestatusCollector, livestatusCache, nagiosCollector, dumpFileCollector, nagfluxCollector, influx}, resultQueue)
}

//Wait till the Performance Data is sent.
func cleanUp(itemsToStop []Stoppable, resultQueue chan interface{}) {
	log.Info("Cleaning up...")
	if monitoringServer := monitoring.StartMonitoringServer(""); monitoringServer != nil {
		monitoringServer.Stop()
	}
	for _, item := range itemsToStop {
		item.Stop()
		time.Sleep(500 * time.Millisecond)
	}
	log.Debugf("Remaining queries %d", len(resultQueue))
}
