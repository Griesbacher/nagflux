package config

import (
	"gopkg.in/gcfg.v1"
	"sync"
)

var config Config
var mutex = &sync.Mutex{}

//InitConfig creates a config object from the give configpath
func InitConfig(configPath string) {
	var err error
	mutex.Lock()
	err = gcfg.ReadFileInto(&config, configPath)
	mutex.Unlock()
	if err != nil {
		panic(err)
	}
}

//GetConfig returns the static config object
func GetConfig() Config {
	return config
}

//InitConfigFromString creates a config object from the give configstring primary for testing
func InitConfigFromString(configString string) {
	var err error
	mutex.Lock()
	err = gcfg.ReadStringInto(&config, configString)
	mutex.Unlock()
	if err != nil {
		panic(err)
	}
}
