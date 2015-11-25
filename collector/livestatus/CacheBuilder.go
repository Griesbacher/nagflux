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

//CacheBuilder fetches data from livestatus.
type CacheBuilder struct {
	livestatusConnector *Connector
	quit                chan bool
	log                 *factorlog.FactorLog
	downtimeCache       Cache
	mutex               *sync.Mutex
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatusCache = time.Duration(30) * time.Second
	//QueryForServicesInDowntime livestatusquery for services in downtime.
	QueryForServicesInDowntime = `GET services
Columns: downtimes host_name display_name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
	//QueryForHostsInDowntime livestatusquery for hosts in downtime
	QueryForHostsInDowntime = `GET hosts
Columns: downtimes name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
	//QueryForDowntimeid livestatusquery for downtime start/end
	QueryForDowntimeid = `GET downtimes
Columns: id start_time entry_time
OutputFormat: csv

`
)

//NewLivestatusCacheBuilder constructor, which also starts it immediately.
func NewLivestatusCacheBuilder(livestatusConnector *Connector) *CacheBuilder {
	cache := &CacheBuilder{livestatusConnector, make(chan bool, 2), logging.GetLogger(), Cache{make(map[string]map[string]string)}, &sync.Mutex{}}
	go cache.run(intervalToCheckLivestatusCache)
	return cache
}

//Stop signals the cache to stop.
func (builder *CacheBuilder) Stop() {
	builder.quit <- true
	<-builder.quit
	builder.log.Debug("LivestatusCacheBuilder stopped")
}

//Loop which caches livestatus downtimes and waits to quit.
func (builder *CacheBuilder) run(checkInterval time.Duration) {
	newCache := builder.createLivestatusCache()
	builder.mutex.Lock()
	builder.downtimeCache = newCache
	builder.mutex.Unlock()
	for {
		select {
		case <-builder.quit:
			builder.quit <- true
			return
		case <-time.After(checkInterval):
			newCache = builder.createLivestatusCache()
			builder.mutex.Lock()
			builder.downtimeCache = newCache
			builder.mutex.Unlock()
		}
	}
}

//Builds host/service map which are in downtime
func (builder CacheBuilder) createLivestatusCache() Cache {
	result := Cache{make(map[string]map[string]string)}
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

//IsServiceInDowntime returns true if the host/service is in downtime
func (builder CacheBuilder) IsServiceInDowntime(host, service, time string) bool {
	result := false
	builder.mutex.Lock()
	if _, hostExists := builder.downtimeCache.downtime[host]; hostExists {
		if _, serviceExists := builder.downtimeCache.downtime[host][service]; serviceExists {
			if builder.downtimeCache.downtime[host][service] <= time {
				result = true
			}
		}
	}

	builder.mutex.Unlock()
	return result
}
