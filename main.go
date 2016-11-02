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

//Buffer size.
const resultQueueLength = 1000.0

//nagfluxVersion contains the current Github-Release
const nagfluxVersion string = "v0.2.6"

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
	resultQueues := map[data.Datatype]chan collector.Printable{}
	stoppables := []Stoppable{}
	if len(cfg.Main.FieldSeparator) < 1 {
		panic("FieldSeparator is too short!")
	}
	pro := statistics.NewPrometheusServer(cfg.Monitoring.PrometheusAddress)
	pro.WatchResultQueueLength(resultQueues)
	fieldSeparator := []rune(cfg.Main.FieldSeparator)[0]

	config.PauseNagflux.Store(false)

	if cfg.Influx.Enabled {
		resultQueues[data.InfluxDB] = make(chan collector.Printable, cfg.Main.BufferSize)
		influx := influx.ConnectorFactory(resultQueues[data.InfluxDB], cfg.Influx.Address, cfg.Influx.Arguments, cfg.Main.DumpFile, cfg.Influx.Version, cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, cfg.Influx.CreateDatabaseIfNotExists)
		stoppables = append(stoppables, influx)
		influxDumpFileCollector := nagflux.NewDumpfileCollector(resultQueues[data.InfluxDB], cfg.Main.DumpFile, data.InfluxDB)
		stoppables = append(stoppables, influxDumpFileCollector)
	}

	if cfg.Elasticsearch.Enabled {
		resultQueues[data.Elasticsearch] = make(chan collector.Printable, cfg.Main.BufferSize)
		elasticsearch := elasticsearch.ConnectorFactory(resultQueues[data.Elasticsearch], cfg.Elasticsearch.Address, cfg.Elasticsearch.Index, cfg.Main.DumpFile, cfg.Elasticsearch.Version, cfg.Main.InfluxWorker, cfg.Main.MaxInfluxWorker, true)
		stoppables = append(stoppables, elasticsearch)
		elasticDumpFileCollector := nagflux.NewDumpfileCollector(resultQueues[data.Elasticsearch], cfg.Main.DumpFile, data.Elasticsearch)
		stoppables = append(stoppables, elasticDumpFileCollector)
	}

	//Some time for the dumpfile to fill the queue
	time.Sleep(time.Duration(100) * time.Millisecond)

	liveconnector := &livestatus.Connector{log, cfg.Livestatus.Address, cfg.Livestatus.Type}
	livestatusCollector := livestatus.NewLivestatusCollector(resultQueues, liveconnector, true)
	livestatusCache := livestatus.NewLivestatusCacheBuilder(liveconnector)

	if cfg.ModGearman.Enabled {
		log.Infof("Mod_Gearman: %s [%s]", cfg.ModGearman.Address, cfg.ModGearman.Queue)
		secret := modGearman.GetSecret(cfg.ModGearman.Secret, cfg.ModGearman.SecretFile)
		for i := 0; i < cfg.ModGearman.Worker; i++ {
			gearmanWorker := modGearman.NewGearmanWorker(cfg.ModGearman.Address,
				cfg.ModGearman.Queue,
				secret,
				resultQueues,
				livestatusCache,
			)
			stoppables = append(stoppables, gearmanWorker)
		}
	}

	log.Info("Nagios Spoolfile Folder: ", cfg.Main.NagiosSpoolfileFolder)
	nagiosCollector := spoolfile.NagiosSpoolfileCollectorFactory(cfg.Main.NagiosSpoolfileFolder, cfg.Main.NagiosSpoolfileWorker, resultQueues, livestatusCache)

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

//Wait till the Performance Data is sent.
func cleanUp(itemsToStop []Stoppable, resultQueues map[data.Datatype]chan collector.Printable) {
	log.Info("Cleaning up...")
	for i := len(itemsToStop) - 1; i >= 0; i-- {
		itemsToStop[i].Stop()
		time.Sleep(500 * time.Millisecond)
	}
	for _, q := range resultQueues {
		log.Debugf("Remaining queries %d", len(q))
	}
}
