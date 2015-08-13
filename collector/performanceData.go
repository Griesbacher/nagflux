package collector
import "fmt"

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
}

func (p PerformanceData) String() string {
	tableName := fmt.Sprintf(`%s%s%s%s%s%s%s%s%s`,
		p.hostname, p.fieldseperator,
		p.service, p.fieldseperator,
		p.command, p.fieldseperator,
		p.performanceLabel, p.fieldseperator,
		p.performanceType)
	if p.unit != ""{
		tableName += fmt.Sprintf(`,unit=%s`,p.unit)
	}
	tableName += fmt.Sprintf(` value=%s %s`,p.value, p.time)
	return tableName
}