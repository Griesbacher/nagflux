# nagflux
A connector which copies performancedata from Nagios/Icinga to InfluxDB

##Install
```
go get github.com/griesbacher/nagflux
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
