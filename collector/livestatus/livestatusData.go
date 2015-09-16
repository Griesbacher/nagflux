package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/logging"
	"strings"
)

//This interface should be used to push data into the queue.
type Printable interface {
	Print(version float32) string
}

//Contains basic data extracted from livestatusqueries.
type LivestatusData struct {
	host_name            string
	service_display_name string
	comment              string
	entry_time           string
	author               string
}

//Generates the Influxdb tablename.
func (live LivestatusData) getTablename() string {
	return fmt.Sprintf("%s&%s&messages", live.host_name, live.service_display_name)
}

//Generates the linedata which can be parsed from influxdb
func (live LivestatusData) genInfluxLine(tags string) string {
	tags += ",author=" + live.author
	return fmt.Sprintf("%s%s value=\"%s\" %s", live.getTablename(), tags, strings.TrimSpace(live.comment), live.entry_time+"000")
}

//Adds notification types to the livestatus data
type LivestatusNotificationData struct {
	LivestatusData
	notification_type string
}

//Prints the data in influxdb lineformat
func (notification LivestatusNotificationData) Print(version float32) string {
	if version >= 0.9 {
		var tags string
		if notification.notification_type == "HOST NOTIFICATION" {
			tags = ",type=host_notification"
		} else if notification.notification_type == "SERVICE NOTIFICATION" {
			tags = ",type=service_notification"
		} else {
			logging.GetLogger().Warn("This notification type is not supported:" + notification.notification_type)
		}
		return notification.genInfluxLine(tags)
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

//Prints the data in influxdb lineformat
func (comment LivestatusCommentData) Print(version float32) string {
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

//Prints the data in influxdb lineformat
func (downtime LivestatusDowntimeData) Print(version float32) string {
	if version >= 0.9 {
		tags := ",type=downtime,author=" + downtime.author
		start := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime start\n"+downtime.comment), downtime.entry_time+"000")
		end := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime end\n"+downtime.comment), downtime.end_time+"000")
		return start + "\n" + end
	} else {
		logging.GetLogger().Fatalf("This influxversion [%f] given in the config is not supportet", version)
		return ""
	}
}
