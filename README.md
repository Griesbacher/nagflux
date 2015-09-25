# nagflux [![Circle CI](https://circleci.com/gh/Griesbacher/nagflux/tree/master.svg?style=svg)](https://circleci.com/gh/Griesbacher/nagflux/tree/master)
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
