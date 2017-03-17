package spoolfile

import (
	"fmt"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/helper"
)

//PerformanceData represents the nagios perfdata
type PerformanceData struct {
	collector.Filterable
	Hostname         string
	Service          string
	Command          string
	PerformanceLabel string
	Unit             string
	Time             string
	Tags             map[string]string
	Fields           map[string]string
}

//PrintForInfluxDB prints the data in influxdb lineformat
func (p PerformanceData) PrintForInfluxDB(version string) string {
	if helper.VersionOrdinal(version) >= helper.VersionOrdinal("0.9") {
		tableName := fmt.Sprintf(`metrics,host=%s`, helper.SanitizeInfluxInput(p.Hostname))
		if p.Service == "" {
			tableName += fmt.Sprintf(`,service=%s`, helper.SanitizeInfluxInput(config.GetConfig().InfluxDBGlobal.HostcheckAlias))
		} else {
			tableName += fmt.Sprintf(`,service=%s`, helper.SanitizeInfluxInput(p.Service))
		}
		tableName += fmt.Sprintf(`,command=%s,performanceLabel=%s`,
			helper.SanitizeInfluxInput(p.Command),
			helper.SanitizeInfluxInput(p.PerformanceLabel),
		)
		if len(p.Tags) > 0 {
			tableName += fmt.Sprintf(`,%s`, helper.PrintMapAsString(helper.SanitizeMap(p.Tags), ",", "="))
		}
		if p.Unit != "" {
			tableName += fmt.Sprintf(`,unit=%s`, p.Unit)
		}

		tableName += fmt.Sprintf(` %s`, helper.PrintMapAsString(helper.SanitizeMap(p.Fields), ",", "="))
		tableName += fmt.Sprintf(" %s\n", p.Time)
		return tableName
	}
	return ""
}

//PrintForElasticsearch prints in the elasticsearch json format
func (p PerformanceData) PrintForElasticsearch(version, index string) string {
	if helper.VersionOrdinal(version) >= helper.VersionOrdinal("2.0") {
		if p.Service == "" {
			p.Service = config.GetConfig().InfluxDBGlobal.HostcheckAlias
		}
		head := fmt.Sprintf(`{"index":{"_index":"%s","_type":"metrics"}}`, helper.GenIndex(index, p.Time)) + "\n"
		data := fmt.Sprintf(
			`{"timestamp":%s,"host":"%s","service":"%s","command":"%s","performanceLabel":"%s"`,
			p.Time,
			helper.SanitizeElasicInput(p.Hostname),
			helper.SanitizeElasicInput(p.Service),
			helper.SanitizeElasicInput(p.Command),
			helper.SanitizeElasicInput(p.PerformanceLabel),
		)
		if p.Unit != "" {
			data += fmt.Sprintf(`,"unit":"%s"`, helper.SanitizeElasicInput(p.Unit))
		}
		data += helper.CreateJSONFromStringMap(p.Tags)
		data += helper.CreateJSONFromStringMap(p.Fields)
		data += "}\n"
		return head + data
	}
	return ""
}
