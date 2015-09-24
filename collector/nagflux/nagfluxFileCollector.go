package nagflux

import (
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//Provides a interface to nagflux, in which you could insert influxdb queries.
type NagfluxFileCollector struct {
	quit    chan bool
	results chan interface{}
	folder  string
	log     *factorlog.FactorLog
}

//Constructor, which also starts the collector.
func NewNagfluxFileCollector(results chan interface{}, folder string) *NagfluxFileCollector {
	s := &NagfluxFileCollector{make(chan bool, 1), results, folder, logging.GetLogger()}
	go s.run()
	return s
}

//Stops the Collector.
func (nfc *NagfluxFileCollector) Stop() {
	nfc.quit <- true
	<-nfc.quit
	nfc.log.Debug("NagfluxFileCollector stoped")
}

//Checks if the files are old enough, if so they will be added in the queue
func (nfc NagfluxFileCollector) run() {
	for {
		select {
		case <-nfc.quit:
			nfc.quit <- true
			return
		case <-time.After(spoolfile.IntervalToCheckDirectory):
			for _, currentFile := range spoolfile.FilesInDirectoryOlderThanX(nfc.folder, spoolfile.MinFileAgeInSeconds) {
				data, err := ioutil.ReadFile(currentFile)
				if err != nil {
					break
				}
				for _, line := range strings.SplitAfter(string(data), "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					select {
					case <-nfc.quit:
						nfc.quit <- true
						return
					case nfc.results <- line:
					case <-time.After(time.Duration(1) * time.Minute):
						nfc.log.Warn("NagfluxFileCollector: Could not write to buffer")
					}
				}
				err = os.Remove(currentFile)
				if err != nil {
					logging.GetLogger().Warn(err)
				}
			}
		}
	}
}
