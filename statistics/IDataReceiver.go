package statistics

//DataReceiver interface represents someone who can handle queries and a statistics user
type DataReceiver interface {
	ReceiveQueries(string, QueriesPerTime)
	SetStatisticsUser(User)
}
