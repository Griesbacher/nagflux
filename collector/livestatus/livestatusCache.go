package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"sync"
	"time"
	"strings"
	"strconv"
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
		cache.downtime[host] = map[string]string{service:start}
	}else if _, serviceExists := cache.downtime[host][service]; !serviceExists {
		cache.downtime[host][service] = start
	}else {
		oldTimestamp ,_ := strconv.Atoi(cache.downtime[host][service])
		newTimestamp ,_  := strconv.Atoi(start)
		if  oldTimestamp > newTimestamp {
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
Columns: id start_time
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
func (live *LivestatusCacheBuilder) Stop() {
	live.quit <- true
	<-live.quit
	live.log.Debug("LivestatusCacheBuilder stopped")
}

//Loop which caches livestatus downtimes and waits to quit.
func (cache *LivestatusCacheBuilder) run() {
	newCache := cache.createLivestatusCache()
	cache.mutex.Lock()
	cache.downtimeCache = newCache
	cache.mutex.Unlock()
	for {
		select {
		case <-cache.quit:
			cache.quit <- true
			return
		case <-time.After(intervalToCheckLivestatusCache):
			newCache = cache.createLivestatusCache()
			cache.mutex.Lock()
			cache.downtimeCache = newCache
			cache.mutex.Unlock()
		}
	}
}

//Builds host/service map which are in downtime
func (cache LivestatusCacheBuilder) createLivestatusCache() LivestatusCache {
	result := LivestatusCache{make(map[string]map[string]string)}

	downtimeCsv := make(chan []string)
	hostServiceCsv := make(chan []string)
	finished := make(chan bool)
	go cache.livestatusConnector.connectToLivestatus(QueryForDowntimeid, downtimeCsv, finished)
	go cache.livestatusConnector.connectToLivestatus(QueryForHostsInDowntime, hostServiceCsv, finished)
	go cache.livestatusConnector.connectToLivestatus(QueryForServicesInDowntime, hostServiceCsv, finished)

	jobsFinished := 0
	//contains id to starttime
	downtimes := map[string]string{}
	for jobsFinished < 3 {
		select {
		case downtimesLine := <-downtimeCsv:
			for _, id := range strings.Split(downtimesLine[0], ",") {
				downtimes[id] = downtimesLine[1]
			}
		case <-finished:
			jobsFinished++
			for jobsFinished < 3 {
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
				}
			}
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
			if cache.downtimeCache.downtime[host][service] < time {
				result = true
			}
		}
	}

	cache.mutex.Unlock()
	return result
}
