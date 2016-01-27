package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"time"
)

//Collector fetches data from livestatus.
type Collector struct {
	quit                chan bool
	jobs                map[string]chan collector.Printable
	livestatusConnector *Connector
	log                 *factorlog.FactorLog
	fieldSeperator      string
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatus = time.Duration(2) * time.Minute
	//QueryForNotifications livestatusquery for notifications.
	QueryForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
Filter: time < %d
Negate:
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
)

//NewLivestatusCollector constructor, which also starts it immediately.
func NewLivestatusCollector(jobs map[string]chan collector.Printable, livestatusConnector *Connector, fieldSeperator string) *Collector {
	live := &Collector{make(chan bool, 2), jobs, livestatusConnector, logging.GetLogger(), fieldSeperator}
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
	go live.requestPrintablesFromLivestatus(QueryForNotifications, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForComments, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForDowntimes, true, printables, finished)
	jobsFinished := 0
	for jobsFinished < 3*len(live.jobs) {
		for _, j := range live.jobs {
			select {
			case job := <-printables:
				j <- job
			case <-finished:
				jobsFinished++
			case <-time.After(intervalToCheckLivestatus / 3):
				live.log.Debug("requestPrintablesFromLivestatus timed out")
			}
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
			case QueryForNotifications:
				if printable := live.handleQueryForNotifications(line); printable != nil {
					printables <- printable
				} else {
					live.log.Warn("The notification type is unkown:" + line[0])
				}
			case QueryForComments:
				if len(line) == 6 {
					printables <- CommentData{Data{live.fieldSeperator, line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForComments out of range", line)
				}
			case QueryForDowntimes:
				if len(line) == 6 {
					printables <- DowntimeData{Data{live.fieldSeperator, line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForDowntimes out of range", line)
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
			return &NotificationData{Data{live.fieldSeperator, line[4], "", line[9], line[1], line[8]}, line[0], line[5]}
		} else if len(line) == 9 {
			return &NotificationData{Data{live.fieldSeperator, line[4], "", line[7], line[1], line[2]}, line[0], line[5]}
		}
	case "SERVICE NOTIFICATION":
		if len(line) == 11 {
			//Custom
			return &NotificationData{Data{live.fieldSeperator, line[4], line[5], line[10], line[1], line[9]}, line[0], line[6]}
		} else if len(line) == 10 {
			return &NotificationData{Data{live.fieldSeperator, line[4], line[5], line[8], line[1], line[2]}, line[0], line[6]}
		}

	}
	return nil
}
