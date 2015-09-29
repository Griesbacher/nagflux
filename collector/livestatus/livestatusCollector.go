//Contains livestatus related collectors.
package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"time"
)

//Fetches data from livestatus.
type LivestatusCollector struct {
	quit                chan bool
	jobs                chan interface{}
	livestatusConnector *LivestatusConnector
	log                 *factorlog.FactorLog
	fieldSeperator      string
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatus = time.Duration(2) * time.Minute
	//Livestatusquery for notifications.
	QueryForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
Filter: time < %d
Negate:
OutputFormat: csv

`
	//Livestatusquery for comments
	QueryForComments = `GET comments
Columns: host_name service_display_name comment entry_time author entry_type
Filter: entry_time > %d
OutputFormat: csv

`
	//Livestatusquery for downtimes
	QueryForDowntimes = `GET downtimes
Columns: host_name service_display_name comment entry_time author end_time
Filter: entry_time > %d
OutputFormat: csv

`
)

//Constructor, which also starts it immediately.
func NewLivestatusCollector(jobs chan interface{}, livestatusConnector *LivestatusConnector, fieldSeperator string) *LivestatusCollector {
	live := &LivestatusCollector{make(chan bool, 2), jobs, livestatusConnector, logging.GetLogger(), fieldSeperator}
	go live.run()
	return live
}

//Signals the collector to stop.
func (live *LivestatusCollector) Stop() {
	live.quit <- true
	<-live.quit
	live.log.Debug("LivestatusCollector stoped")
}

//Loop which checks livestats for data or waits to quit.
func (live LivestatusCollector) run() {
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
func (live LivestatusCollector) queryData() {
	printables := make(chan Printable)
	finished := make(chan bool)
	go live.requestPrintablesFromLivestatus(QueryForNotifications, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForComments, true, printables, finished)
	go live.requestPrintablesFromLivestatus(QueryForDowntimes, true, printables, finished)
	jobsFinished := 0
	for jobsFinished < 3 {
		select {
		case job := <-printables:
			live.jobs <- job
		case <-finished:
			jobsFinished++
		case <-time.After(intervalToCheckLivestatus / 3):
			live.log.Debug("requestPrintablesFromLivestatus timed out")
		}
	}
}

func (live LivestatusCollector) requestPrintablesFromLivestatus(query string, addTimestampToQuery bool, printables chan Printable, outerFinish chan bool) {
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
				if line[0] == "HOST NOTIFICATION" {
					if len(line) == 10 {
						//Custom
						printables <- LivestatusNotificationData{LivestatusData{live.fieldSeperator, line[4], "", line[9], line[1], line[8]}, line[0], line[5]}
					} else if len(line) == 9 {
						printables <- LivestatusNotificationData{LivestatusData{live.fieldSeperator, line[4], "", line[7], line[1], line[2]}, line[0], line[5]}
					}
				} else if line[0] == "SERVICE NOTIFICATION" {
					if len(line) == 11 {
						//Custom
						printables <- LivestatusNotificationData{LivestatusData{live.fieldSeperator, line[4], line[5], line[10], line[1], line[9]}, line[0], line[6]}
					} else if len(line) == 10 {
						printables <- LivestatusNotificationData{LivestatusData{live.fieldSeperator, line[4], line[5], line[8], line[1], line[2]}, line[0], line[6]}
					}
				} else {
					live.log.Warn("The notification type is unkown:" + line[0])
				}
			case QueryForComments:
				if len(line) == 6 {
					printables <- LivestatusCommentData{LivestatusData{live.fieldSeperator, line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForComments out of range", line)
				}
			case QueryForDowntimes:
				if len(line) == 6 {
					printables <- LivestatusDowntimeData{LivestatusData{live.fieldSeperator, line[0], line[1], line[2], line[3], line[4]}, line[5]}
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
