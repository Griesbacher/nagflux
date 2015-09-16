//Contains livestatus related collectors.
package livestatus

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/griesbacher/nagflux/logging"
	"github.com/kdar/factorlog"
	"net"
	"strings"
	"time"
)

//Fetches data from livestatus.
type LivestatusCollector struct {
	quit              chan bool
	jobs              chan interface{}
	livestatusAddress string
	connectionType    string
	log               *factorlog.FactorLog
}

const (
	//Updateinterval on livestatus data.
	IntervalToCheckLivestatus = time.Duration(1) * time.Minute
	//Livestatusquery for notifications.
	QueryForNotifications = `GET log
Columns: type time contact_name message
Filter: type ~ .*NOTIFICATION
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
OutputFormat: json

`
)

//Constructor, which also starts it immediately.
func NewLivestatusCollector(jobs chan interface{}, livestatusAddress, connectionType string) *LivestatusCollector {
	live := &LivestatusCollector{make(chan bool, 2), jobs, livestatusAddress, connectionType, logging.GetLogger()}
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
		case <-time.After(IntervalToCheckLivestatus):
			for _, result := range dump.queryLivestatus(QueryForNotifications) {
				dump.jobs <- result
			}
			for _, result := range dump.queryLivestatus(QueryForComments) {
				dump.jobs <- result
			}
			for _, result := range dump.queryLivestatus(QueryForDowntimes) {
				dump.jobs <- result
			}
		}
	}
}

//Queries the livestatus and returns an object which can be printed
func (live LivestatusCollector) queryLivestatus(query string) []Printable {
	queryWithTimestamp := fmt.Sprintf(query, time.Now().Add(IntervalToCheckLivestatus/100*-150).Unix())
	var csvString []string
	var conn net.Conn
	switch live.connectionType {
	case "tcp":
		conn, _ = net.Dial("tcp", live.livestatusAddress)
	case "file":
		conn, _ = net.Dial("unix", live.livestatusAddress)
	default:
		live.log.Critical("Connection type is unkown, options are: tcp, file. Input:" + live.connectionType)
		live.quit <- true
	}
	defer conn.Close()
	fmt.Fprintf(conn, queryWithTimestamp)
	reader := bufio.NewReader(conn)
	length := 1
	for length > 0 {
		message, _, _ := reader.ReadLine()
		length = len(message)
		if length > 0 {
			csvString = append(csvString, string(message))
		}
	}
	result := []Printable{}
	for _, line := range csvString {
		csvReader := csv.NewReader(strings.NewReader(line))
		csvReader.Comma = ';'
		records, err := csvReader.Read()
		if err != nil {
			live.log.Fatal(err)
		}

		switch query {
		case QueryForNotifications:
			if records[0] == "HOST NOTIFICATION" {
				result = append(result, LivestatusNotificationData{LivestatusData{records[4], "", records[7], records[1], records[2]}, records[0]})
			} else if records[0] == "SERVICE NOTIFICATION" {
				result = append(result, LivestatusNotificationData{LivestatusData{records[4], records[5], records[8], records[1], records[2]}, records[0]})
			}
		case QueryForComments:
			result = append(result, LivestatusCommentData{LivestatusData{records[0], records[1], records[2], records[3], records[4]}, records[5]})
		case QueryForDowntimes:
			result = append(result, LivestatusDowntimeData{LivestatusData{records[0], records[1], records[2], records[3], records[4]}, records[5]})
		default:
			live.log.Fatal("Found unkown query type" + query)
		}
	}
	return result
}
