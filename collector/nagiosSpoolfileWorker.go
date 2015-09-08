package collector

import (
	"errors"
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

type NagiosSpoolfileWorker struct {
	workerId       int
	quit           chan bool
	jobs           chan string
	results        chan interface{}
	statistics     statistics.DataReceiver
	fieldseperator string
}

const hostPerfdata string = "HOSTPERFDATA"
const servicePerfdata string = "SERVICEPERFDATA"

const hostType string = "HOST"
const serviceType string = "SERVICE"

const hostname string = "HOSTNAME"
const timet string = "TIMET"
const checkcommand string = "CHECKCOMMAND"
const servicedesc string = "SERVICEDESC"

func NagiosSpoolfileWorkerGenerator(jobs chan string, results chan interface{}, fieldseperator string) func() *NagiosSpoolfileWorker {
	workerId := 0
	return func() *NagiosSpoolfileWorker {
		s := &NagiosSpoolfileWorker{workerId, make(chan bool), jobs, results, statistics.NewCmdStatisticReceiver(), fieldseperator}
		workerId++
		go s.run()
		return s
	}
}

func (w *NagiosSpoolfileWorker) Stop() {
	w.quit <- true
	<-w.quit
	logging.GetLogger().Debug("SpoolfileWorker stopped")
}

func (w *NagiosSpoolfileWorker) run() {
	var file string
	for {
		select {
		case <-w.quit:
			w.quit <- true
			return
		case file = <-w.jobs:
			startTime := time.Now()
			data, err := ioutil.ReadFile(file)
			if err != nil {
				break
			}
			lines := strings.SplitAfter(string(data), "\n")
			queries := 0
			for _, line := range lines {
				splittedPerformanceData := helper.StringToMap(line, "\t", "::")
				for singlePerfdata := range w.performanceDataIterator(splittedPerformanceData) {
					select {
					case <-w.quit:
						w.quit <- true
						return
					case w.results <- singlePerfdata:
						queries++
					}
				}
			}
			err = os.Remove(file)
			if err != nil {
				logging.GetLogger().Warn(err)
			}
			w.statistics.ReceiveQueries("read/parsed", statistics.QueriesPerTime{queries, time.Since(startTime)})
		}
	}
}

func (w *NagiosSpoolfileWorker) performanceDataIterator(input map[string]string) <-chan PerformanceData {
	regexPerformancelable, err := regexp.Compile(`([^=]+)=(U|[\d\.\-]+)([\w\/%]*);?([\d\.\-:~@]+)?;?([\d\.\-:~@]+)?;?([\d\.\-]+)?;?([\d\.\-]+)?;?\s*`)
	if err != nil {
		logging.GetLogger().Error("Regex creation failed:", err)
	}

	ch := make(chan PerformanceData)
	var typ string
	if isHostPerformanceData(input) {
		typ = hostType
	} else if isServicePerformanceData(input) {
		typ = serviceType
	} else {
		if len(input) > 1 {
			logging.GetLogger().Info("Line does not match the scheme", input)
		}
		close(ch)
		return ch
	}

	go func() {
		for _, value := range regexPerformancelable.FindAllStringSubmatch(input[typ+"PERFDATA"], -1) {
			perf := PerformanceData{
				hostname:         helper.SanitizeInfluxInput(input[hostname]),
				command:          helper.SanitizeInfluxInput(splitCommandInput(input[typ+checkcommand])),
				time:             castTimeFromSToMs(input[timet]),
				performanceLabel: helper.SanitizeInfluxInput(value[1]),
				unit:             helper.SanitizeInfluxInput(value[3]),
				fieldseperator:   w.fieldseperator,
				tags:             map[string]string{},
			}
			if typ == hostType {
				perf.service = ""
			} else {
				perf.service = helper.SanitizeInfluxInput(input[servicedesc])
			}

			for i, data := range value {
				if i > 1 && i != 3 && data != "" {
					performanceType, err := indexToperformanceType(i)
					if err != nil {
						logging.GetLogger().Warn(err, value)
						continue
					}
					if performanceType == "warn" || performanceType == "crit" {
						//Range handling
						rangeRegex := regexp.MustCompile(`[\d\.\-]+`)
						rangeHits := rangeRegex.FindAllStringSubmatch(data, -1)
						if len(rangeHits) == 1 {
							perf.tags["type"] = "normal"
							perf.tags["fill"] = "none"
							perf.value = helper.StringIntToStringFloat(rangeHits[0][0])
							perf.performanceType = performanceType
							ch <- perf
						} else if len(rangeHits) == 2 {
							//If there is a range with no infinity as border, create two points
							perf.performanceType = performanceType
							if strings.Contains(data, "@") {
								perf.tags["fill"] = "inner"
							} else {
								perf.tags["fill"] = "outer"
							}

							for i, tag := range []string{"min", "max"} {
								tmpPerf := perf
								tmpPerf.tags = helper.CopyMap(perf.tags)
								tmpPerf.tags["type"] = tag
								tmpPerf.value = helper.StringIntToStringFloat(rangeHits[i][0])
								ch <- tmpPerf
							}
						} else {
							logging.GetLogger().Warn("Regexmatching went wrong", rangeHits)
						}

					} else {
						perf.value = helper.StringIntToStringFloat(data)
						perf.performanceType = performanceType
						ch <- perf
					}
				}
			}
		}
		close(ch)
	}()
	return ch
}

func splitCommandInput(command string) string {
	return strings.Split(command, "!")[0]
}

func castTimeFromSToMs(time string) string {
	return time + "000"
}

func isHostPerformanceData(input map[string]string) bool {
	return input["DATATYPE"] == hostPerfdata
}

func isServicePerformanceData(input map[string]string) bool {
	return input["DATATYPE"] == servicePerfdata
}

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
