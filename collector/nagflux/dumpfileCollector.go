package nagflux

import (
	"bytes"
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"io"
	"os"
	"time"
)

//DumpfileCollector collects queries from old runs, which could not been completed.
type DumpfileCollector struct {
	quit      chan bool
	jobs      chan collector.Printable
	dumpFile  string
	log       *factorlog.FactorLog
	IsRunning bool
	dataType  data.Datatype
}

func GenDumpfileName(filename string, ending data.Datatype) string {
	return fmt.Sprintf("%s.%s", filename, ending)
}

//NewDumpfileCollector constructor, which also starts the collector
func NewDumpfileCollector(jobs chan collector.Printable, dumpFile string, datatype data.Datatype) *DumpfileCollector {
	s := &DumpfileCollector{make(chan bool, 2), jobs, GenDumpfileName(dumpFile, datatype), logging.GetLogger(), true, datatype}
	go s.run()
	return s
}

//Stop stops the Collector.
func (dump *DumpfileCollector) Stop() {
	dump.quit <- true
	<-dump.quit
	dump.IsRunning = false
	dump.log.Debug("DumpfileCollector stoped")
}

//Searches for old file and parses it.
func (dump DumpfileCollector) run() {
	if _, err := os.Stat(dump.dumpFile); os.IsNotExist(err) {
		dump.log.Debugf("Dumpfile: %s not found, skipping... (Everything is fine)", dump.dumpFile)
	} else {
		if file, err := os.Open(dump.dumpFile); err != nil {
			dump.log.Warn(err)
		} else {
			buffer := bytes.NewBuffer(nil)
			dump.log.Infof("Reading Dumpfile")
			if _, err := io.Copy(buffer, file); err != nil {
				dump.log.Error(err)
			}
			file.Close()
			select {
			case <-dump.quit:
				dump.quit <- true
				return
			case dump.jobs <- collector.SimplePrintable{Text: string(buffer.Bytes()), Datatype: dump.dataType}:
				os.Remove(dump.dumpFile)
			case <-time.After(time.Duration(10) * time.Second):
				dump.log.Debugf("Timeout: %s", dump.dumpFile)
			}
		}
	}
	dump.quit <- true
}
