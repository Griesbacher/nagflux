package config

import "sync/atomic"

//PauseNagflux is used to sync the state of the influxdb
var PauseNagflux atomic.Value
