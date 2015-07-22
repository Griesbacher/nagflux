package collector

import (
	"errors"
	"fmt"
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

type PerformanceData struct {
	hostname         string
	service          string
	command          string
	performanceLabel string
	performanceType  string
	unit             string
	time             string
	value            string
	fieldseperator   string
}

func (p PerformanceData) String() string {
	return fmt.Sprintf(`%s%s%s%s%s%s%s%s%s value=%s %s`,
		p.hostname, p.fieldseperator,
		p.service, p.fieldseperator,
		p.command, p.fieldseperator,
		p.performanceLabel, p.fieldseperator,
		p.performanceType,
		p.value, p.time)
}

type SpoolfileWorker struct {
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

func SpoolfileWorkerGenerator(jobs chan string, results chan interface{}, fieldseperator string) func() *SpoolfileWorker {
	workerId := 0
	return func() *SpoolfileWorker {
		s := &SpoolfileWorker{workerId, make(chan bool), jobs, results, statistics.NewCmdStatisticReceiver(), fieldseperator}
		workerId++
		go s.run()
		return s
	}
}

func (w *SpoolfileWorker) Stop() {
	w.quit <- true
	<-w.quit
	logging.GetLogger().Debug("SpoolfileWorker stopped")
}

func (w *SpoolfileWorker) run() {
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

func (w *SpoolfileWorker) performanceDataIterator(input map[string]string) <-chan PerformanceData {
	regexPerformancelable, err := regexp.Compile(`([^=]+)=(U|[\d\.\-]+)([\w\/%]*);?([\d\.\-:~@]+)?;?([\d\.\-:~@]+)?;?([\d\.\-]+)?;?([\d\.\-]+)?;?\s*`)
	if err != nil {
		logging.GetLogger().Error("Regex creation failed:", err)
	}
	var typ string
	if isHostPerformanceData(input) {
		typ = hostType
	} else if isServicePerformanceData(input) {
		typ = serviceType
	}

	ch := make(chan PerformanceData)
	go func() {
		for _, value := range regexPerformancelable.FindAllStringSubmatch(input[typ+"PERFDATA"], -1) {
			perf := PerformanceData{
				hostname:         w.cleanForInflux(input[hostname]),
				command:          w.cleanForInflux(strings.Split(input[typ+checkcommand], "!")[0]),
				time:             w.cleanForInflux(input[timet]),
				performanceLabel: w.cleanForInflux(value[1]),
				unit:             w.cleanForInflux(value[3]),
				fieldseperator:   w.fieldseperator,
			}
			if typ == hostType {
				perf.service = ""
			} else {
				perf.service = w.cleanForInflux(input[servicedesc])
			}

			for i, data := range value {
				if i > 1 && i != 3 && data != "" {
					perf.value = helper.StringIntToStringFloat(data)
					perf.performanceType, err = indexToperformanceType(i)
					ch <- perf
				}
			}
		}
		close(ch)
	}()
	return ch
}

func (w *SpoolfileWorker) cleanForInflux(input string) string {
	input = strings.Replace(input, "\\", "\\\\", -1)
	input = strings.Replace(input, " ", "\\ ", -1)
	input = strings.Replace(input, ",", "\\,", -1)
	return input
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
