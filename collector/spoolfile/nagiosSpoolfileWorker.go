package spoolfile

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/statistics"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	nagfluxTags   string = "NAGFLUX:TAG"
	nagfluxField  string = "NAGFLUX:FIELD"
	nagfluxTarget string = "NAGFLUX:TARGET"

	hostPerfdata string = "HOSTPERFDATA"

	servicePerfdata string = "SERVICEPERFDATA"

	hostType    string = "HOST"
	serviceType string = "SERVICE"

	hostname     string = "HOSTNAME"
	timet        string = "TIMET"
	checkcommand string = "CHECKCOMMAND"
	servicedesc  string = "SERVICEDESC"
)

var (
	checkMulitRegex       = regexp.MustCompile(`^(.*::)(.*)`)
	rangeRegex            = regexp.MustCompile(`[\d\.\-]+`)
	regexPerformancelable = regexp.MustCompile(`([^=]+)=(U|[\d\.,\-]+)([\w\/%]*);?([\d\.,\-:~@]+)?;?([\d\.,\-:~@]+)?;?([\d\.,\-]+)?;?([\d\.,\-]+)?;?\s*`)
	regexAltCommand       = regexp.MustCompile(`.*\[(.*)\]\s?$`)
)

//NagiosSpoolfileWorker parses the given spoolfiles and adds the extraced perfdata to the queue.
type NagiosSpoolfileWorker struct {
	workerID               int
	quit                   chan bool
	jobs                   chan string
	results                collector.ResultQueues
	livestatusCacheBuilder *livestatus.CacheBuilder
	fileBufferSize         int
	defaultTarget          collector.Filterable
}

//NewNagiosSpoolfileWorker returns a new NagiosSpoolfileWorker.
func NewNagiosSpoolfileWorker(workerID int, jobs chan string, results collector.ResultQueues,
	livestatusCacheBuilder *livestatus.CacheBuilder, fileBufferSize int, defaultTarget collector.Filterable) *NagiosSpoolfileWorker {
	return &NagiosSpoolfileWorker{
		workerID:               workerID,
		quit:                   make(chan bool),
		jobs:                   jobs,
		results:                results,
		livestatusCacheBuilder: livestatusCacheBuilder,
		fileBufferSize:         fileBufferSize,
		defaultTarget:          defaultTarget,
	}
}

//NagiosSpoolfileWorkerGenerator generates a worker and starts it.
func NagiosSpoolfileWorkerGenerator(jobs chan string, results collector.ResultQueues,
	livestatusCacheBuilder *livestatus.CacheBuilder, fileBufferSize int, defaultTarget collector.Filterable) func() *NagiosSpoolfileWorker {
	workerID := 0
	return func() *NagiosSpoolfileWorker {
		s := NewNagiosSpoolfileWorker(workerID, jobs, results, livestatusCacheBuilder, fileBufferSize, defaultTarget)
		workerID++
		go s.run()
		return s
	}
}

//Stop stops the worker
func (w *NagiosSpoolfileWorker) Stop() {
	w.quit <- true
	<-w.quit
	logging.GetLogger().Debug("SpoolfileWorker stopped")
}

//Waits for files to parse and sends the data to the main queue.
func (w *NagiosSpoolfileWorker) run() {
	promServer := statistics.GetPrometheusServer()
	var file string
	for {
		select {
		case <-w.quit:
			w.quit <- true
			return
		case file = <-w.jobs:
			promServer.SpoolFilesInQueue.Set(float64(len(w.jobs)))
			startTime := time.Now()
			logging.GetLogger().Debug("Reading file: ", file)
			filehandle, err := os.OpenFile(file, os.O_RDONLY, os.ModePerm)
			if err != nil {
				logging.GetLogger().Warn("NagiosSpoolfileWorker: Opening file error: ", err)
				break
			}
			reader := bufio.NewReaderSize(filehandle, w.fileBufferSize)
			queries := 0
			line, isPrefix, err := reader.ReadLine()
			for err == nil && !isPrefix {
				splittedPerformanceData := helper.StringToMap(string(line), "\t", "::")
				for singlePerfdata := range w.PerformanceDataIterator(splittedPerformanceData) {
					for _, r := range w.results {
						select {
						case <-w.quit:
							w.quit <- true
							return
						case r <- singlePerfdata:
							queries++
						case <-time.After(time.Duration(10) * time.Second):
							logging.GetLogger().Warn("NagiosSpoolfileWorker: Could not write to buffer")
						}
					}
				}
				line, isPrefix, err = reader.ReadLine()
			}
			if err != nil && err != io.EOF {
				logging.GetLogger().Warn(err)
			}
			if isPrefix {
				logging.GetLogger().Warn("NagiosSpoolfileWorker: filebuffer is too small")
			}
			filehandle.Close()
			err = os.Remove(file)
			if err != nil {
				logging.GetLogger().Warn(err)
			}
			timeDiff := float64(time.Since(startTime).Nanoseconds() / 1000000)
			if timeDiff >= 0 {
				promServer.SpoolFilesParsedDuration.Add(timeDiff)

			}
			if queries >= 0 {
				promServer.SpoolFilesLines.Add(float64(queries))
			}
		case <-time.After(time.Duration(5) * time.Minute):
			logging.GetLogger().Debug("NagiosSpoolfileWorker: Got nothing to do")
		}
	}
}

