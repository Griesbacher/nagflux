package livestatus

//Printable this interface should be used to push data into the queue.
type Printable interface {
	Print(version float32) string
}
