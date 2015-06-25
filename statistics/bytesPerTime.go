package statistics

import (
	"fmt"
	"math"
	"time"
)

type QueriesPerTime struct {
	Queries int
	Time    time.Duration
}

func (queriesPerTime QueriesPerTime) String() string {
	queriesPerSecond := float64(queriesPerTime.Queries) / queriesPerTime.Time.Seconds()
	if math.IsNaN(queriesPerSecond) {
		queriesPerSecond = 0
	}

	return fmt.Sprintf("%0.2f", queriesPerSecond)
}

func (queriesPerTime *QueriesPerTime) Add(second QueriesPerTime) {
	queriesPerTime.Queries += second.Queries
	queriesPerTime.Time += second.Time
}

func (queriesPerTime *QueriesPerTime) Reset() {
	queriesPerTime.Queries = 0
	queriesPerTime.Time = time.Duration(0) * time.Nanosecond
}
