package main

import (
	"code.google.com/p/gcfg"
	"flag"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/influx"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/monitoring"
	"github.com/griesbacher/nagflux/statistics"
	"github.com/kdar/factorlog"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

type Config struct {
	Main struct {
		SpoolfileFolder string
		SpoolfileWorker int
		InfluxWorker    int
		MaxInfluxWorker int
		NumberOfCPUs    int
		DumpFile        string
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

	log.Info("Spoolfile Folder: ", cfg.Main.SpoolfileFolder)
	resultQueue := make(chan interface{}, int(resultQueueLength))
	influx := influx.InfluxConnectorFactory(resultQueue, cfg.Influx.Address, cfg.Main.DumpFile, cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, cfg.Influx.Version)

	dumpFileCollector := collector.NewDumpfileCollector(resultQueue, cfg.Main.DumpFile)
	//Some time for the dumpfile to fill the queue
	time.Sleep(time.Duration(100) * time.Millisecond)

	collector := collector.SpoolfileCollectorFactory(cfg.Main.SpoolfileFolder, cfg.Main.SpoolfileWorker, resultQueue, cfg.Grafana.FieldSeperator)

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
		cleanUp(collector, dumpFileCollector, influx, resultQueue)
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
			log.Debug(len(resultQueue), idleTime)

			if idleTime > 0.25 {
				influx.RemoveWorker()
			} else if idleTime < 0.1 && float64(len(resultQueue)) > resultQueueLength*0.8 {
				influx.AddWorker()
			}
		}
	}

	cleanUp(collector, dumpFileCollector, influx, resultQueue)
}

//Wait till the Performance Data is sent
func cleanUp(collector *collector.SpoolfileCollector, dumpFileCollector *collector.DumpfileCollector, influx *influx.InfluxConnector, resultQueue chan interface{}) {
	log.Info("Cleaning up...")
	if monitoringServer := monitoring.StartMonitoringServer(""); monitoringServer != nil {
		monitoringServer.Stop()
	}
	dumpFileCollector.Stop()
	collector.Stop()
	time.Sleep(1 * time.Second)
	influx.Stop()
	log.Debugf("Remaining queries %d", len(resultQueue))
}
