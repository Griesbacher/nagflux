//Contains all the config related stuff
package config

//Represents the config file.
type Config struct {
	Main struct {
		NagiosSpoolfileFolder  string
		NagiosSpoolfileWorker  int
		InfluxWorker           int
		MaxInfluxWorker        int
		DumpFile               string
		NagfluxSpoolfileFolder string
	}
	Log struct {
		LogFile     string
		MinSeverity string
	}
	Monitoring struct {
		WebserverPort string
	}
	Influx struct {
		Address                   string
		Arguments                 string
		Version                   float32
		CreateDatabaseIfNotExists bool
	}
	Grafana struct {
		FieldSeperator string
	}
	Livestatus struct {
		Type    string
		Address string
	}
}
