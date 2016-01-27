package livestatus

import (
	"strconv"
)

//Cache contains stored data
type Cache struct {
	downtime map[string]map[string]string
}

func (cache *Cache) addDowntime(host, service, start string) {
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
