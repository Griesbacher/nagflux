package collector

//Printable this interface should be used to push data into the queue.
type Printable interface {
	PrintForInfluxDB(version string) string
	PrintForElasticsearch(version, index string) string
	TestTargetFilter(string) bool
}
