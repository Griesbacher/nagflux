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
	return fmt.Sprintf(`%s%s%s%s%s%s%s%s%s value=%s %s`,
		p.hostname, p.fieldseperator,
		p.service, p.fieldseperator,
		p.command, p.fieldseperator,
		p.performanceLabel, p.fieldseperator,
		p.performanceType,
		p.value, p.time)
}