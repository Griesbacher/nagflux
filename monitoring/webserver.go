package monitoring

import (
	"encoding/json"
	"fmt"
	"github.com/kdar/factorlog"
	"griesbacher.org/nagflux/helper"
	"griesbacher.org/nagflux/logging"
	"griesbacher.org/nagflux/statistics"
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"
)

type MonitoringServer struct {
	port            string
	quit            chan bool
	log             *factorlog.FactorLog
	statisticUser   *statistics.SimpleStatisticsUser
	statisticValues map[string][]int
}

var singleMonitoringServer *MonitoringServer = nil
var mutex = &sync.Mutex{}

func StartMonitoringServer(port string) *MonitoringServer {
	mutex.Lock()
	if singleMonitoringServer == nil && port != "" {
		singleMonitoringServer = &MonitoringServer{port, make(chan bool), logging.GetLogger(), statistics.NewSimpleStatisticsUser(), make(map[string][]int)}
		singleMonitoringServer.statisticUser.SetDataReceiver(statistics.NewCmdStatisticReceiver())
		go singleMonitoringServer.run()
	}
	mutex.Unlock()
	return singleMonitoringServer
}

func (server MonitoringServer) Stop() {
	server.quit <- true
	<-server.quit
	server.log.Debug("MonitoringServer stopped")
}
func (server MonitoringServer) run() {
	go server.startWebServer()
	for {
		select {
		case <-server.quit:
			server.quit <- true
			return
		case <-time.After(time.Duration(1) * time.Minute):
			server.updateStatistic()
		}
	}
}

func (server MonitoringServer) handler(w http.ResponseWriter, r *http.Request) {
	jsonData, err := json.Marshal(server.generateOutputStatistic())
	if err == nil {
		fmt.Fprintf(w, string(jsonData))
	} else {
		fmt.Fprintf(w, err.Error())
	}
}

func (server MonitoringServer) startWebServer() {
	http.HandleFunc("/", server.handler)
	http.ListenAndServe(server.port, nil)
}

func (server MonitoringServer) updateStatistic() {
	for _, key := range server.statisticUser.GetDataTypes() {
		queriesSend, _, err := server.statisticUser.GetData(key)
		if err == nil {
			server.statisticValues[key] = append([]int{queriesSend.Queries}, server.statisticValues[key]...)
			if len(server.statisticValues[key]) > 15 {
				server.statisticValues[key] = server.statisticValues[key][:15]
			}
		}
	}
}

var timeInterval = []int{1, 5, 15}

func (server MonitoringServer) generateOutputStatistic() map[string]map[string]int {
	summedData := make(map[string]map[string]int)
	for key, value := range server.statisticValues {
		for _, numberOfMinutes := range timeInterval {
			lastIndex := int(math.Min(float64(numberOfMinutes), float64(len(server.statisticValues[key])))) - 1
			if summedData[key] == nil {
				summedData[key] = make(map[string]int)
			}
			summedData[key][strconv.Itoa(numberOfMinutes)] = helper.SumIntSliceTillPos(value, lastIndex) / (lastIndex + 1)
		}
	}
	return summedData
}
