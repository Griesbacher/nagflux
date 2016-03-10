package config

//Config Represents the config file.
type Config struct {
	Main struct {
		NagiosSpoolfileFolder  string
		NagiosSpoolfileWorker  int
		InfluxWorker           int
		MaxInfluxWorker        int
		DumpFile               string
		NagfluxSpoolfileFolder string
		FieldSeparator         string
		BufferSize int
	}
	ModGearman struct {
		Enabled    bool
		Address    string
		Queue      string
		Secret     string
		SecretFile string
	}
	Log struct {
		LogFile     string
		MinSeverity string
	}
	Monitoring struct {
		WebserverPort string
	}
	Influx struct {
		Enabled                   bool
		Address                   string
		Arguments                 string
		Version                   string
		CreateDatabaseIfNotExists bool
		NastyString               string
		NastyStringToReplace      string
		HostcheckAlias            string
	}
	Livestatus struct {
		Type    string
		Address string
	}
	Elasticsearch struct {
		Enabled          bool
		Address          string
		Index            string
		Version          string
		HostcheckAlias   string
		NumberOfShards   int
		NumberOfReplicas int
		IndexRotation    string
	}
}
