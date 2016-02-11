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

//PrintForInfluxDB prints the data in influxdb lineformat
func (comment CommentData) PrintForInfluxDB(version float32) string {
	comment.sanitizeValues()
	if version >= 0.9 {
		var tags string
		if text := commentIDToText(comment.entryType); text != "" {
			tags = ",type=" + text
		}
		return comment.genInfluxLine(tags)
	}
	logging.GetLogger().Criticalf("This influxversion [%f] given in the config is not supportet", version)
	panic("")
}

//PrintForElasticsearch prints in the elasticsearch json format
func (comment CommentData) PrintForElasticsearch(version float32, index string) string {
	if version >= 2.0 {
		typ := commentIDToText(comment.entryType)
		return comment.genElasticLineWithValue(index, typ, comment.comment, comment.entryTime)
	}
	logging.GetLogger().Criticalf("This influxversion [%f] given in the config is not supportet", version)
	panic("")
}

func commentIDToText(id string) string {
	switch id {
	case "1":
		return "comment"
	case "2":
		return "downtime"
	case "3":
		return "flapping"
	case "4":
		return "acknowledgement"
	}
	logging.GetLogger().Warn("This comment type is not supported:" + id)
	return ""
}
