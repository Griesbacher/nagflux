package collector

import "github.com/griesbacher/nagflux/data"

//SimplePrintable can be used to send strings as printable
type SimplePrintable struct {
	Filterable
	Text     string
	Datatype data.Datatype
}

//PrintForInfluxDB generates an String for InfluxDB
func (p SimplePrintable) PrintForInfluxDB(version string) string {
	if p.Datatype == data.InfluxDB {
		return p.Text
	}
	return ""
}

//PrintForElasticsearch generates an String for Elasticsearch
func (p SimplePrintable) PrintForElasticsearch(version, index string) string {
	if p.Datatype == data.Elasticsearch {
		return p.Text
	}
	return ""
}