//PerformanceDataIterator returns an iterator to loop over generated perf data.
func (w *NagiosSpoolfileWorker) PerformanceDataIterator(input map[string]string) <-chan PerformanceData {
	ch := make(chan PerformanceData)
	typ := findType(input)
	if typ == "" {
		if len(input) > 1 {
			logging.GetLogger().Info("Line does not match the scheme", input)
		}
		close(ch)
		return ch
	}

	currentCommand := w.searchAltCommand(input[typ+"PERFDATA"], input[typ+checkcommand])
	currentTime := helper.CastStringTimeFromSToMs(input[timet])
	currentService := ""
	if typ != hostType {
		currentService = input[servicedesc]
	}

	go func() {
		perfSlice := regexPerformancelable.FindAllStringSubmatch(input[typ+"PERFDATA"], -1)
		currentCheckMultiLabel := ""
		//try to find a check_multi prefix
		if len(perfSlice) > 0 && len(perfSlice[0]) > 1 {
			currentCheckMultiLabel = getCheckMultiRegexMatch(perfSlice[0][1])
		}

	item:
		for _, value := range perfSlice {
			// Allows to add tags and fields to spoolfileentries
			tag := map[string]string{}
			if tagString, ok := input[nagfluxTags]; ok {
				tag = helper.StringToMap(tagString, " ", "=")
			}
			field := map[string]string{}
			if fieldString, ok := input[nagfluxField]; ok {
				field = helper.StringToMap(fieldString, " ", "=")
			}
			var target collector.Filterable
			if targetString, ok := input[nagfluxTarget]; ok {
				target = collector.Filterable{Filter: targetString}
			} else {
				target = collector.AllFilterable
			}

			perf := PerformanceData{
				hostname:         input[hostname],
				service:          currentService,
				command:          currentCommand,
				time:             currentTime,
				performanceLabel: value[1],
				unit:             value[3],
				tags:             tag,
				fields:           field,
				Filterable:       target,
			}

			if currentCheckMultiLabel != "" {
				//if an check_multi prefix was found last time
				//test if the current one has also one
				if potentialNextOne := getCheckMultiRegexMatch(perf.performanceLabel); potentialNextOne == "" {
					// if not put the last one in front the current
					perf.performanceLabel = currentCheckMultiLabel + perf.performanceLabel
				} else {
					// else remember the current prefix for the next one
					currentCheckMultiLabel = potentialNextOne
				}
			}

			for i, data := range value {
				data = strings.Replace(data, ",", ".", -1)
				if i > 1 && i != 3 && data != "" {
					performanceType, err := indexToperformanceType(i)
					if err != nil {
						logging.GetLogger().Warn(err, value)
						continue
					}

					//Add downtime tag if needed
					if performanceType == "value" && w.livestatusCacheBuilder != nil && w.livestatusCacheBuilder.IsServiceInDowntime(perf.hostname, perf.service, input[timet]) {
						perf.tags["downtime"] = "true"
					}

					if performanceType == "warn" || performanceType == "crit" {
						//Range handling
						fillLabel := performanceType + "-fill"
						rangeHits := rangeRegex.FindAllStringSubmatch(data, -1)
						if len(rangeHits) == 1 {
							perf.tags[fillLabel] = "none"
							perf.fields[performanceType] = helper.StringIntToStringFloat(rangeHits[0][0])

						} else if len(rangeHits) == 2 {
							//If there is a range with no infinity as border, create two points
							if strings.Contains(data, "@") {
								perf.tags[fillLabel] = "inner"
							} else {
								perf.tags[fillLabel] = "outer"
							}

							for i, tag := range []string{"min", "max"} {
								tagKey := fmt.Sprintf("%s-%s", performanceType, tag)
								perf.fields[tagKey] = helper.StringIntToStringFloat(rangeHits[i][0])
							}
						} else {
							logging.GetLogger().Warnf("Could not parse warn/crit value. Host: %v, Service: %v, Element: %v, Wholedata: %v", perf.hostname, perf.service, data, value)
						}

					} else {
						if !helper.IsStringANumber(data) {
							continue item
						}
						perf.fields[performanceType] = helper.StringIntToStringFloat(data)

					}
				}
			}
			ch <- perf
		}
		close(ch)
	}()
	return ch
}

func getCheckMultiRegexMatch(perfData string) string {
	regexResult := checkMulitRegex.FindAllStringSubmatch(perfData, -1)
	if len(regexResult) == 1 && len(regexResult[0]) == 3 {
		return regexResult[0][1]
	}
	return ""
}

func findType(input map[string]string) string {
	var typ string
	if isHostPerformanceData(input) {
		typ = hostType
	} else if isServicePerformanceData(input) {
		typ = serviceType
	}
	return typ
}

//searchAltCommand looks for alternative command name in perfdata
func (w *NagiosSpoolfileWorker) searchAltCommand(perfData, command string) string {
	result := command
	search := regexAltCommand.FindAllStringSubmatch(perfData, 1)
	if len(search) == 1 && len(search[0]) == 2 {
		result = search[0][1]
	}
	return splitCommandInput(result)
}

//Cuts the command at the first !.
func splitCommandInput(command string) string {
	return strings.Split(command, "!")[0]
}

//Tests if perfdata is of type hostperfdata.
func isHostPerformanceData(input map[string]string) bool {
	return input["DATATYPE"] == hostPerfdata
}

//Tests if perfdata is of type serviceperfdata.
func isServicePerformanceData(input map[string]string) bool {
	return input["DATATYPE"] == servicePerfdata
}

//Converts the index of the perftype to an string.
func indexToperformanceType(index int) (string, error) {
	switch index {
	case 2:
		return "value", nil
	case 4:
		return "warn", nil
	case 5:
		return "crit", nil
	case 6:
		return "min", nil
	case 7:
		return "max", nil
	default:
		return "", errors.New("Illegale index: " + strconv.Itoa(index))
	}
}
