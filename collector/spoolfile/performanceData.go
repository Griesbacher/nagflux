package spoolfile

import (
	"fmt"
	"github.com/griesbacher/nagflux/config"
	"github.com/griesbacher/nagflux/helper"
	"strings"
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

func (p PerformanceData) PrintForInfluxDB(version float32) string {
	if version >= 0.9 {
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
			tableName += fmt.Sprintf(`,unit="%s"`, p.unit)
		}

		tableName += fmt.Sprintf(` %s`, helper.PrintMapAsString(helper.SanitizeMap(p.fields), ",", "="))
		tableName += fmt.Sprintf(" %s\n", p.time)
		return tableName
	}
	return ""
}

func (p PerformanceData) PrintForElasticsearch(version float32, index string) string {
	if version >= 2 {
		if p.service == "" {
			p.service = "hostcheck"
		} else {
			p.service = p.service
		}
		head := fmt.Sprintf(`{"index":{"_index":"%s","_type":"metrics"}}`, index) + "\n"
		data := fmt.Sprintf(
			`{"value":%s,"@timestamp":%s,"@hostname":"%s","@service":"%s","@command":"%s","@performanceLabel":"%s"}`,
			"", //p.value,
			p.time,
			strings.Replace(p.hostname, `\`, "", -1),
			strings.Replace(p.service, `\`, "", -1),
			strings.Replace(p.command, `\`, "", -1),
			strings.Replace(p.performanceLabel, `\`, "", -1),
			//strings.Replace(p.performanceType, `\`, "", -1),
		) + "\n"
		return head + data
	}
	return ""
}
