package nagflux

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
)

//Printable converts from nagfluxfile format to X
type Printable struct {
	Table     string
	Timestamp string
	Value     string
	tags      map[string]string
	fields    map[string]string
}

//PrintForInfluxDB prints the data in influxdb lineformat
func (p Printable) PrintForInfluxDB(version float32) string {
	line := helper.SanitizeInfluxInput(p.Table)
	p.tags = helper.SanitizeMap(p.tags)
	if len(p.tags) > 0 {
		line += fmt.Sprintf(`,%s`, helper.PrintMapAsString(helper.SanitizeMap(p.tags), ",", "="))
	}
	p.fields = helper.SanitizeMap(p.fields)
	line += fmt.Sprintf(` value=%s`, p.Value)
	if len(p.fields) > 0 {
		line += fmt.Sprintf(`,%s`, helper.PrintMapAsString(helper.SanitizeMap(p.fields), ",", "="))
	}
	return fmt.Sprintf("%s %s", line, p.Timestamp)
}

//PrintForElasticsearch prints in the elasticsearch json format
func (p Printable) PrintForElasticsearch(version float32, index string) string {
	return ""
}

/*
func convertStringToX(input string, dataType data.Datatype) string {
	if _, err := strconv.ParseFloat(input, 32); err == nil {
		//Float
		return input
	} else if _, err := strconv.ParseInt(input, 10, 0); err == nil {
		//Int
		return input
	}else if _, err := strconv.ParseBool(input); err == nil {
		//Bool
		return input
	}
	//String
	if data.InfluxDB == data.InfluxDB {
		return fmt.Sprintf(`"%s"`, input)
	}
	return input
}
*/
