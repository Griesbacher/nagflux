package collector

import "github.com/griesbacher/nagflux/data"

//SimplePrintable can be used to send strings as printable
type SimplePrintable struct {
	Text     string
	Datatype data.Datatype
}

func (p SimplePrintable) PrintForInfluxDB(version float32) string {
	if p.Datatype == data.InfluxDB {
		return p.Text
	}
	return ""
}

func (p SimplePrintable) PrintForElasticsearch(version float32, index string) string {
	if p.Datatype == data.Elasticsearch {
		return p.Text
	}
	return ""
}
