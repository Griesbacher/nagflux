package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"regexp"
	"time"
)

//Collector fetches data from livestatus.
type Collector struct {
	quit                chan bool
	jobs                map[data.Datatype]chan collector.Printable
	livestatusConnector *Connector
	log                 *factorlog.FactorLog
	logQuery            string
}

const (
	//Updateinterval on livestatus data for Icinga2.
	intervalToCheckLivestatus = time.Duration(2) * time.Minute
	QueryLivestatusVersion    = `GET status
Columns: livestatus_version
OutputFormat: csv

`
	//QueryIcinga2ForNotifications livestatusquery for notifications with Icinga2 Livestatus.
	QueryIcinga2ForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
Filter: time < %d
Negate:
OutputFormat: csv

`
	//QueryNagiosForNotifications livestatusquery for notifications with nagioslike Livestatus.
	QueryNagiosForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
Filter: time > %d
OutputFormat: csv

`
	//QueryForComments livestatusquery for comments
	QueryForComments = `GET comments
Columns: host_name service_display_name comment entry_time author entry_type
Filter: entry_time > %d
OutputFormat: csv

`
	//QueryForDowntimes livestatusquery for downtimes
	QueryForDowntimes = `GET downtimes
Columns: host_name service_display_name comment entry_time author end_time
Filter: entry_time > %d
OutputFormat: csv

`
	//Nagios nagioslike Livestatus
	Nagios = iota
	//Icinga2 icinga2like Livestatus
	Icinga2 = iota
	//Naemon naemonlike Livestatus
	Naemon = iota
)

//NewLivestatusCollector constructor, which also starts it immediately.
func NewLivestatusCollector(jobs map[data.Datatype]chan collector.Printable, livestatusConnector *Connector, detectVersion bool) *Collector {
	live := &Collector{make(chan bool, 2), jobs, livestatusConnector, logging.GetLogger(), QueryNagiosForNotifications}
	if detectVersion {
		switch getLivestatusVersion(live) {
		case Nagios:
			live.log.Debug("Livestatus type Nagios")
			live.logQuery = QueryNagiosForNotifications
		case Icinga2:
			live.log.Debug("Livestatus type Icinga2")
			live.logQuery = QueryIcinga2ForNotifications
		case Naemon:
			live.log.Debug("Livestatus type Naemon")
			live.logQuery = QueryNagiosForNotifications
		}
	}
	go live.run()
	return live
}

//Stop signals the collector to stop.
func (live *Collector) Stop() {
	live.quit <- true
	<-live.quit
	live.log.Debug("LivestatusCollector stoped")
}

//Loop which checks livestats for data or waits to quit.
func (live Collector) run() {
	live.queryData()
	for {
		select {
		case <-live.quit:
			live.quit <- true
			return
		case <-time.After(intervalToCheckLivestatus):
			live.queryData()
		}
	}
}

//Queries livestatus and returns the data to the gobal queue
func (live Collector) queryData() {
	printables := make(chan collector.Printable)
	finished := make(chan bool)
	go live.requestPrintablesFromLivestatus(live.logQuery, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForComments, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForDowntimes, true, printables, finished)
	jobsFinished := 0
	for jobsFinished < 3 {
		select {
		case job := <-printables:
			for _, j := range live.jobs {
				j <- job
			}
		case <-finished:
			jobsFinished++
		case <-time.After(intervalToCheckLivestatus / 3):
			live.log.Infof("requestPrintablesFromLivestatus timed out. ")
		}
	}
}

func (live Collector) requestPrintablesFromLivestatus(query string, addTimestampToQuery bool, printables chan collector.Printable, outerFinish chan bool) {
	queryWithTimestamp := query
	if addTimestampToQuery {
		queryWithTimestamp = addTimestampToLivestatusQuery(query)
	}

	csv := make(chan []string)
	finished := make(chan bool)
	go live.livestatusConnector.connectToLivestatus(queryWithTimestamp, csv, finished)

	for {
		select {
		case line := <-csv:
			switch query {
			case QueryNagiosForNotifications:
				if printable := live.handleQueryForNotifications(line); printable != nil {
					printables <- printable
				} else {
					live.log.Warn("The notification type is unkown:" + line[0])
				}
			case QueryIcinga2ForNotifications:
				if printable := live.handleQueryForNotifications(line); printable != nil {
					printables <- printable
				} else {
					live.log.Warn("The notification type is unkown:" + line[0])
				}
			case QueryForComments:
				if len(line) == 6 {
					printables <- CommentData{Data{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForComments out of range", line)
				}
			case QueryForDowntimes:
				if len(line) == 6 {
					printables <- DowntimeData{Data{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForDowntimes out of range", line)
				}
			case QueryLivestatusVersion:
				if len(line) == 1 {
					printables <- collector.SimplePrintable{Text: line[0], Datatype: data.InfluxDB}
				} else {
					live.log.Warn("QueryLivestatusVersion out of range", line)
				}
			default:
				live.log.Fatal("Found unkown query type" + query)
			}
		case <-finished:
			outerFinish <- true
			return
		case <-time.After(intervalToCheckLivestatus / 3):
			live.log.Warn("connectToLivestatus timed out")
		}
	}
}

func addTimestampToLivestatusQuery(query string) string {
	return fmt.Sprintf(query, time.Now().Add(intervalToCheckLivestatus/100*-150).Unix())
}

func (live Collector) handleQueryForNotifications(line []string) *NotificationData {
	switch line[0] {
	case "HOST NOTIFICATION":
		if len(line) == 10 {
			//Custom
			return &NotificationData{Data{line[4], "", line[9], line[1], line[8]}, line[0], line[5]}
		} else if len(line) == 9 {
			return &NotificationData{Data{line[4], "", line[7], line[1], line[2]}, line[0], line[5]}
		} else {
			live.log.Warn("HOST NOTIFICATION, undefinded linelenght: ", len(line), " Line:", line)
		}
	case "SERVICE NOTIFICATION":
		if len(line) == 11 {
			//Custom
			return &NotificationData{Data{line[4], line[5], line[10], line[1], line[9]}, line[0], line[6]}
		} else if len(line) == 10 || len(line) == 9 {
			return &NotificationData{Data{line[4], line[5], line[8], line[1], line[2]}, line[0], line[6]}
		} else {
			live.log.Warn("SERVICE NOTIFICATION, undefinded linelenght: ", len(line), " Line:", line)
		}

	}
	return nil
}

func getLivestatusVersion(live *Collector) int {
	printables := make(chan collector.Printable, 1)
	live.requestPrintablesFromLivestatus(QueryLivestatusVersion, false, printables, make(chan bool, 1))
	var version string
	select {
	case versionPrintable := <-printables:
		version = versionPrintable.PrintForInfluxDB("0")
	case <-time.After(time.Duration(5) * time.Second):
	}

	live.log.Debug("Livestatus version: ", version)
	if icinga2, _ := regexp.MatchString(`^r[\d\.-]+$`, version); icinga2 {
		return Icinga2
	} else if nagios, _ := regexp.MatchString(`^[\d\.]+p[[\d\.]]+$`, version); nagios {
		return Nagios
	} else if neamon, _ := regexp.MatchString(`^[\d\.]+-naemon$`, version); neamon {
		return Naemon
	}
	live.log.Warn("Could not detect livestatus type, with version: ", version, " asuming Nagios")
	return -1
}
