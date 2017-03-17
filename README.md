[![GoDoc](https://godoc.org/github.com/Griesbacher/nagflux?status.svg)](https://godoc.org/github.com/Griesbacher/nagflux)
[![Go Report Card](http://goreportcard.com/badge/Griesbacher/nagflux)](http:/goreportcard.com/report/Griesbacher/nagflux)
[![Circle CI](https://circleci.com/gh/Griesbacher/nagflux/tree/master.svg?style=svg)](https://circleci.com/gh/Griesbacher/nagflux/tree/master)
[![Coverage Status](https://coveralls.io/repos/Griesbacher/nagflux/badge.svg?branch=master&service=github)](https://coveralls.io/github/Griesbacher/nagflux?branch=master)
# Nagflux
#### A connector which transforms performancedata from Nagios/Icinga(2)/Naemon to InfluxDB/Elasticsearch
Nagflux collects data from the NagiosSpoolfileFolder and adds informations from Livestatus. This data is sent to an InfluxDB, to get displayed by Grafana. Therefor is the tool [Histou](https://github.com/Griesbacher/histou) gives you the possibility to add Templates to Grafana.
<p>Nagflux can be seen as the process_perfdata.pl script from PNP4Nagios.</p>

## Dependencies

```
Golang 1.5+
```

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
|main|FieldSeperator|This char is used to separate the logical parts of the tablenames. This char has to be an char which is not allowed in one of those: host-, servicename, command, perfdata|
|main|FileBufferSize|This is the size of the buffer which is used to read files from disk, if you have huge checks or a lot of them you maybe recive error messages that your buffer is too small and that's the point to change it|
|Log|MinSeverity|INFO is default an enough for the most. DEBUG give you a lot more data but it's mostly just spamming|
|InfluxDBGlobal|Version|Currentliy the only supported Version of InfluxDB is 0.9+|
|Influx "name"|Address|The URL of the InfluxDB-API|
|Influx "name"|Arguments|Here you can set your user name and password as well as the database. **The precision has to be ms!**|
|Influx "name"|NastyString/NastyStringToReplace|These keys are to avoid a bug in InfluxDB and should disappear when the bug is fixed|
|Influx "name"|StopPullingDataIfDown|This is used to tell Nagflux, if this Influxdb is down to stop reading new data. That's useful if you're using spoolfiles. But if you're using gearman set this always to false because by default gearman will not buffer the data endlessly|

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

## Dataflow
There are basically two ways for Nagflux to receive data:
- Spoolfiles: They are for useful if Nagflux is running at the same machine as Nagios
- Gearman: If you have a distributed setup, that's the way to go
<p>With both ways you could enrich your performance data with additional informations from livestatus. Like downtimes, notifications and so.<p>

![Dataflow Image](https://raw.githubusercontent.com/Griesbacher/nagflux/master/doc/NagfluxDataflow.png "Nagflux Dataflow")

## OMD
Nagflux is fully integrated in [OMD-Labs](https://github.com/ConSol/omd), as well as Histou is. Therefor if you wanna try it out, it's maybe easier to install OMD-Labs.

## DEMO
This Dockercontainer contains OMD and everything is preconfigured to use Nagflux/Histou/Grafana/InfluxDB: https://github.com/Griesbacher/docker-omd-grafana

## Presentations
- Here is a presentation I held about Nagflux and Histou in 2016, only in German, sorry: [Slides](http://www.slideshare.net/PhilipGriesbacher/monitoring-workshop-kiel-2016-performancedaten-visualisierung-mit-grafana-influxdb)
- That's the first one from 2015, also only in German. [Slides](https://www.netways.de/fileadmin/images/Events_Trainings/Events/OSMC/2015/Slides_2015/Grafana_meets_Monitoring_Vorstellung_einer_Komplettloesung-Philip_Griesbacher.pdf) - [Video](https://www.youtube.com/watch?v=rY6N2H0UCFQ)