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
