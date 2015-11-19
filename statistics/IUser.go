package statistics

import "time"

//User interface which represents someone who can handle statistics
type User interface {
	SetDataReceiver(DataReceiver)
	ObtainQueries(string, QueriesPerTime)
	GetData(string) (QueriesPerTime, time.Duration, error)
	GetDataTypes() []string
}
