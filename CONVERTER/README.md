# Nagflux Convert - v0.1.0 --> v0.2.0
Dumps tables from InfluxDB to files in the Nagflux format. This makes it possible to convert from the old InfluxDB-Layout to a new one.

## Usage
```
usage: converter.py [-h] [--url URL] [--file FILE] [--target TARGET]
                    [--fieldSeparator SEPARATOR] [--hostcheckAlias ALIAS]
                    [tablenames [tablenames ...]]

Dumps Tables from InfluxDB and writes them to files

positional arguments:
  tablenames            List of tabelnames

optional arguments:
  -h, --help            show this help message and exit
  --url URL             URL to the InfluxDB with username, password...
                        Default: http://127.0.0.1:8086/query?db=mydb
  --file FILE           File with tablenames, one tablename per line
  --target TARGET       Target folder. Default: dump
  --fieldSeparator SEPARATOR
                        The fieldSeparator of nagflux genericfile format.
                        Default: &
  --hostcheckAlias ALIAS
                        The fictional name for an hostcheck. Default:
                        hostcheck
```
### Get tablenames
``` bash
$ influx -database mydb -execute 'show series' | grep "name: " | cut -c 7-
```
Store them in a file and pass the filepath with --file to the dumper.

### Example
``` bash
$ influx -database mydb -execute 'show series' | grep "name: " | cut -c 7- > influx_list
$ python converter.py --file influx_list --url 'http://InfluxDB:8086/query?db=mydb'
```
There should be a folder called dump, within on file per table.

Or dump just a few tables:
``` bash
$ python dumper.py --url 'http://InfluxDB:8086/query?db=mydb' 'mySpecialTable1' 'mySpecialTable2'
```
