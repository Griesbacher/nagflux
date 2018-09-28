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
		BufferSize             int
		FileBufferSize         int
		DefaultTarget          string
	}
	ModGearman map[string]*struct {
		Enabled    bool
		Address    string
		Queue      string
		Secret     string
		SecretFile string
		Worker     int
	}
	Log struct {
		LogFile     string
		MinSeverity string
	}
	Monitoring struct {
		PrometheusAddress string
	}
	InfluxDBGlobal struct {
		CreateDatabaseIfNotExists bool
		NastyString               string
		NastyStringToReplace      string
		HostcheckAlias            string
		ClientTimeout		      int
	}
	InfluxDB map[string]*struct {
		Enabled               bool
		Address               string
		Arguments             string
		Version               string
		StopPullingDataIfDown bool
	}
	Livestatus struct {
		Type          string
		Address       string
		MinutesToWait int
		Version       string
	}
	ElasticsearchGlobal struct {
		HostcheckAlias   string
		NumberOfShards   int
		NumberOfReplicas int
		IndexRotation    string
	}
	Elasticsearch map[string]*struct {
		Enabled bool
		Address string
		Index   string
		Version string
	}
	JSONFileExport map[string]*struct {
		Enabled               bool
		Path                  string
		AutomaticFileRotation int
	}
}
