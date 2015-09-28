package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"strconv"
	"strings"
	"sync"
	"time"
)

//Fetches data from livestatus.
type LivestatusCacheBuilder struct {
	livestatusConnector *LivestatusConnector
	quit                chan bool
	log                 *factorlog.FactorLog
	downtimeCache       LivestatusCache
	mutex               *sync.Mutex
}

type LivestatusCache struct {
	downtime map[string]map[string]string
}

func (cache *LivestatusCache) addDowntime(host, service, start string) {
	if _, hostExists := cache.downtime[host]; !hostExists {
		cache.downtime[host] = map[string]string{service: start}
	} else if _, serviceExists := cache.downtime[host][service]; !serviceExists {
		cache.downtime[host][service] = start
	} else {
		oldTimestamp, _ := strconv.Atoi(cache.downtime[host][service])
		newTimestamp, _ := strconv.Atoi(start)
		//Take timestamp if its newer
		if oldTimestamp > newTimestamp {
			cache.downtime[host][service] = start
		}
	}
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatusCache = time.Duration(30) * time.Second
	//Livestatusquery for services in downtime.
	QueryForServicesInDowntime = `GET services
Columns: downtimes host_name display_name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
	//Livestatusquery for hosts in downtime
	QueryForHostsInDowntime = `GET hosts
Columns: downtimes name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
	QueryForDowntimeid = `GET downtimes
Columns: id start_time entry_time
OutputFormat: csv

`
)

//Constructor, which also starts it immediately.
func NewLivestatusCacheBuilder(livestatusConnector *LivestatusConnector) *LivestatusCacheBuilder {
	cache := &LivestatusCacheBuilder{livestatusConnector, make(chan bool, 2), logging.GetLogger(), LivestatusCache{make(map[string]map[string]string)}, &sync.Mutex{}}
	go cache.run()
	return cache
}

//Signals the cache to stop.
func (builder *LivestatusCacheBuilder) Stop() {
	builder.quit <- true
	<-builder.quit
	builder.log.Debug("LivestatusCacheBuilder stopped")
}

//Loop which caches livestatus downtimes and waits to quit.
func (builder *LivestatusCacheBuilder) run() {
	newCache := builder.createLivestatusCache()
	builder.mutex.Lock()
	builder.downtimeCache = newCache
	builder.mutex.Unlock()
	for {
		select {
		case <-builder.quit:
			builder.quit <- true
			return
		case <-time.After(intervalToCheckLivestatusCache):
			newCache = builder.createLivestatusCache()
			builder.mutex.Lock()
			builder.downtimeCache = newCache
			builder.mutex.Unlock()
		}
	}
}

//Builds host/service map which are in downtime
func (builder LivestatusCacheBuilder) createLivestatusCache() LivestatusCache {
	result := LivestatusCache{make(map[string]map[string]string)}
	downtimeCsv := make(chan []string)
	finishedDowntime := make(chan bool)
	hostServiceCsv := make(chan []string)
	finished := make(chan bool)
	go builder.livestatusConnector.connectToLivestatus(QueryForDowntimeid, downtimeCsv, finishedDowntime)
	go builder.livestatusConnector.connectToLivestatus(QueryForHostsInDowntime, hostServiceCsv, finished)
	go builder.livestatusConnector.connectToLivestatus(QueryForServicesInDowntime, hostServiceCsv, finished)

	jobsFinished := 0
	//contains id to starttime
	downtimes := map[string]string{}
	for jobsFinished < 2 {
		select {
		case downtimesLine := <-downtimeCsv:
			startTime, _ := strconv.Atoi(downtimesLine[1])
			entryTime, _ := strconv.Atoi(downtimesLine[2])
			latestTime := startTime
			if startTime < entryTime {
				latestTime = entryTime
			}
			for _, id := range strings.Split(downtimesLine[0], ",") {
				downtimes[id] = fmt.Sprint(latestTime)
			}
		case <-finishedDowntime:
			for jobsFinished < 2 {
				select {
				case hostService := <-hostServiceCsv:
					for _, id := range strings.Split(hostService[0], ",") {
						if len(hostService) == 2 {
							result.addDowntime(hostService[1], "", downtimes[id])
						} else if len(hostService) == 3 {
							result.addDowntime(hostService[1], hostService[2], downtimes[id])
						}
					}
				case <-finished:
					jobsFinished++
				case <-time.After(intervalToCheckLivestatusCache / 3):
					builder.log.Debug("Livestatus(host/service) timed out")
					return result
				}
			}
		case <-time.After(intervalToCheckLivestatusCache / 3):
			builder.log.Debug("Livestatus(downtimes) timed out")
			return result
		}
	}
	return result
}

//Returns true if the host/service is in downtime
func (cache LivestatusCacheBuilder) IsServiceInDowntime(host, service, time string) bool {
	result := false
	cache.mutex.Lock()
	if _, hostExists := cache.downtimeCache.downtime[host]; hostExists {
		if _, serviceExists := cache.downtimeCache.downtime[host][service]; serviceExists {
			if cache.downtimeCache.downtime[host][service] <= time {
				result = true
			}
		}
	}

	cache.mutex.Unlock()
	return result
}
