package statistics

import (
	"fmt"
	"math"
	"time"
)

//QueriesPerTime is value per time
type QueriesPerTime struct {
	Queries int
	Time    time.Duration
}

//String prints the object in a nice format
func (queriesPerTime QueriesPerTime) String() string {
	queriesPerSecond := float64(queriesPerTime.Queries) / queriesPerTime.Time.Seconds()
	if math.IsNaN(queriesPerSecond) {
		queriesPerSecond = 0
	}

	return fmt.Sprintf("%0.2f", queriesPerSecond)
}

//Add allows to increment the value
func (queriesPerTime *QueriesPerTime) Add(second QueriesPerTime) {
	queriesPerTime.Queries += second.Queries
	queriesPerTime.Time += second.Time
}

//Reset sets the value to zero
func (queriesPerTime *QueriesPerTime) Reset() {
	queriesPerTime.Queries = 0
	queriesPerTime.Time = time.Duration(0) * time.Nanosecond
}
