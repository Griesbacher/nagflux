package collector

import (
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"time"
	"strings"
	"os"
)

type NagfluxFileCollector struct {
	quit    chan bool
	results chan interface{}
	folder  string
	log     *factorlog.FactorLog
}

func NewNagfluxFileCollector(results chan interface{}, folder string) *NagfluxFileCollector {
	s := &NagfluxFileCollector{make(chan bool, 1), results, folder, logging.GetLogger()}
	go s.run()
	return s
}

func (nfc *NagfluxFileCollector) Stop() {
	nfc.quit <- true
	<-nfc.quit
	nfc.log.Debug("NagfluxFileCollector stoped")
}

func (nfc NagfluxFileCollector) run() {
	for {
		select {
		case <-nfc.quit:
			nfc.quit <- true
			return
		case <-time.After(IntervalToCheckDirectory):
			for _, currentFile := range FilesInDirectoryOlderThanX(nfc.folder, MinFileAgeInSeconds) {
				data, err := ioutil.ReadFile(currentFile)
				if err != nil {
					break
				}
				for _, line := range strings.SplitAfter(string(data), "\n") {
					line = strings.TrimSpace(line)
					select {
					case <-nfc.quit:
						nfc.quit <- true
						return
					case nfc.results <- line:
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
