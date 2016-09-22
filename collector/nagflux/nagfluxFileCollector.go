package nagflux

import (
	"encoding/csv"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/spoolfile"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"os"
	"time"
	"github.com/griesbacher/nagflux/config"
)

//FileCollector provides a interface to nagflux, in which you could insert influxdb queries.
type FileCollector struct {
	quit           chan bool
	results        map[data.Datatype]chan collector.Printable
	folder         string
	log            *factorlog.FactorLog
	fieldSeparator rune
}

/*
table&timestamp&value&t_tag&f_field
foo&1&2&123&"""asdf"""
*/

var requiredFields = []string{"table", "time"}

//NewNagfluxFileCollector constructor, which also starts the collector.
func NewNagfluxFileCollector(results map[data.Datatype]chan collector.Printable, folder string, fieldSeparator rune) *FileCollector {
	s := &FileCollector{make(chan bool, 1), results, folder, logging.GetLogger(), fieldSeparator}
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
			pause := config.PauseNagflux.Load().(bool)
			if pause{
				logging.GetLogger().Debugln("NagfluxFileCollector in pause")
				continue
			}
			for _, currentFile := range spoolfile.FilesInDirectoryOlderThanX(nfc.folder, spoolfile.MinFileAge) {
				for _, p := range nfc.parseFile(currentFile) {
					for _, r := range nfc.results {
						select {
						case <-nfc.quit:
							nfc.quit <- true
							return
						case r <- p:
						case <-time.After(time.Duration(1) * time.Minute):
							nfc.log.Warn("NagfluxFileCollector: Could not write to buffer")
						}
					}
				}
				err := os.Remove(currentFile)
				if err != nil {
					logging.GetLogger().Warn(err)
				}
			}
		}
	}
}

func (nfc FileCollector) parseFile(filename string) []Printable {
	result := []Printable{}
	csvfile, err := os.Open(filename)
	if err != nil {
		nfc.log.Warn(err)
		return result
	}
	defer csvfile.Close()
	reader := csv.NewReader(csvfile)
	reader.Comma = nfc.fieldSeparator
	records, err := reader.ReadAll()
	if err != nil {
		nfc.log.Warn(err)
		return result
	}
	if !helper.Contains(records[0], requiredFields) {
		nfc.log.Warnf("The file %s doesn't contain all of these fields: %s", filename, requiredFields)
		return result
	}

	tagIndices := map[int]string{}
	fieldIndices := map[int]string{}

	for i, v := range records[0] {
		if len(v) > 1 && v[:2] == "t_" {
			tagIndices[i] = v[2:]
		} else if len(v) > 1 && v[:2] == "f_" {
			fieldIndices[i] = v[2:]
		} else if helper.Contains(requiredFields, []string{v}) {
			continue
		} else {
			nfc.log.Warnf("This column does not fit the requirements: %s. Tags should start with t_, fields with f_", v)
		}
	}

	for i, r := range records {
		if i == 0 {
			continue
		}
		currentPrintable := Printable{tags: map[string]string{}, fields: map[string]string{}}
		for i, v := range r {
			if v != "" {
				if records[0][i] == requiredFields[0] {
					currentPrintable.Table = v
				} else if records[0][i] == requiredFields[1] {
					currentPrintable.Timestamp = v
				} else if val, ok := tagIndices[i]; ok {
					currentPrintable.tags[val] = v
				} else if val, ok := fieldIndices[i]; ok {
					currentPrintable.fields[val] = v
				} else {
					nfc.log.Warnf("This should not happen: %s->%s", records[0][i], v)
				}
			}
		}
		result = append(result, currentPrintable)
	}
	return result
}
