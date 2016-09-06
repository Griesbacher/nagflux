package spoolfile

import (
	"errors"
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/collector/livestatus"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/griesbacher/nagflux/statistics"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//NagiosSpoolfileWorker parses the given spoolfiles and adds the extraced perfdata to the queue.
type NagiosSpoolfileWorker struct {
	workerID               int
	quit                   chan bool
	jobs                   chan string
	results                map[data.Datatype]chan collector.Printable
	statistics             statistics.DataReceiver
	livestatusCacheBuilder *livestatus.CacheBuilder
}

const hostPerfdata string = "HOSTPERFDATA"
const servicePerfdata string = "SERVICEPERFDATA"

const hostType string = "HOST"
const serviceType string = "SERVICE"

const hostname string = "HOSTNAME"
const timet string = "TIMET"
const checkcommand string = "CHECKCOMMAND"
const servicedesc string = "SERVICEDESC"

var rangeRegex = regexp.MustCompile(`[\d\.\-]+`)
var regexPerformancelable = regexp.MustCompile(`([^=]+)=(U|[\d\.\-]+)([\w\/%]*);?([\d\.\-:~@]+)?;?([\d\.\-:~@]+)?;?([\d\.\-]+)?;?([\d\.\-]+)?;?\s*`)
var regexAltCommand = regexp.MustCompile(`.*\[(.*)\]\s?$`)

//NewNagiosSpoolfileWorker returns a new NagiosSpoolfileWorker.
func NewNagiosSpoolfileWorker(workerID int, jobs chan string, results map[data.Datatype]chan collector.Printable, livestatusCacheBuilder *livestatus.CacheBuilder) *NagiosSpoolfileWorker {
	return &NagiosSpoolfileWorker{workerID, make(chan bool), jobs, results, statistics.NewCmdStatisticReceiver(), livestatusCacheBuilder}
}

//NagiosSpoolfileWorkerGenerator generates a worker and starts it.
func NagiosSpoolfileWorkerGenerator(jobs chan string, results map[data.Datatype]chan collector.Printable, livestatusCacheBuilder *livestatus.CacheBuilder) func() *NagiosSpoolfileWorker {
	workerID := 0
	return func() *NagiosSpoolfileWorker {
		s := NewNagiosSpoolfileWorker(workerID, jobs, results, livestatusCacheBuilder)
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
	var file string
	for {
		select {
		case <-w.quit:
			w.quit <- true
			return
		case file = <-w.jobs:
			startTime := time.Now()
			logging.GetLogger().Debug("Reading file: ", file)
			data, err := ioutil.ReadFile(file)
			if err != nil {
				break
			}
			lines := strings.SplitAfter(string(data), "\n")
			queries := 0
			for _, line := range lines {
				splittedPerformanceData := helper.StringToMap(line, "\t", "::")
				for singlePerfdata := range w.PerformanceDataIterator(splittedPerformanceData) {
					for _, r := range w.results {
						select {
						case <-w.quit:
							w.quit <- true
							return
						case r <- singlePerfdata:
							queries++
						case <-time.After(time.Duration(1) * time.Minute):
							logging.GetLogger().Warn("NagiosSpoolfileWorker: Could not write to buffer")
						}
					}
				}
			}
			err = os.Remove(file)
			if err != nil {
				logging.GetLogger().Warn(err)
			}
			w.statistics.ReceiveQueries("read/parsed", statistics.QueriesPerTime{Queries: queries / len(w.results), Time: time.Since(startTime)})
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

	currentCommand := w.searchAltCommand(input[typ + "PERFDATA"], input[typ + checkcommand])
	currentTime := helper.CastStringTimeFromSToMs(input[timet])
	currentService := ""
	if typ != hostType {
		currentService = input[servicedesc]
	}

	go func() {
		for _, value := range regexPerformancelable.FindAllStringSubmatch(input[typ + "PERFDATA"], -1) {
			perf := PerformanceData{
				hostname:         input[hostname],
				service:          currentService,
				command:          currentCommand,
				time:             currentTime,
				performanceLabel: value[1],
				unit:             value[3],
				tags:             map[string]string{},
				fields:           map[string]string{},
			}

			for i, data := range value {
				if i > 1 && i != 3 && data != "" {
					performanceType, err := indexToperformanceType(i)
					if err != nil {
						logging.GetLogger().Warn(err, value)
						continue
					}

					//Add downtime tag if needed
					if performanceType == "value" && w.livestatusCacheBuilder.IsServiceInDowntime(perf.hostname, perf.service, input[timet]) {
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
							logging.GetLogger().Warn("Regexmatching went wrong", rangeHits)
						}

					} else {
						if helper.IsStringANumber(data) {
							perf.fields[performanceType] = helper.StringIntToStringFloat(data)
						}

					}
				}
			}
			ch <- perf
		}
		close(ch)
	}()
	return ch
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
