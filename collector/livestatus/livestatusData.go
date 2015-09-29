package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"strings"
)

//This interface should be used to push data into the queue.
type Printable interface {
	Print(version float32) string
}

//Contains basic data extracted from livestatusqueries.
type LivestatusData struct {
	fieldSeperator       string
	host_name            string
	service_display_name string
	comment              string
	entry_time           string
	author               string
}

//Escape all bad chars.
func (live *LivestatusData) sanitizeValues() {
	live.host_name = helper.SanitizeInfluxInput(live.host_name)
	live.service_display_name = helper.SanitizeInfluxInput(live.service_display_name)
	live.entry_time = helper.SanitizeInfluxInput(live.entry_time)
	live.author = helper.SanitizeInfluxInput(live.author)
}

//Generates the Influxdb tablename.
func (live LivestatusData) getTablename() string {
	return fmt.Sprintf("%s%s%s%smessages", live.host_name, live.fieldSeperator, live.service_display_name, live.fieldSeperator)
}

//Generates the linedata which can be parsed from influxdb
func (live LivestatusData) genInfluxLine(tags string) string {
	return live.genInfluxLineWithValue(tags, live.comment)
}

//Generates the linedata which can be parsed from influxdb
func (live LivestatusData) genInfluxLineWithValue(tags, text string) string {
	tags += ",author=" + live.author
	return fmt.Sprintf("%s%s value=\"%s\" %s", live.getTablename(), tags, text, helper.CastStringTimeFromSToMs(live.entry_time))
}

//Adds notification types to the livestatus data
type LivestatusNotificationData struct {
	LivestatusData
	notification_type  string
	notification_level string
}

func (notification *LivestatusNotificationData) sanitizeValues() {
	notification.LivestatusData.sanitizeValues()
	notification.notification_type = helper.SanitizeInfluxInput(notification.notification_type)
	notification.notification_level = helper.SanitizeInfluxInput(notification.notification_level)
}

//Prints the data in influxdb lineformat
func (notification LivestatusNotificationData) Print(version float32) string {
	notification.sanitizeValues()
	if version >= 0.9 {
		var tags string
		if notification.notification_type == "HOST\\ NOTIFICATION" {
			tags = ",type=host_notification"
		} else if notification.notification_type == "SERVICE\\ NOTIFICATION" {
			tags = ",type=service_notification"
		} else {
			logging.GetLogger().Warn("This notification type is not supported:" + notification.notification_type)
		}
		value := fmt.Sprintf("%s:<br> %s", strings.TrimSpace(notification.notification_level), notification.comment)
		return notification.genInfluxLineWithValue(tags, value)
	} else {
		logging.GetLogger().Fatalf("This influxversion [%f] given in the config is not supportet", version)
		return ""
	}
}

//Adds Comments types to the livestatus data
type LivestatusCommentData struct {
	LivestatusData
	entry_type string
}

func (comment *LivestatusCommentData) sanitizeValues() {
	comment.LivestatusData.sanitizeValues()
	comment.entry_type = helper.SanitizeInfluxInput(comment.entry_type)
}

//Prints the data in influxdb lineformat
func (comment LivestatusCommentData) Print(version float32) string {
	comment.sanitizeValues()
	if version >= 0.9 {
		var tags string
		if comment.entry_type == "1" {
			tags = ",type=comment"
		} else if comment.entry_type == "2" {
			tags = ",type=downtime"
		} else if comment.entry_type == "3" {
			tags = ",type=flapping"
		} else if comment.entry_type == "4" {
			tags = ",type=acknowledgement"
		} else {
			logging.GetLogger().Warn("This comment type is not supported:" + comment.entry_type)
		}
		return comment.genInfluxLine(tags)
	} else {
		logging.GetLogger().Fatalf("This influxversion [%f] given in the config is not supportet", version)
		return ""
	}
}

//Adds Comments types to the livestatus data
type LivestatusDowntimeData struct {
	LivestatusData
	end_time string
}

func (downtime *LivestatusDowntimeData) sanitizeValues() {
	downtime.LivestatusData.sanitizeValues()
	downtime.end_time = helper.SanitizeInfluxInput(downtime.end_time)
}

//Prints the data in influxdb lineformat
func (downtime LivestatusDowntimeData) Print(version float32) string {
	downtime.sanitizeValues()
	if version >= 0.9 {
		tags := ",type=downtime,author=" + downtime.author
		start := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime start: <br>"+downtime.comment), helper.CastStringTimeFromSToMs(downtime.entry_time))
		end := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime end: <br>"+downtime.comment), helper.CastStringTimeFromSToMs(downtime.end_time))
		return start + "\n" + end
	} else {
		logging.GetLogger().Fatalf("This influxversion [%f] given in the config is not supportet", version)
		return ""
	}
}
