package livestatus

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"strings"
)

//Data contains basic data extracted from livestatusqueries.
type Data struct {
	hostName           string
	serviceDisplayName string
	comment            string
	entryTime          string
	author             string
}

//Escape all bad chars.
func (live *Data) sanitizeValues() {
	live.hostName = helper.SanitizeInfluxInput(live.hostName)
	live.serviceDisplayName = helper.SanitizeInfluxInput(live.serviceDisplayName)
	live.entryTime = helper.SanitizeInfluxInput(live.entryTime)
	live.author = helper.SanitizeInfluxInput(live.author)
}

//Generates the Influxdb tablename.
func (live Data) getTablename() string {
	return fmt.Sprintf("messages,host=%s,service=%s", live.hostName, live.serviceDisplayName)
}

//Generates the linedata which can be parsed from influxdb
func (live Data) genInfluxLine(tags string) string {
	return live.genInfluxLineWithValue(tags, live.comment)
}

//Generates the linedata which can be parsed from influxdb
func (live Data) genInfluxLineWithValue(tags, text string) string {
	tags += ",author=" + live.author
	return fmt.Sprintf("%s%s message=\"%s\" %s", live.getTablename(), tags, text, helper.CastStringTimeFromSToMs(live.entryTime))
}

func (live Data) genElasticLineWithValue(index, typ, value, timestamp string) string {
	value = strings.Replace(value, `"`, `\"`, -1)
	head := fmt.Sprintf(`{"index":{"_index":"%s","_type":"messages"}}`, index) + "\n"
	data := fmt.Sprintf(`{"timestamp":%s,"message":"%s","author":"%s","host":"%s","service":"%s","type":"%s"}`+"\n",
		helper.CastStringTimeFromSToMs(timestamp), value, live.author, live.hostName, live.serviceDisplayName, typ,
	)
	return head + data
}
