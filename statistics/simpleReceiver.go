package statistics

import "sync"

type simpleReceiver struct {
	users []User
}

var singleReceiver *simpleReceiver
var mutex = &sync.Mutex{}

//NewCmdStatisticReceiver creates a new simpleReciver
func NewCmdStatisticReceiver() DataReceiver {
	mutex.Lock()
	if singleReceiver == nil {
		singleReceiver = new(simpleReceiver)
	}
	mutex.Unlock()
	return singleReceiver
}

//ReceiveQueries sends the data to the user
func (statistic simpleReceiver) ReceiveQueries(dataType string, monitored QueriesPerTime) {
	for _, user := range statistic.users {
		user.ObtainQueries(dataType, monitored)
	}
}

//SetStatisticsUser setter
func (statistic *simpleReceiver) SetStatisticsUser(user User) {
	statistic.users = append(statistic.users, user)
}
