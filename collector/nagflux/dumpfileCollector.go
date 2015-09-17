package nagflux

import (
	"bufio"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"os"
)

type DumpfileCollector struct {
	quit      chan bool
	jobs      chan interface{}
	dumpFile  string
	log       *factorlog.FactorLog
	IsRunning bool
}

func NewDumpfileCollector(jobs chan interface{}, dumpFile string) *DumpfileCollector {
	s := &DumpfileCollector{make(chan bool, 2), jobs, dumpFile, logging.GetLogger(), true}
	go s.run()
	return s
}

func (dump *DumpfileCollector) Stop() {
	dump.quit <- true
	<-dump.quit
	dump.IsRunning = false
	dump.log.Debug("DumpfileCollector stoped")
}

func (dump DumpfileCollector) run() {
	if _, err := os.Stat(dump.dumpFile); os.IsNotExist(err) {
		dump.log.Infof("Dumpfile: %s not found, skipping... (Everything is fine)", dump.dumpFile)
	} else {
		if file, err := os.Open(dump.dumpFile); err != nil {
			dump.log.Warn(err)
		} else {
			dump.log.Infof("Reading Dumpfile")
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				select {
				case <-dump.quit:
					dump.quit <- true
					return
				case dump.jobs <- scanner.Text():
				}
			}
			os.Remove(dump.dumpFile)
		}
	}
	dump.quit <- true
}
