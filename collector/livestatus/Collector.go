package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/data"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"regexp"
	"strings"
	"time"
)

//Collector fetches data from livestatus.
type Collector struct {
	quit                chan bool
	jobs                collector.ResultQueues
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
func NewLivestatusCollector(jobs collector.ResultQueues, livestatusConnector *Connector, detectVersion string) *Collector {
	live := &Collector{
		quit:                make(chan bool, 2),
		jobs:                jobs,
		livestatusConnector: livestatusConnector,
		log:                 logging.GetLogger(),
		logQuery:            QueryNagiosForNotifications,
	}
	if detectVersion == "" {
		switch getLivestatusVersion(live) {
		case Nagios:
			live.log.Info("Livestatus type: Nagios")
		case Icinga2:
			live.log.Info("Livestatus type: Icinga2")
			live.logQuery = QueryIcinga2ForNotifications
		case Naemon:
			live.log.Info("Livestatus type: Naemon")
		}
	} else {
		switch detectVersion {
		case "Nagios":
			live.log.Info("Setting Livestatus version to: Nagios")
		case "Icinga2":
			live.log.Info("Setting Livestatus version to: Icinga2")
			live.logQuery = QueryIcinga2ForNotifications
		case "Naemon":
			live.log.Info("Setting Livestatus version to: Naemon")
		default:
			live.log.Info("Given Livestatusversion is unkown, using Nagios")
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
		case <-time.After(intervalToCheckLivestatus):
			live.log.Warn("Livestatus timed out... (Collector.queryData())")
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
				}
			case QueryIcinga2ForNotifications:
				if printable := live.handleQueryForNotifications(line); printable != nil {
					printables <- printable
				}
			case QueryForComments:
				if len(line) == 6 {
					printables <- CommentData{collector.AllFilterable, Data{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForComments out of range", line)
				}
			case QueryForDowntimes:
				if len(line) == 6 {
					printables <- DowntimeData{collector.AllFilterable, Data{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForDowntimes out of range", line)
				}
			case QueryLivestatusVersion:
				if len(line) == 1 {
					printables <- collector.SimplePrintable{Filterable: collector.AllFilterable, Text: line[0], Datatype: data.InfluxDB}
				} else {
					live.log.Warn("QueryLivestatusVersion out of range", line)
				}
			default:
				live.log.Fatal("Found unknown query type" + query)
			}
		case result := <-finished:
			outerFinish <- result
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
			//Custom: host_name, "", message, timestamp, author, notification_type, state
			return &NotificationData{collector.AllFilterable, Data{line[4], "", line[9], line[1], line[8]}, line[0], line[5]}
		} else if len(line) == 9 {
			return &NotificationData{collector.AllFilterable, Data{line[4], "", line[7], line[1], line[2]}, line[0], line[5]}
		} else if len(line) == 8 {
			return &NotificationData{collector.AllFilterable, Data{line[4], "", line[7], line[1], line[2]}, line[0], line[5]}
		} else {
			live.log.Warn("HOST NOTIFICATION, undefinded linelenght: ", len(line), " Line:", helper.SPrintStringSlice(line))
		}
	case "SERVICE NOTIFICATION":
		if len(line) == 11 {
			//Custom
			return &NotificationData{collector.AllFilterable, Data{line[4], line[5], line[10], line[1], line[9]}, line[0], line[6]}
		} else if len(line) == 10 || len(line) == 9 {
			return &NotificationData{collector.AllFilterable, Data{line[4], line[5], line[8], line[1], line[2]}, line[0], line[6]}
		} else {
			live.log.Warn("SERVICE NOTIFICATION, undefinded linelenght: ", len(line), " Line:", helper.SPrintStringSlice(line))
		}
	default:
		if strings.Contains(line[0], "NOTIFICATION SUPPRESSED") {
			live.log.Debugf("Ignoring suppressed Notification: '%s', Line: %s", line[0], helper.SPrintStringSlice(line))
		} else {
			live.log.Warnf("The notification type is unknown: '%s', whole line: '%s'", line[0], helper.SPrintStringSlice(line))
		}
	}
	return nil
}

func getLivestatusVersion(live *Collector) int {
	printables := make(chan collector.Printable, 1)
	finished := make(chan bool, 1)
	var version string
	live.requestPrintablesFromLivestatus(QueryLivestatusVersion, false, printables, finished)
	i := 0
	oneMinute := time.Duration(1) * time.Minute
	roundsToWait := config.GetConfig().Livestatus.MinutesToWait
Loop:
	for roundsToWait != 0 {
		select {
		case versionPrintable := <-printables:
			version = versionPrintable.PrintForInfluxDB("0", 0)
			break Loop
		case <-time.After(oneMinute):
			if i < roundsToWait {
				go live.requestPrintablesFromLivestatus(QueryLivestatusVersion, false, printables, finished)
			} else {
				break Loop
			}
			i++
		case fin := <-finished:
			if !fin {
				live.log.Infof(
					"Could not detect livestatus version, waiting for %s %d times( %d/%d )...",
					oneMinute, roundsToWait, i, roundsToWait,
				)
			}
		}
	}

	live.log.Info("Livestatus version: ", version)
	if icinga2, _ := regexp.MatchString(`^r[\d\.-]+$`, version); icinga2 {
		return Icinga2
	} else if nagios, _ := regexp.MatchString(`^[\d\.]+p[\d\.]+$`, version); nagios {
		return Nagios
	} else if neamon, _ := regexp.MatchString(`^[\d\.]+(-naemon)?$`, version); neamon {
		return Naemon
	}
	live.log.Warn("Could not detect livestatus type, with version: ", version, ". Asuming Nagios")
	return -1
}
