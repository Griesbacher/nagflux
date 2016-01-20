package spoolfile

import (
	"fmt"
	"github.com/griesbacher/nagflux/helper"
	"errors"
"strings"
)

//PerformanceData represents the nagios perfdata
type PerformanceData struct {
	hostname         string
	service          string
	command          string
	performanceLabel string
	performanceType  string
	unit             string
	time             string
	value            string
	fieldseperator   string
	tags             map[string]string
}

func (p *PerformanceData) genTablename() string {
	return fmt.Sprintf(`%s%s%s%s%s%s%s%s%s`,
		p.hostname, p.fieldseperator,
		p.service, p.fieldseperator,
		p.command, p.fieldseperator,
		p.performanceLabel, p.fieldseperator,
		p.performanceType)
}

func (p *PerformanceData) String() string {
	tableName := p.genTablename()
	if p.unit != "" {
		tableName += fmt.Sprintf(`,unit=%s`, p.unit)
	}

	if len(p.tags) > 0 {
		tableName += fmt.Sprintf(`,%s`, helper.PrintMapAsString(p.tags, ",", "="))
	}

	tableName += fmt.Sprintf(` value=%s %s`, p.value, p.time)
	return tableName
}

func (p* PerformanceData) PrintForElastic(ver float32, index string) (string, error) {
	if ver >= 2 {
		table := strings.Replace(p.genTablename(), `\`, "", -1)
		head := fmt.Sprintf(`{"index":{"_index":"%s","_type":"metric"}}`, index)+"\n"
		data := fmt.Sprintf(`{"value":%s,"@timestamp":%s,"@table":"%s"}`, p.value, p.time, table)+"\n"
		return head + data, nil
	}
	return "", errors.New("This Elasticsearch-Version is not supported")
}
