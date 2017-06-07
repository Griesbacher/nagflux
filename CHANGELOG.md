## v0.4.1 - 07.06.2017
### Feature
- Unicode support in Units like µs
- If performancedata contains only an U the value field will be missing, but a field unknown with the value true will be added

## v0.4.0 - 17.03.2017
### Feature
- Livestatusversion can be set in the config.
- JSONFileExport, to write collected data to an JSON File.

## v0.4.0-b1 - 16.03.2017
### Feature
- It is possible to define multiple Influxdbs, each addressed by an name, as well es Elasticsearchs, they are all called "targets". 
If the data contains an NAGFLUX:TARGET field, this one is used to direct this certain data to a specific target. 
If this tag is not added to the data, the config defines the default tag, by default "all" which means that the data is 
will be send to all targets. So "all" is a magic word and should not be uses as target name. See issue [#25](https://github.com/Griesbacher/nagflux/issues/25).
- If an Influxdb is not reachable the option "StopPullingDataIfDown" decides if reading new data into Nagflux should go on or not.

### Fix
- Some minor bugs should be fixed.

## v0.3.1 - 13.03.2017
### Fix
- Allow comma separated Performancedata
- Spoolfilebuffer is now configurable
- Nagflux Tags and Fields are ignoring empty or malformed data. Valid but ignored Performancedate would be: NAGFLUX:TAG::$_SERVICENAGFLUX_TAG$ or NAGFLUX:TAG::  

## v0.3.0 - 07.12.2016
### Fix
- Duplicate values over perflabels on one service should be fixed

## v0.2.9 - 30.11.2016
### Feature
- If an Spoolfileline contains e.g. NAGFLUX:TAG::serv=server1 ID=1	NAGFLUX:FIELD::counter_a=123 counter_b=456 the serv and ID will be stored as tag and both counters as fields. This makes it possible to add addition information to your Nagiospoolfiles.

## v0.2.8 - 23.11.2016
### Fix
- out of memory error on big files

### Feature
- Livestatustimout is configurable
- Multiple ModGearman server are supported

## v0.2.7 - 16.11.2016
### Fix
- Livestatus index out of bound error
- Less Livestatus log entries on an error(downtime)

## v0.2.6 - 26.10.2016
### Feature
- Version is shown within the help message
- check_multi prefixes will be expanded if not done by the core

## v0.2.5 - 22.09.2016
### Fix
- Deadlock when InfluxDB is not running, again !?

## v0.2.4 - 20.09.2016
### Fix
- Deadlock when InfluxDB is not running
- Pass connection args when checking for database
- Missing logfile fix

### Feature
- use of vendor-folder
- Prometheus api

### Breaks
- When using go1.5 the envvar GO15VENDOREXPERIMENT should be set to 1 

## v0.2.3 - 08.09.2016
### Fix
- ignore selfsigned ssl certs
- livestatus detection improved
- wait for influxdb on start
- pause fileparsing when influxdb is not reachable
- skip non digit perfdata(U are ignored)


## v0.2.2 - 17.05.2016
### Fix
- livestatus ServiceNotifications with just 9 entries
- nagflux fileimport exception when column name is too short
- nagios livestatus query for performance issues

## v0.2.1.1 - 08.04.2016
### Features
- mod_gearman key will be cut if it's too long

## v0.2.1 - 30.03.2016
### Features
- mod_gearman support (experimental)

### Breaks
- New Nagflux import format

## v0.2.0 - 10.03.2016

### Features
- Elasticsearch support
- New Importformat

### Fixes
-  Version Bug

### Breaks
- The old InfluxDB layout is not valid anymore. To convert the old data use the Pythonscript in CONVERTER.

## v0.1.0 - 01.12.2015

### Features
- Everything :wink:
