package livestatus

import (
	"github.com/griesbacher/nagflux/helper"
	"github.com/griesbacher/nagflux/logging"
)

//CommentData adds Comments types to the livestatus data
type CommentData struct {
	Data
	entryType string
}

func (comment *CommentData) sanitizeValues() {
	comment.Data.sanitizeValues()
	comment.entryType = helper.SanitizeInfluxInput(comment.entryType)
}

//Print srints the data in influxdb lineformat
func (comment CommentData) PrintForInfluxDB(version float32) string {
	comment.sanitizeValues()
	if version >= 0.9 {
		var tags string
		if comment.entryType == "1" {
			tags = ",type=comment"
		} else if comment.entryType == "2" {
			tags = ",type=downtime"
		} else if comment.entryType == "3" {
			tags = ",type=flapping"
		} else if comment.entryType == "4" {
			tags = ",type=acknowledgement"
		} else {
			logging.GetLogger().Warn("This comment type is not supported:" + comment.entryType)
		}
		return comment.genInfluxLine(tags)
	}
	logging.GetLogger().Criticalf("This influxversion [%f] given in the config is not supportet", version)
	panic("")
}

func (comment CommentData) PrintForElasticsearch(version float32, index string) string {
	return ""
}
