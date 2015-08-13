package collector

import (
	"github.com/griesbacher/nagflux/logging"
	"io/ioutil"
	"path"
	"time"
)

const (
	MinFileAgeInSeconds      = time.Duration(60) * time.Second
	IntervalToCheckDirectory = time.Duration(5) * time.Second
)

type NagiosSpoolfileCollector struct {
	quit           chan bool
	jobs           chan string
	spoolDirectory string
	workers        []*NagiosSpoolfileWorker
}

func NagiosSpoolfileCollectorFactory(spoolDirectory string, workerAmount int, results chan interface{}, fieldseperator string) *NagiosSpoolfileCollector {
	s := &NagiosSpoolfileCollector{make(chan bool), make(chan string, 100), spoolDirectory, make([]*NagiosSpoolfileWorker, workerAmount)}

	gen := NagiosSpoolfileWorkerGenerator(s.jobs, results, fieldseperator)

	for w := 0; w < workerAmount; w++ {
		s.workers[w] = gen()
	}

	go s.run()
	return s
}

func (s *NagiosSpoolfileCollector) Stop() {
	s.quit <- true
	<-s.quit
	for _, worker := range s.workers {
		worker.Stop()
	}
	logging.GetLogger().Debug("SpoolfileCollector stopped")
}

func (s *NagiosSpoolfileCollector) run() {
	for {
		select {
		case <-s.quit:
			s.quit <- true
			return
		case <-time.After(IntervalToCheckDirectory):
			files, _ := ioutil.ReadDir(s.spoolDirectory)
			for _, currentFile := range files {
				if isItTime(currentFile.ModTime(), MinFileAgeInSeconds) {
					currentPath := path.Join(s.spoolDirectory, currentFile.Name())
					select {
					case <-s.quit:
						s.quit <- true
						return
					case s.jobs <- currentPath:
					}
				}
			}
		}
	}
}

func isItTime(timeStamp time.Time, duration time.Duration) bool {
	return time.Now().After(timeStamp.Add(duration))
}
