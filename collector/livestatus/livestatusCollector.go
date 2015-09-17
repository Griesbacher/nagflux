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
}

const (
	//Updateinterval on livestatus data.
	intervalToCheckLivestatus = time.Duration(2) * time.Minute
	//Livestatusquery for notifications.
	QueryForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
Filter: time > %d
OutputFormat: csv

`
	//Livestatusquery for comments
	QueryForComments = `GET comments
Columns: host_name service_display_name comment entry_time author entry_type
Filter: time > %d
OutputFormat: csv

`
	//Livestatusquery for downtimes
	QueryForDowntimes = `GET downtimes
Columns: host_name service_display_name comment entry_time author end_time
Filter: time > %d
OutputFormat: csv

`
)

//Constructor, which also starts it immediately.
func NewLivestatusCollector(jobs chan interface{}, livestatusConnector *LivestatusConnector) *LivestatusCollector {
	live := &LivestatusCollector{make(chan bool, 2), jobs, livestatusConnector, logging.GetLogger()}
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
func (dump LivestatusCollector) run() {
	for {
		select {
		case <-dump.quit:
			dump.quit <- true
			return
		case <-time.After(intervalToCheckLivestatus):
			printables := make(chan Printable)
			finished := make(chan bool)
			go dump.requestPrintablesFromLivestatus(QueryForNotifications, true, printables, finished)
			go dump.requestPrintablesFromLivestatus(QueryForNotifications, true, printables, finished)
			go dump.requestPrintablesFromLivestatus(QueryForDowntimes, true, printables, finished)
			jobsFinished := 0
			for jobsFinished < 3 {
				select {
				case job := <-printables:
					dump.jobs <- job
				case <-finished:
					jobsFinished++
				}
			}
		}
	}
}

func (live LivestatusCollector) requestPrintablesFromLivestatus(query string, addTimestampToQuery bool, printables chan Printable, outerFinish chan bool) {
	queryWithTimestamp := query
	if addTimestampToQuery {
		queryWithTimestamp = fmt.Sprintf(query, time.Now().Add(intervalToCheckLivestatus/100*-150).Unix())
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
					printables <- LivestatusNotificationData{LivestatusData{line[4], "", line[7], line[1], line[2]}, line[0]}
				} else if line[0] == "SERVICE NOTIFICATION" {
					printables <- LivestatusNotificationData{LivestatusData{line[4], line[5], line[8], line[1], line[2]}, line[0]}
				} else {
					live.log.Warn("The notification type is unkown:" + line[0])
				}
			case QueryForComments:
				if len(line) == 6 {
					printables <- LivestatusCommentData{LivestatusData{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForComments out of range", line)
				}
			case QueryForDowntimes:
				if len(line) == 6 {
					printables <- LivestatusDowntimeData{LivestatusData{line[0], line[1], line[2], line[3], line[4]}, line[5]}
				} else {
					live.log.Warn("QueryForDowntimes out of range", line)
				}
			default:
				live.log.Fatal("Found unkown query type" + query)
			}
		case <-finished:
			outerFinish <- true
			return
		case <-time.After(time.Duration(10000) * time.Millisecond):
			live.log.Warn("connectToLivestatus timed out")
		}
	}
}
