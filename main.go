package main

import (
	"flag"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/influx"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/monitoring"
	"github.com/griesbacher/nagflux/statistics"
	"github.com/kdar/factorlog"
	"gopkg.in/gcfg.v1"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type Stoppable interface {
	Stop()
}

type Config struct {
	Main struct {
		NagiosSpoolfileFolder  string
		NagiosSpoolfileWorker  int
		InfluxWorker           int
		MaxInfluxWorker        int
		NumberOfCPUs           int
		DumpFile               string
		NagfluxSpoolfileFolder string
	}
	Log struct {
		LogFile     string
		MinSeverity string
	}
	Monitoring struct {
		WebserverPort string
	}
	Influx struct {
		Address string
		Version float32
	}
	Grafana struct {
		FieldSeperator string
	}
}

const updateRate = 120
const resultQueueLength = 1000.0

var log *factorlog.FactorLog

func main() {
	//Parse Args
	var configPath string
	flag.StringVar(&configPath, "configPath", "config.gcfg", "path to the config file")
	flag.Parse()

	//Load config
	var cfg Config
	err := gcfg.ReadFileInto(&cfg, configPath)
	if err != nil {
		panic(err)
	}

	//Create Logger
	logging.InitLogger(cfg.Log.LogFile, cfg.Log.MinSeverity)
	log = logging.GetLogger()

	//Set CPUs to use
	runtime.GOMAXPROCS(cfg.Main.NumberOfCPUs)
	log.Infof("Using %d of %d CPUs", cfg.Main.NumberOfCPUs, runtime.NumCPU())

	log.Info("Spoolfile Folder: ", cfg.Main.NagiosSpoolfileFolder)
	resultQueue := make(chan interface{}, int(resultQueueLength))
	influx := influx.InfluxConnectorFactory(resultQueue, cfg.Influx.Address, cfg.Main.DumpFile, cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, cfg.Influx.Version)

	dumpFileCollector := collector.NewDumpfileCollector(resultQueue, cfg.Main.DumpFile)
	//Some time for the dumpfile to fill the queue
	time.Sleep(time.Duration(100) * time.Millisecond)

	nagiosCollector := collector.NagiosSpoolfileCollectorFactory(cfg.Main.NagiosSpoolfileFolder, cfg.Main.NagiosSpoolfileWorker, resultQueue, cfg.Grafana.FieldSeperator)

	nagfluxCollector := collector.NewNagfluxFileCollector(resultQueue, cfg.Main.NagfluxSpoolfileFolder)

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
		cleanUp(nagiosCollector, dumpFileCollector, influx, nagfluxCollector, resultQueue)
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

	cleanUp(nagiosCollector, dumpFileCollector, influx, nagfluxCollector, resultQueue)
}

//Wait till the Performance Data is sent
func cleanUp(nagiosCollector, dumpFileCollector, influx, nagfluxCollector Stoppable, resultQueue chan interface{}) {
	log.Info("Cleaning up...")
	if monitoringServer := monitoring.StartMonitoringServer(""); monitoringServer != nil {
		monitoringServer.Stop()
	}
	dumpFileCollector.Stop()
	nagiosCollector.Stop()
	nagfluxCollector.Stop()
	time.Sleep(1 * time.Second)
	influx.Stop()
	log.Debugf("Remaining queries %d", len(resultQueue))
}
