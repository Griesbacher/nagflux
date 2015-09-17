package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
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
	data map[string][]string
}

func (cache *LivestatusCache) addDowntime(host, service string) {
	if _, hostExists := cache.data[host]; hostExists {
		cache.data[host] = append(cache.data[host], service)
	} else {
		cache.data[host] = []string{service}
	}
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatusCache = time.Duration(30) * time.Second
	//Livestatusquery for services in downtime.
	QueryForServicesInDowntime = `GET services
Columns: host_name display_name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
	//Livestatusquery for hosts in downtime
	QueryForHostsInDowntime = `GET hosts
Columns: name
Filter: scheduled_downtime_depth > 0
OutputFormat: csv

`
)

//Constructor, which also starts it immediately.
func NewLivestatusCacheBuilder(livestatusConnector *LivestatusConnector) *LivestatusCacheBuilder {
	cache := &LivestatusCacheBuilder{livestatusConnector, make(chan bool, 2), logging.GetLogger(), LivestatusCache{make(map[string][]string)}, &sync.Mutex{}}
	go cache.run()
	return cache
}

//Signals the cache to stop.
func (live *LivestatusCacheBuilder) Stop() {
	live.quit <- true
	<-live.quit
	live.log.Debug("LivestatusCacheBuilder stoped")
}

//Loop which caches livestatus downtimes and waits to quit.
func (cache *LivestatusCacheBuilder) run() {
	newCache := cache.createLivestatusCache(QueryForHostsInDowntime, QueryForServicesInDowntime)
	cache.mutex.Lock()
	cache.downtimeCache = newCache
	cache.mutex.Unlock()
	for {
		select {
		case <-cache.quit:
			cache.quit <- true
			return
		case <-time.After(intervalToCheckLivestatusCache):
			newCache = cache.createLivestatusCache(QueryForHostsInDowntime, QueryForServicesInDowntime)
			cache.mutex.Lock()
			cache.downtimeCache = newCache
			cache.mutex.Unlock()
		}
	}
}

//Builds host/service map which are in downtime
func (cache LivestatusCacheBuilder) createLivestatusCache(hostQuery, serviceQuery string) LivestatusCache {
	result := LivestatusCache{make(map[string][]string)}

	csv := make(chan []string)
	finished := make(chan bool)
	go cache.livestatusConnector.connectToLivestatus(hostQuery, csv, finished)
	go cache.livestatusConnector.connectToLivestatus(serviceQuery, csv, finished)

	jobsFinished := 0
	for jobsFinished < 2 {
		select {
		case line := <-csv:
			if len(line) == 1 {
				result.addDowntime(line[0], "")
			} else if len(line) == 2 {
				result.addDowntime(line[0], line[1])
			}
		case <-finished:
			jobsFinished++
		}
	}
	return result
}

//Returns true if the host/service is in downtime
func (cache LivestatusCacheBuilder) IsServiceInDowntime(host, service string) bool {
	cache.mutex.Lock()
	if _, hostExists := cache.downtimeCache.data[host]; hostExists {
		for _, cachedService := range cache.downtimeCache.data[host] {
			if service == cachedService {
				cache.mutex.Unlock()
				return true
			}
		}
	}
	cache.mutex.Unlock()
	return false
}
