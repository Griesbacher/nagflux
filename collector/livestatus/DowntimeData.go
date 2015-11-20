package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
	"strings"
)

//DowntimeData adds Comments types to the livestatus data
type DowntimeData struct {
	Data
	endTime string
}

func (downtime *DowntimeData) sanitizeValues() {
	downtime.Data.sanitizeValues()
	downtime.endTime = helper.SanitizeInfluxInput(downtime.endTime)
}

//Print prints the data in influxdb lineformat
func (downtime DowntimeData) Print(version float32) string {
	downtime.sanitizeValues()
	if version >= 0.9 {
		tags := ",type=downtime,author=" + downtime.author
		start := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime start: <br>"+downtime.comment), helper.CastStringTimeFromSToMs(downtime.entryTime))
		end := fmt.Sprintf("%s%s value=\"%s\" %s", downtime.getTablename(), tags, strings.TrimSpace("Downtime end: <br>"+downtime.comment), helper.CastStringTimeFromSToMs(downtime.endTime))
		return start + "\n" + end
	}
	logging.GetLogger().Fatalf("This influxversion [%f] given in the config is not supportet", version)
	return ""
}
