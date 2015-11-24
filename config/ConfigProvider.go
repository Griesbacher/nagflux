package config

import (
	"gopkg.in/gcfg.v1"
	"sync"
)

var config Config
var mutex = &sync.Mutex{}

func InitConfig(configPath string) {
	var err error
	mutex.Lock()
	err = gcfg.ReadFileInto(&config, configPath)
	mutex.Unlock()
	if err != nil {
		panic(err)
	}
}

func GetConfig() Config {
	return config
}
