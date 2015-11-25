[![GoDoc](https://godoc.org/github.com/Griesbacher/nagflux?status.svg)](https://godoc.org/github.com/Griesbacher/nagflux)
[![Go Report Card](http://goreportcard.com/badge/Griesbacher/nagflux)](http:/goreportcard.com/report/Griesbacher/nagflux)
[![Circle CI](https://circleci.com/gh/Griesbacher/nagflux/tree/master.svg?style=svg)](https://circleci.com/gh/Griesbacher/nagflux/tree/master)
[![Coverage Status](https://coveralls.io/repos/Griesbacher/nagflux/badge.svg?branch=master&service=github)](https://coveralls.io/github/Griesbacher/nagflux?branch=master)
# nagflux
A connector which copies performancedata from Nagios/Icinga to InfluxDB

##Install
```
go get -u github.com/griesbacher/nagflux
go build github.com/griesbacher/nagflux
```

##Start
If the configfile is in the same folder as the executable:
```
./nagflux
```
else:
```
./nagflux -configPath=/path/to/config.gcfg
```
