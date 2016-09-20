package modGearman

import (
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/helper/crypto"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"github.com/mikespook/gearman-go/worker"
	"time"
)

//GearmanWorker queries the gearmanserver and adds the extraced perfdata to the queue.
type GearmanWorker struct {
	quit                  chan bool
	results               map[data.Datatype]chan collector.Printable
	nagiosSpoolfileWorker *spoolfile.NagiosSpoolfileWorker
	aesECBDecrypter       *crypto.AESECBDecrypter
	worker                *worker.Worker
	log                   *factorlog.FactorLog
	jobQueue              string
	pauseChannel          chan bool
}

//NewGearmanWorker generates a new GearmanWorker.
//leave the key empty to disable encryption, otherwise the gearmanpacketes are expected to be encrpyten with AES-ECB 128Bit and a 32 Byte Key.
func NewGearmanWorker(address, queue, key string, results map[data.Datatype]chan collector.Printable, livestatusCacheBuilder *livestatus.CacheBuilder, pauseChannel chan bool) *GearmanWorker {
	var decrypter *crypto.AESECBDecrypter
	if key != "" {
		byteKey := ShapeKey(key, 32)
		var err error
		decrypter, err = crypto.NewAESECBDecrypter(byteKey)
		if err != nil {
			panic(err)
		}
	}
	worker := &GearmanWorker{
		quit:                  make(chan bool),
		results:               results,
		nagiosSpoolfileWorker: spoolfile.NewNagiosSpoolfileWorker(-1, make(chan string), make(map[data.Datatype]chan collector.Printable), livestatusCacheBuilder),
		aesECBDecrypter:       decrypter,
		worker:                createGearmanWorker(address),
		log:                   logging.GetLogger(),
		jobQueue:              queue,
		pauseChannel:          pauseChannel,
	}
	go worker.run()
	go worker.handleLoad()
	go worker.handlePause()

	return worker
}

func createGearmanWorker(address string) *worker.Worker {
	w := worker.New(worker.Unlimited)
	w.AddServer("tcp4", address)
	return w
}

func (g GearmanWorker) startGearmanWorker() error {
	g.worker.ErrorHandler = func(err error) {
		if err.Error() == "EOF" {
			g.log.Warn("Gearmand did not response. Connection closed")
		} else {
			g.log.Warn(err)
		}
		g.run()
	}
	g.worker.AddFunc(g.jobQueue, g.handelJob, worker.Unlimited)
	if err := g.worker.Ready(); err != nil {
		return err
	}
	go g.worker.Work()
	return nil
}

//Stop stops the worker
func (g GearmanWorker) Stop() {
	g.worker.Close()
	g.quit <- true
	<-g.quit
	logging.GetLogger().Debug("GearmanWorker stopped")
}

func (g GearmanWorker) run() {
	for {
		if err := g.startGearmanWorker(); err != nil {
			g.log.Warn(err)
			time.Sleep(time.Duration(30) * time.Second)
		} else {
			return
		}
	}
}

func (g GearmanWorker) handleLoad() {
	bufferLimit := int(float32(config.GetConfig().Main.BufferSize) * 0.90)
	for {
		for _, r := range g.results {
			if len(r) > bufferLimit && g.worker != nil {
				g.worker.Lock()
				for len(r) > bufferLimit {
					time.Sleep(time.Duration(100) * time.Millisecond)
				}
				g.worker.Unlock()
			}
		}
		select {
		case <-g.quit:
			g.quit <- true
			return
		case <-time.After(time.Duration(1) * time.Second):
		}
	}
}

func (g GearmanWorker) handlePause() {
	var pause bool
	for {
		select {
		case <-g.quit:
			g.quit <- true
			return
		case pause = <- g.pauseChannel:
			logging.GetLogger().Info("Gearman-Worker recived paussignal: ", pause)
			if pause {
				g.worker.Lock()
			} else {
				g.worker.Unlock()
			}
		case <-time.After(time.Duration(1) * time.Second):
		}
	}
}

func (g GearmanWorker) handelJob(job worker.Job) ([]byte, error) {
	secret := job.Data()
	if g.aesECBDecrypter != nil {
		var err error
		secret, err = g.aesECBDecrypter.Decypt(secret)
		if err != nil {
			g.log.Warn(err, ". Data: ", string(job.Data()))
			return job.Data(), nil
		}
	}
	splittedPerformanceData := helper.StringToMap(string(secret), "\t", "::")
	g.log.Debug("[ModGearman] ", string(job.Data()))
	g.log.Debug("[ModGearman] ", splittedPerformanceData)
	for singlePerfdata := range g.nagiosSpoolfileWorker.PerformanceDataIterator(splittedPerformanceData) {
		for _, r := range g.results {
			select {
			case r <- singlePerfdata:
			case <-time.After(time.Duration(1) * time.Minute):
				logging.GetLogger().Warn("GearmanWorker: Could not write to buffer")
			}
		}
	}
	return job.Data(), nil
}
