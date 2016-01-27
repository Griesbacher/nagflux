package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"strings"
)

//NotificationData adds notification types to the livestatus data
type NotificationData struct {
	Data
	notificationType  string
	notificationLevel string
}

func (notification *NotificationData) sanitizeValues() {
	notification.Data.sanitizeValues()
	notification.notificationType = helper.SanitizeInfluxInput(notification.notificationType)
	notification.notificationLevel = helper.SanitizeInfluxInput(notification.notificationLevel)
}

//Print prints the data in influxdb lineformat
func (notification NotificationData) PrintForInfluxDB(version float32) string {
	notification.sanitizeValues()
	if version >= 0.9 {
		var tags string
		if notification.notificationType == `HOST\ NOTIFICATION` {
			tags = ",type=host_notification"
		} else if notification.notificationType == `SERVICE\ NOTIFICATION` {
			tags = ",type=service_notification"
		} else {
			logging.GetLogger().Warn("This notification type is not supported:" + notification.notificationType)
		}
		value := fmt.Sprintf("%s:<br> %s", strings.TrimSpace(notification.notificationLevel), notification.comment)
		return notification.genInfluxLineWithValue(tags, value)
	}
	logging.GetLogger().Criticalf("This influxversion [%f] given in the config is not supported", version)
	panic("")
}

func (notification NotificationData) PrintForElasticsearch(version float32, index string) string {
	return ""
}
