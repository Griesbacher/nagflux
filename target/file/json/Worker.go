package json

import (
	"encoding/json"
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/data"
	"github.com/kdar/factorlog"
	"io/ioutil"
	"os"
	"path"
	"time"
)

type JSONFileWorker struct {
	rotationDuration time.Duration
	rotation         bool
	jobs             chan collector.Printable
	target           data.Target
	path             string
	log              *factorlog.FactorLog
	IsRunning        bool
	quit             chan bool
}

//NewJSONFileWorker creates a new JSONFileWorker
func NewJSONFileWorker(log *factorlog.FactorLog, rotation int, jobs chan collector.Printable, target data.Target, path string) *JSONFileWorker {
	w := &JSONFileWorker{
		jobs:      jobs,
		target:    target,
		path:      path,
		log:       log,
		IsRunning: true,
		quit:      make(chan bool, 2),
	}
	if _, err := os.Stat(path); err != nil {
		err = os.Mkdir(path, os.ModeDir)
		if err != nil {
			log.Panic("Creating JSON Folder err:", err)
			return nil
		}
	}
	if rotation < 0 {
		log.Criticalf("JSONFile(%s) rotation mussn't below zero %d", target, rotation)
		return nil
	} else if rotation == 0 {
		w.rotationDuration = time.Duration(1) * time.Second
		w.rotation = false
	} else {
		w.rotationDuration = time.Duration(rotation) * time.Second
		w.rotation = true
	}
	go w.run()
	return w
}

//Stop stops the Dumper.
func (t *JSONFileWorker) Stop() {
	if t.IsRunning {
		t.quit <- true
		<-t.quit
		t.IsRunning = false
		t.log.Debug("TemplateFileWorker stopped")
	}
}

func (t JSONFileWorker) run() {
	var queries []collector.Printable
	var query collector.Printable
	go func() {
		for t.IsRunning {
			select {
			case <-time.After(t.rotationDuration):
				t.writeData(queries)
				queries = queries[:0]
			}
		}
	}()
	for {
		select {
		case <-t.quit:
			t.IsRunning = false
			t.quit <- true
			return
		case query = <-t.jobs:
			if query.TestTargetFilter(t.target.Name) {
				queries = append(queries, query)
			}
		case <-time.After(time.Duration(20) * time.Second):
		}
	}
}

func (t JSONFileWorker) writeData(data []collector.Printable) {
	if len(data) == 0 {
		return
	}
	filePath := t.getFilename()
	if t.rotation {
		if _, err := os.Stat(filePath); err == nil {
			t.log.Debugf("JSON file(%s) already exists, waiting for an second", filePath)
			time.Sleep(time.Duration(1) * time.Second)
			t.writeData(data)
		}
		out, err := json.Marshal(data)
		if err != nil {
			t.log.Critical("JSON rotation marshal err:", err)
			return
		}
		err = ioutil.WriteFile(filePath, out, 0644)
		if err != nil {
			t.log.Critical("JSON rotation write err:", err)
		}
	} else {
		dataToWrite := []byte("")
		for _, d := range data {
			out, err := json.Marshal(d)
			if err != nil {
				t.log.Critical("JSON no rotation marshal err:", err)
				return
			}
			dataToWrite = append(dataToWrite, out...)
			dataToWrite = append(dataToWrite, []byte("\n")...)
		}

		if _, err := os.Stat(filePath); err == nil {
			if f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, 0600); err != nil {
				t.log.Critical(err)
			} else {
				_, err = f.Write(dataToWrite)
				if err != nil {
					t.log.Critical(err)
				}
				err = f.Close()
				if err != nil {
					t.log.Critical(err)

				}
			}
		} else {
			err = ioutil.WriteFile(filePath, dataToWrite, 0644)
			if err != nil {
				t.log.Critical("JSON no rotation write err:", err)
			}
		}
	}
}

func (t JSONFileWorker) getFilename() string {
	if t.rotation {
		return path.Join(t.path, fmt.Sprintf("perfdata_%d", time.Now().Unix()))
	}
	return path.Join(t.path, "perfdata")
}
