package collector

//Printable this interface should be used to push data into the queue.
type Printable interface {
	PrintForInfluxDB(version float32) string
	PrintForElasticsearch(version float32, index string) string
}
