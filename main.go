package main

import (
	"flag"
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/collector/modGearman"
	"github.com/griesbacher/nagflux/collector/nagflux"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/statistics"
	"github.com/griesbacher/nagflux/target/elasticsearch"
	"github.com/griesbacher/nagflux/target/influx"
	"github.com/kdar/factorlog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//Stoppable represents every daemonlike struct which can be stopped
type Stoppable interface {
	Stop()
}

//Interval of the main loop, in which the amount of workers are calculated.
const updateRate = 1

//nagfluxVersion contains the current Github-Release
const nagfluxVersion string = "v0.3.1"

var log *factorlog.FactorLog
var quit = make(chan bool)

func main() {
	//Parse Args
	var configPath string
	var printver bool
	flag.Usage = func() {
		fmt.Println(`Nagflux by Philip Griesbacher`, nagfluxVersion, `
Commandline Parameter:
-configPath Path to the config file. If no file path is given the default is ./config.gcfg.
-V Print version and exit`)
	}
	flag.StringVar(&configPath, "configPath", "config.gcfg", "path to the config file")
	flag.BoolVar(&printver, "V", false, "print version and exit")
	flag.Parse()

	//Print version and exit
	if printver {
		fmt.Println(nagfluxVersion)
		os.Exit(0)
	}

	//Load config
	config.InitConfig(configPath)
	cfg := config.GetConfig()

	//Create Logger
	logging.InitLogger(cfg.Log.LogFile, cfg.Log.MinSeverity)
	log = logging.GetLogger()
	log.Info(`Started Nagflux `, nagfluxVersion)
	log.Debugf("Using Config: %s", configPath)
	resultQueues := collector.ResultQueues{}
	stoppables := []Stoppable{}
	if len(cfg.Main.FieldSeparator) < 1 {
		panic("FieldSeparator is too short!")
	}
	pro := statistics.NewPrometheusServer(cfg.Monitoring.PrometheusAddress)
	pro.WatchResultQueueLength(resultQueues)
	fieldSeparator := []rune(cfg.Main.FieldSeparator)[0]

	for name, value := range cfg.InfluxDB {
		if value == nil || !(*value).Enabled {
			continue
		}
		influxConfig := (*value)
		target := data.Target{Name: name, Datatype: data.InfluxDB}
		config.StoreValue(target, false)
		resultQueues[target] = make(chan collector.Printable, cfg.Main.BufferSize)
		influx := influx.ConnectorFactory(
			resultQueues[target],
			influxConfig.Address, influxConfig.Arguments, cfg.Main.DumpFile, influxConfig.Version,
			cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, cfg.InfluxDBGlobal.CreateDatabaseIfNotExists,
			influxConfig.StopPullingDataIfDown, target,
		)
		stoppables = append(stoppables, influx)
		influxDumpFileCollector := nagflux.NewDumpfileCollector(resultQueues[target], cfg.Main.DumpFile, target, cfg.Main.FileBufferSize)
		waitForDumpfileCollector(influxDumpFileCollector)
		stoppables = append(stoppables, influxDumpFileCollector)
	}

	for name, value := range cfg.Elasticsearch {
		if value == nil || !(*value).Enabled {
			continue
		}
		elasticConfig := (*value)
		target := data.Target{Name: name, Datatype: data.Elasticsearch}
		resultQueues[target] = make(chan collector.Printable, cfg.Main.BufferSize)
		config.StoreValue(target, false)
		elasticsearch := elasticsearch.ConnectorFactory(
			resultQueues[target],
			elasticConfig.Address, elasticConfig.Index, cfg.Main.DumpFile, elasticConfig.Version,
			cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, true,
		)
		stoppables = append(stoppables, elasticsearch)
		elasticDumpFileCollector := nagflux.NewDumpfileCollector(resultQueues[target], cfg.Main.DumpFile, target, cfg.Main.FileBufferSize)
		waitForDumpfileCollector(elasticDumpFileCollector)
		stoppables = append(stoppables, elasticDumpFileCollector)
	}

	//Some time for the dumpfile to fill the queue
	time.Sleep(time.Duration(100) * time.Millisecond)

	liveconnector := &livestatus.Connector{log, cfg.Livestatus.Address, cfg.Livestatus.Type}
	livestatusCollector := livestatus.NewLivestatusCollector(resultQueues, liveconnector, true)
	livestatusCache := livestatus.NewLivestatusCacheBuilder(liveconnector)

	for name, data := range cfg.ModGearman {
		if data == nil || !(*data).Enabled {
			continue
		}
		log.Infof("Mod_Gearman: %s - %s [%s]", name, (*data).Address, (*data).Queue)
		secret := modGearman.GetSecret((*data).Secret, (*data).SecretFile)
		for i := 0; i < (*data).Worker; i++ {
			gearmanWorker := modGearman.NewGearmanWorker((*data).Address,
				(*data).Queue,
				secret,
				resultQueues,
				livestatusCache,
			)
			stoppables = append(stoppables, gearmanWorker)
		}
	}

	log.Info("Nagios Spoolfile Folder: ", cfg.Main.NagiosSpoolfileFolder)
	nagiosCollector := spoolfile.NagiosSpoolfileCollectorFactory(
		cfg.Main.NagiosSpoolfileFolder,
		cfg.Main.NagiosSpoolfileWorker,
		resultQueues,
		livestatusCache,
		cfg.Main.FileBufferSize,
		collector.Filterable{Filter: cfg.Main.DefaultTarget},
	)

	log.Info("Nagflux Spoolfile Folder: ", cfg.Main.NagfluxSpoolfileFolder)
	nagfluxCollector := nagflux.NewNagfluxFileCollector(resultQueues, cfg.Main.NagfluxSpoolfileFolder, fieldSeparator)

	//Listen for Interrupts
	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, syscall.SIGINT)
	signal.Notify(interruptChannel, syscall.SIGTERM)
	go func() {
		<-interruptChannel
		log.Warn("Got Interrupted")
		stoppables = append(stoppables, []Stoppable{livestatusCollector, livestatusCache, nagiosCollector, nagfluxCollector}...)
		cleanUp(stoppables, resultQueues)
		quit <- true
	}()
loop:
	//Main loop
	for {
		select {
		case <-time.After(time.Duration(updateRate) * time.Second):
		/*queriesSend, measureTime, err := statisticUser.GetData("send")
			if err != nil {
				continue
			}
			idleTime := (measureTime.Seconds() - queriesSend.Time.Seconds() / float64(influx.AmountWorkers())) / updateRate
			log.Debugf("Buffer len: %d - Idletime in percent: %0.2f ", len(resultQueues[0]), idleTime * 100)

		//TODO: fix worker spawn by type
			if idleTime > 0.25 {
				influx.RemoveWorker()
			} else if idleTime < 0.1 && float64(len(resultQueues[0])) > resultQueueLength * 0.8 {
				influx.AddWorker()
			}*/
		case <-quit:
			break loop
		}
	}
}

func waitForDumpfileCollector(dump *nagflux.DumpfileCollector) {
	if dump != nil {
		for i := 0; i < 30 && dump.IsRunning; i++ {
			time.Sleep(time.Duration(2) * time.Second)
		}
	}
}

//Wait till the Performance Data is sent.
func cleanUp(itemsToStop []Stoppable, resultQueues collector.ResultQueues) {
	log.Info("Cleaning up...")
	for i := len(itemsToStop) - 1; i >= 0; i-- {
		itemsToStop[i].Stop()
		time.Sleep(500 * time.Millisecond)
	}
	for _, q := range resultQueues {
		log.Debugf("Remaining queries %d", len(q))
	}
}
