package nagflux

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
)

//NagfluxPrintable converts from nagfluxfile format to X
type NagfluxPrintable struct {
	Table     string
	Timestamp string
	Value     string
	Store     map[string]string
}

func (p NagfluxPrintable) PrintForInfluxDB(version float32) string {
	line := helper.SanitizeInfluxInput(p.Table)
	cleanStore := map[string]string{}
	for k, v := range p.Store {
		cleanStore[helper.SanitizeInfluxInput(k)] = helper.SanitizeInfluxInput(v)
	}
	tags := helper.PrintMapAsString(cleanStore, ",", "=")
	if tags != "" {
		line += "," + tags
	}
	return fmt.Sprintf("%s value=%s %s", line, helper.SanitizeInfluxInput(p.Value), p.Timestamp)
}

func (p NagfluxPrintable) PrintForElasticsearch(version float32, index string) string {
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
