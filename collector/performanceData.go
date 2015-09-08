package collector

import (
	"fmt"
	"strings"
)

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

func (p *PerformanceData) String() string {
	tableName := fmt.Sprintf(`%s%s%s%s%s%s%s%s%s`,
		p.hostname, p.fieldseperator,
		p.service, p.fieldseperator,
		p.command, p.fieldseperator,
		p.performanceLabel, p.fieldseperator,
		p.performanceType)
	if p.unit != "" {
		tableName += fmt.Sprintf(`,unit=%s`, p.unit)
	}

	if len(p.tags) > 0 {
		tableName += fmt.Sprintf(`,%s`, strings.Replace(strings.Replace(strings.Trim(fmt.Sprintf("%s",p.tags),"map[]")," ", ",", -1),":", "=", -1))
	}

	tableName += fmt.Sprintf(` value=%s %s`, p.value, p.time)
	return tableName
}
