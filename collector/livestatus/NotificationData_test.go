package livestatus

import (
	"github.com/griesbacher/nagflux/logging"
	"testing"
)

func TestSanitizeValuesNotification(t *testing.T) {
	t.Parallel()
	notification := NotificationData{Data: Data{fieldSeperator: "&", hostName: "host 1"}, notificationType: "HOST NOTIFICATION", notificationLevel: "WARN"}
	notification.sanitizeValues()

	if notification.notificationType != `HOST\ NOTIFICATION` {
		t.Errorf("The notificationType should be escaped. Expected: %s Got: %s", `HOST\ NOTIFICATION`, notification.notificationType)
	}
}

func TestPrintNotification(t *testing.T) {
	t.Parallel()
	logging.InitTestLogger()
	notification := NotificationData{Data: Data{fieldSeperator: "&", hostName: "host 1", author: "philip"}, notificationType: "HOST NOTIFICATION", notificationLevel: "WARN"}
	if !didThisPanic(notification.Print, 0.8) {
		t.Error("Printed for unsuported influxdb version but got a response")
	}

	result := notification.Print(0.9)
	if result != `host\ 1&&messages,type=host_notification,author=philip value="WARN:<br> " 000` {
		t.Error("Result does not match the expected")
	}

	notification2 := NotificationData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, notificationType: "SERVICE NOTIFICATION", notificationLevel: "WARN"}
	result2 := notification2.Print(0.9)
	if result2 != `host\ 1&service\ 1&messages,type=service_notification,author=philip value="WARN:<br> " 000` {
		t.Error("Result does not match the expected")
	}

	notification3 := NotificationData{Data: Data{fieldSeperator: "&", hostName: "host 1", serviceDisplayName: "service 1", author: "philip"}, notificationType: "NULL NOTIFICATION", notificationLevel: "WARN"}
	result3 := notification3.Print(0.9)
	if result3 != `host\ 1&service\ 1&messages,author=philip value="WARN:<br> " 000` {
		t.Error("Result does not match the expected")
	}
}

func didThisPanic(f func(float32) string, arg float32) (result bool) {
	defer func() {
		if rec := recover(); rec != nil {
			result = true
		}
	}()
	f(arg)
	return false
}
