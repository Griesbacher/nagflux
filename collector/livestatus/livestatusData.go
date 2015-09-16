package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/logging"
	"strings"
)

type Printable interface {
	Print(version float32) string
}

type LivestatusData struct {
	host_name            string
	service_display_name string
	comment              string
	entry_time           string
	author               string
}

func (live LivestatusData) getTablename() string {
	return fmt.Sprintf("%s&%s&messages", live.host_name, live.service_display_name)
}

func (live LivestatusData) genInfluxLine(tags string) string {
	tags += ",author=" + live.author
	return fmt.Sprintf("%s%s value=\"%s\" %s", live.getTablename(), tags, strings.TrimSpace(live.comment), live.entry_time+"000")
}

type LivestatusNotificationData struct {
	LivestatusData
	notification_type string
}

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

type LivestatusCommentData struct {
	LivestatusData
	entry_type string
}

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

type LivestatusDowntimeData struct {
	LivestatusData
	end_time string
}

func (downtime LivestatusDowntimeData) Print(version float32) string {
	return ""
}
