package config

import (
	"github.com/griesbacher/nagflux/data"
	"sync"
)

//PauseMap is a map to store if an target requested pause or not
type PauseMap map[data.Target]bool

//pauseNagflux is used to sync the state of the influxdb
var pauseNagflux = PauseMap{}

var objMutex = &sync.Mutex{}

//IsAnyTargetOnPause will return true if any target requested pause, false otherwise
func IsAnyTargetOnPause() bool {
	objMutex.Lock()
	result := false
	for _, v := range pauseNagflux {
		if v {
			result = true
			break
		}
	}
	objMutex.Unlock()
	return result
}

func StoreValue(target data.Target, value bool) {
	objMutex.Lock()
	pauseNagflux[target] = value
	objMutex.Unlock()
}
