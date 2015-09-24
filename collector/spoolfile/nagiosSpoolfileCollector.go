package spoolfile

import (
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/logging"
	"io/ioutil"
	"path"
	"time"
)

const (
	MinFileAgeInSeconds      = time.Duration(60) * time.Second
	IntervalToCheckDirectory = time.Duration(5) * time.Second
)

//Scans the nagios spoolfile folder and delegates the files to its workers.
type NagiosSpoolfileCollector struct {
	quit           chan bool
	jobs           chan string
	spoolDirectory string
	workers        []*NagiosSpoolfileWorker
}

//Creates the give amount of Woker and starts them.
func NagiosSpoolfileCollectorFactory(spoolDirectory string, workerAmount int, results chan interface{}, fieldseperator string, livestatusCacheBuilder *livestatus.LivestatusCacheBuilder) *NagiosSpoolfileCollector {
	s := &NagiosSpoolfileCollector{make(chan bool), make(chan string, 100), spoolDirectory, make([]*NagiosSpoolfileWorker, workerAmount)}

	gen := NagiosSpoolfileWorkerGenerator(s.jobs, results, fieldseperator, livestatusCacheBuilder)

	for w := 0; w < workerAmount; w++ {
		s.workers[w] = gen()
	}

	go s.run()
	return s
}

//Stops his workers and itself.
func (s *NagiosSpoolfileCollector) Stop() {
	s.quit <- true
	<-s.quit
	for _, worker := range s.workers {
		worker.Stop()
	}
	logging.GetLogger().Debug("SpoolfileCollector stopped")
}

//Delegates the files to its workers.
func (s *NagiosSpoolfileCollector) run() {
	for {
		select {
		case <-s.quit:
			s.quit <- true
			return
		case <-time.After(IntervalToCheckDirectory):
			files, _ := ioutil.ReadDir(s.spoolDirectory)
			for _, currentFile := range files {
				select {
				case <-s.quit:
					s.quit <- true
					return
				case s.jobs <- path.Join(s.spoolDirectory, currentFile.Name()):
				case <-time.After(time.Duration(1) * time.Minute):
					logging.GetLogger().Warn("NagiosSpoolfileCollector: Could not write to buffer")
				}
			}
		}
	}
}

//Returns a list of file, of a folder, names which are older then a certain duration.
func FilesInDirectoryOlderThanX(folder string, age time.Duration) []string {
	files, _ := ioutil.ReadDir(folder)
	var oldFiles []string
	for _, currentFile := range files {
		if IsItTime(currentFile.ModTime(), age) {
			oldFiles = append(oldFiles, path.Join(folder, currentFile.Name()))
		}
	}
	return oldFiles
}

//Checks if the timestamp plus duration is in the past.
func IsItTime(timeStamp time.Time, duration time.Duration) bool {
	return time.Now().After(timeStamp.Add(duration))
}
