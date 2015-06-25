package statistics

import "sync"

type DataReceiver interface {
	ReceiveQueries(string, QueriesPerTime)
	SetStatisticsUser(StatisticsUser)
}

type simpleReceiver struct {
	users []StatisticsUser
}

var singleReceiver *simpleReceiver = nil
var mutex = &sync.Mutex{}

func NewCmdStatisticReceiver() *simpleReceiver {
	mutex.Lock()
	if singleReceiver == nil {
		singleReceiver = new(simpleReceiver)
	}
	mutex.Unlock()
	return singleReceiver
}

func (statistic simpleReceiver) ReceiveQueries(dataType string, monitored QueriesPerTime) {
	for _, user := range statistic.users {
		user.ObtainQueries(dataType, monitored)
	}
}

func (statistic *simpleReceiver) SetStatisticsUser(user StatisticsUser) {
	statistic.users = append(statistic.users, user)
}
