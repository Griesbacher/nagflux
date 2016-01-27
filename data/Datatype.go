package data

//Datatype represents the supported databases
type Datatype string

const (
	InfluxDB      Datatype = "influx"
	Elasticsearch Datatype = "elastic"
)
