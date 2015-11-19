package statistics

import (
	"errors"
	"fmt"
	"time"
)

//SimpleStatisticsUser basic statistic user
type SimpleStatisticsUser struct {
	data            map[string]*QueriesPerTime
	collectingSince map[string]time.Time
}

//NewSimpleStatisticsUser creates a new NewSimpleStatisticsUser
func NewSimpleStatisticsUser() *SimpleStatisticsUser {
	return &SimpleStatisticsUser{data: make(map[string]*QueriesPerTime), collectingSince: make(map[string]time.Time)}
}

//SetDataReceiver setter
func (user *SimpleStatisticsUser) SetDataReceiver(receiver DataReceiver) {
	receiver.SetStatisticsUser(user)
}

//ObtainQueries adds queries to its datastore
func (user *SimpleStatisticsUser) ObtainQueries(dataType string, monitored QueriesPerTime) {
	if _, ok := user.data[dataType]; ok {
		user.data[dataType].Add(monitored)
	} else {
		user.collectingSince[dataType] = time.Now()
		user.data[dataType] = &monitored
	}
}

//GetDataTypes returns a list of types of statistic data
func (user SimpleStatisticsUser) GetDataTypes() []string {
	dataTypes := make([]string, len(user.data))
	i := 0
	for data := range user.data {
		dataTypes[i] = data
		i++
	}
	return dataTypes
}

//GetData returns the queries per time and the duration for the queries
func (user SimpleStatisticsUser) GetData(dataType string) (QueriesPerTime, time.Duration, error) {
	if _, ok := user.data[dataType]; !ok {
		return QueriesPerTime{}, time.Duration(0), errors.New("No Data captuered so far")
	}

	data := *user.data[dataType]
	user.data[dataType].Reset()
	collectionSpan := user.collectingSince[dataType]
	user.collectingSince[dataType] = time.Now()
	return data, time.Since(collectionSpan), nil
}

//String prints the user in a readable form
func (user SimpleStatisticsUser) String() string {
	stringToPrint := ""
	for _, typ := range user.GetDataTypes() {
		statistic, duration, err := user.GetData(typ)
		if err != nil {
			stringToPrint += fmt.Sprintf("[%s]: %s [%s|%s]\n", typ, statistic, statistic.Time, duration)
		}
	}
	return stringToPrint
}
