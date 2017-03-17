package data

//Datatype represents the supported databases
type Datatype string

const (
	//InfluxDB enum
	InfluxDB Datatype = "influx"
	//Elasticsearch enum
	Elasticsearch Datatype = "elastic"
	//TemplateFile enum
	JSONFile Datatype = "json"
)
