package nagflux

import (
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//FileCollector provides a interface to nagflux, in which you could insert influxdb queries.
type FileCollector struct {
	quit    chan bool
	results map[data.Datatype]chan collector.Printable
	folder  string
	log     *factorlog.FactorLog
}

//NewNagfluxFileCollector constructor, which also starts the collector.
func NewNagfluxFileCollector(results map[data.Datatype]chan collector.Printable, folder string) *FileCollector {
	s := &FileCollector{make(chan bool, 1), results, folder, logging.GetLogger()}
	go s.run()
	return s
}

//Stop stops the Collector.
func (nfc *FileCollector) Stop() {
	nfc.quit <- true
	<-nfc.quit
	nfc.log.Debug("NagfluxFileCollector stoped")
}

//Checks if the files are old enough, if so they will be added in the queue
func (nfc FileCollector) run() {
	for {
		select {
		case <-nfc.quit:
			nfc.quit <- true
			return
		case <-time.After(spoolfile.IntervalToCheckDirectory):
			for _, currentFile := range spoolfile.FilesInDirectoryOlderThanX(nfc.folder, spoolfile.MinFileAge) {
				data, err := ioutil.ReadFile(currentFile)
				if err != nil {
					break
				}
				for _, line := range strings.SplitAfter(string(data), "\n") {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}
					for range nfc.results {
						select {
						case <-nfc.quit:
							nfc.quit <- true
							return
						// case i <- line: //TODO: create printable
						case <-time.After(time.Duration(1) * time.Minute):
							nfc.log.Warn("NagfluxFileCollector: Could not write to buffer")
						}
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
