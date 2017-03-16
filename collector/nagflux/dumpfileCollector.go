package nagflux

import (
	"bufio"
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
	quit           chan bool
	jobs           chan collector.Printable
	dumpFile       string
	log            *factorlog.FactorLog
	IsRunning      bool
	target         data.Target
	fileBufferSize int
}

//GenDumpfileName returns the name of an dumpfile
func GenDumpfileName(filename string, ending data.Target) string {
	return fmt.Sprintf("%s-%s.%s", filename, ending.Name, ending.Datatype)
}

//NewDumpfileCollector constructor, which also starts the collector
func NewDumpfileCollector(jobs chan collector.Printable, dumpFile string, target data.Target, fileBufferSize int) *DumpfileCollector {
	s := &DumpfileCollector{
		quit:           make(chan bool, 2),
		jobs:           jobs,
		dumpFile:       GenDumpfileName(dumpFile, target),
		log:            logging.GetLogger(),
		IsRunning:      true,
		target:         target,
		fileBufferSize: fileBufferSize,
	}
	go s.run()
	return s
}

//Stop stops the Collector.
func (dump *DumpfileCollector) Stop() {
	if dump.IsRunning {
		dump.quit <- true
		<-dump.quit
		dump.IsRunning = false
		dump.log.Debug("DumpfileCollector stopped")
	}
}

//Searches for old file and parses it.
func (dump *DumpfileCollector) run() {
	if _, err := os.Stat(dump.dumpFile); os.IsNotExist(err) {
		dump.log.Debugf("Dumpfile: %s not found, skipping... (Everything is fine)", dump.dumpFile)
	} else {
		if filehandle, err := os.Open(dump.dumpFile); err != nil {
			dump.log.Warn(err)
		} else {
			dump.log.Infof("Loding dumpfile: %s", dump.dumpFile)
			if dump.target.Datatype == data.InfluxDB {
				reader := bufio.NewReaderSize(filehandle, dump.fileBufferSize)
				line, isPrefix, err := reader.ReadLine()
				for err == nil && !isPrefix {
					select {
					case <-dump.quit:
						dump.quit <- true
						return
					case dump.jobs <- collector.SimplePrintable{
						Filterable: collector.AllFilterable,
						Text:       string(line),
						Datatype:   dump.target.Datatype,
					}:
					case <-time.After(time.Duration(20) * time.Second):
						logging.GetLogger().Warn("DumpfileCollector: Could not write to buffer")
					}
					line, isPrefix, err = reader.ReadLine()
				}
				filehandle.Close()
				if err != nil && err != io.EOF {
					logging.GetLogger().Warn(err)
				}
				if isPrefix {
					logging.GetLogger().Warn("NagfluxDumpfileCollector: filebuffer is too small")
				} else {
					err = os.Remove(dump.dumpFile)
				}
			} else {
				buffer := bytes.NewBuffer(nil)
				dump.log.Infof("Reading Dumpfile")
				if _, err := io.Copy(buffer, filehandle); err != nil {
					dump.log.Error(err)
				}
				filehandle.Close()
				select {
				case <-dump.quit:
					dump.quit <- true
					return
				case dump.jobs <- collector.SimplePrintable{
					Filterable: collector.AllFilterable,
					Text:       string(buffer.Bytes()),
					Datatype:   dump.target.Datatype,
				}:
					os.Remove(dump.dumpFile)
				case <-time.After(time.Duration(10) * time.Second):
					dump.log.Debugf("Timeout: %s", dump.dumpFile)
				}
			}
		}
	}
	dump.IsRunning = false
	dump.log.Debug("DumpfileCollector stopped")
}
