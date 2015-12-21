[![GoDoc](https://godoc.org/github.com/Griesbacher/nagflux?status.svg)](https://godoc.org/github.com/Griesbacher/nagflux)
[![Go Report Card](http://goreportcard.com/badge/Griesbacher/nagflux)](http:/goreportcard.com/report/Griesbacher/nagflux)
[![Circle CI](https://circleci.com/gh/Griesbacher/nagflux/tree/master.svg?style=svg)](https://circleci.com/gh/Griesbacher/nagflux/tree/master)
[![Coverage Status](https://coveralls.io/repos/Griesbacher/nagflux/badge.svg?branch=master&service=github)](https://coveralls.io/github/Griesbacher/nagflux?branch=master)
# Nagflux
#### A connector which transforms performancedata from Nagios/Icinga(2) to InfluxDB
Nagflux collects data from the NagiosSpoolfileFolder and adds informations from Livestatus. This data is sent to an InfluxDB, to get displayed by Grafana. Therefor is the tool [Histou](https://github.com/Griesbacher/histou) gives you the possibility to add Templates to Grafana.
<p>Nagflux can be seen as the process_perfdata.pl script from PNP4Nagios.</p>



## Install
```
go get -u github.com/griesbacher/nagflux
go build github.com/griesbacher/nagflux
```

## Configure
Here are some of the important config-options:

| Section       | Config-Key    | Meaning       |
| ------------- | ------------- | ------------- |
|main|NagiosSpoolfileFolder|This is the folder where nagios/icinga writes its spoolfiles. Icinga2: `/var/spool/icinga2/perfdata`|
|main|NagfluxSpoolfileFolder|In this folder you can dump files with InfluxDBs linequery syntax, the will be shipped to the InfluxDB, the timestamp has to be in ms|
|Log|MinSeverity|INFO is default an enough for the most. DEBUG give you a lot more data but it's mostly just spamming|
|Influx|Version|Currentliy the only supported Version of InfluxDB is 0.9+|
|Influx|Address|The URL of the InfluxDB-API|
|Influx|Arguments|Here you can set your user name and password as well as the database. **The precision has to be ms!**|
|Influx|NastyString/NastyStringToReplace|These keys are to avoid a bug in InfluxDB and should disappear when the bug is fixed|
|Grafana|FieldSeperator|This char is used to separate the logical parts of the tablenames. This char has to be an char which is not allowed in one of those: host-, servicename, command, perfdata|

## Start
If the configfile is in the same folder as the executable:
```
./nagflux
```
else:
```
./nagflux -configPath=/path/to/config.gcfg
```

## Debugging
- If the InfluxDB is not available Nagflux will stop and an log entry will be written.
- If the Livestatus is not available Nagflux will just write an log entry, but additional informations can't be gathered.
- If any part of the Tablename is not valid for the InfluxDB an log entry will written and the data is writen to a file which has the same name as the logfile just with the ending '.dump-errors'. You could fix the errors by hand and copy the lines in the NagfluxSpoolfileFolder
- If the Data can't be send to the InfluxDB, Nagflux will also write them in the '.dump-errors' file, you can handle them the same way.

## OMD
Nagflux is fully integrated in [OMD-Labs](https://github.com/ConSol/omd), as well as Histou is. Therefor if you wanna try it out, it's maybe easier to install OMD-Labs.
