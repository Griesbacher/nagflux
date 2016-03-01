package spoolfile

import (
	"fmt"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/helper"
)

//PerformanceData represents the nagios perfdata
type PerformanceData struct {
	hostname         string
	service          string
	command          string
	performanceLabel string
	unit             string
	time             string
	tags             map[string]string
	fields           map[string]string
}

//PrintForInfluxDB prints the data in influxdb lineformat
func (p PerformanceData) PrintForInfluxDB(version string) string {
	if helper.VersionOrdinal(version) >= helper.VersionOrdinal("0.9") {
		tableName := fmt.Sprintf(`metrics,host=%s`, helper.SanitizeInfluxInput(p.hostname))
		if p.service == "" {
			tableName += fmt.Sprintf(`,service=%s`, helper.SanitizeInfluxInput(config.GetConfig().Influx.HostcheckAlias))
		} else {
			tableName += fmt.Sprintf(`,service=%s`, helper.SanitizeInfluxInput(p.service))
		}
		tableName += fmt.Sprintf(`,command=%s,performanceLabel=%s`,
			helper.SanitizeInfluxInput(p.command),
			helper.SanitizeInfluxInput(p.performanceLabel),
		)
		if len(p.tags) > 0 {
			tableName += fmt.Sprintf(`,%s`, helper.PrintMapAsString(helper.SanitizeMap(p.tags), ",", "="))
		}
		if p.unit != "" {
			tableName += fmt.Sprintf(`,unit=%s`, p.unit)
		}

		tableName += fmt.Sprintf(` %s`, helper.PrintMapAsString(helper.SanitizeMap(p.fields), ",", "="))
		tableName += fmt.Sprintf(" %s\n", p.time)
		return tableName
	}
	return ""
}

//PrintForElasticsearch prints in the elasticsearch json format
func (p PerformanceData) PrintForElasticsearch(version, index string) string {
	if helper.VersionOrdinal(version) >= helper.VersionOrdinal("2.0") {
		if p.service == "" {
			p.service = config.GetConfig().Influx.HostcheckAlias
		}
		head := fmt.Sprintf(`{"index":{"_index":"%s","_type":"metrics"}}`, helper.GenIndex(index, p.time)) + "\n"
		data := fmt.Sprintf(
			`{"timestamp":%s,"host":"%s","service":"%s","command":"%s","performanceLabel":"%s"`,
			p.time,
			helper.SanitizeElasicInput(p.hostname),
			helper.SanitizeElasicInput(p.service),
			helper.SanitizeElasicInput(p.command),
			helper.SanitizeElasicInput(p.performanceLabel),
		)
		if p.unit != "" {
			data += fmt.Sprintf(`,"unit":"%s"`, helper.SanitizeElasicInput(p.unit))
		}
		data += helper.CreateJSONFromStringMap(p.tags)
		data += helper.CreateJSONFromStringMap(p.fields)
		data += "}\n"
		return head + data
	}
	return ""
}
