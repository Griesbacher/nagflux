package modGearman

import (
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/helper/crypto"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"github.com/mikespook/gearman-go/worker"
	"log"
	"net"
	"os"
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
}

//NewGearmanWorker generates a new GearmanWorker.
//leave the key empty to disable encryption, otherwise the gearmanpacketes are expected to be encrpyten with AES-ECB 128Bit and a 32 Byte Key.
func NewGearmanWorker(address, queue, key string, results map[data.Datatype]chan collector.Printable, livestatusCacheBuilder *livestatus.CacheBuilder) *GearmanWorker {
	var decrypter *crypto.AESECBDecrypter
	if key != "" {
		byteKey := FillKey(key, 32)
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
		worker:                worker.New(worker.Unlimited),
		log:                   logging.GetLogger(),
	}
	worker.run(address, queue)
	return worker
}

//Stop stops the worker
func (g GearmanWorker) Stop() {
	//g.quit <- true
	g.worker.Close()
	//<-g.quit
	logging.GetLogger().Debug("GearmanWorker stopped")
}

func (g GearmanWorker) run(address, queue string) {
	g.worker.ErrorHandler = func(e error) {
		log.Println(e)
		if opErr, ok := e.(*net.OpError); ok {
			if !opErr.Temporary() {
				proc, err := os.FindProcess(os.Getpid())
				if err != nil {
					log.Println(err)
				}
				if err := proc.Signal(os.Interrupt); err != nil {
					log.Println(err)
				}
			}
		}
	}
	g.worker.AddServer("tcp4", address)
	g.worker.AddFunc(queue, g.handelJob, worker.Unlimited)
	if err := g.worker.Ready(); err != nil {
		log.Fatal(err)
		return
	}
	go g.worker.Work()

}

func (g GearmanWorker) handelJob(job worker.Job) ([]byte, error) {
	secret := job.Data()
	if g.aesECBDecrypter != nil {
		var err error
		secret, err = g.aesECBDecrypter.Decypt(secret)
		if err != nil {
			g.log.Warn(err)
		}
	}
	splittedPerformanceData := helper.StringToMap(string(secret), "\t", "::")
	for singlePerfdata := range g.nagiosSpoolfileWorker.PerformanceDataIterator(splittedPerformanceData) {
		for _, r := range g.results {
			select {
			case <-g.quit:
				g.quit <- true
				return job.Data(), nil
			case r <- singlePerfdata:
			case <-time.After(time.Duration(1) * time.Minute):
				logging.GetLogger().Warn("GearmanWorker: Could not write to buffer")
			}
		}
	}
	return job.Data(), nil
}
