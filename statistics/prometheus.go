package statistics

import (
	"github.com/prometheus/client_golang/prometheus"
	"sync"
	"net"
	"net/http"
	"github.com/griesbacher/nagflux/logging"
	"time"
	"github.com/griesbacher/nagflux/collector"
	"github.com/griesbacher/nagflux/data"
)

type PrometheusServer struct {
	bufferLength             *prometheus.GaugeVec
	SpoolFilesOnDisk         prometheus.Gauge
	SpoolFilesInQueue        prometheus.Gauge
	SpoolFilesParsedDuration prometheus.Counter
	SpoolFilesParsed         prometheus.Counter
	SpoolFilesLines          prometheus.Counter
	BytesSend                *prometheus.CounterVec
	SendDuration                *prometheus.CounterVec
}

var server PrometheusServer
var p_mutex = &sync.Mutex{}
var prometheusListener net.Listener

func initServerConfig() (PrometheusServer) {
	bufferLength := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "nagflux",
			Subsystem: "main",
			Name:      "buffer_size",
			Help:      "Current Elements in Buffer",
		}, []string{"type"})
	prometheus.MustRegister(bufferLength)
	spoolFilesOnDisk := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nagflux",
			Subsystem: "spoolfile",
			Name:      "disk",
			Help:      "Nagiosspoolfiles left on disk",
		})
	prometheus.MustRegister(spoolFilesOnDisk)
	SpoolFilesInQueue := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "nagflux",
			Subsystem: "spoolfile",
			Name:      "queue",
			Help:      "Nagiosspoolfiles in queue",
		})
	prometheus.MustRegister(SpoolFilesInQueue)
	SpoolFilesParsedDuration := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "nagflux",
			Subsystem: "spoolfile",
			Name:      "files_parsed_milliseconds",
			Help:      "Nagiosspoolfiles parsed in milliseconds",
		})
	prometheus.MustRegister(SpoolFilesParsedDuration)
	SpoolFilesParsed := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "nagflux",
			Subsystem: "spoolfile",
			Name:      "files_parsed_count",
			Help:      "Nagiosspoolfiles parsed count",
		})
	prometheus.MustRegister(SpoolFilesParsed)
	SpoolFilesParsedSize := prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: "nagflux",
			Subsystem: "spoolfile",
			Name:      "parsed_lines",
			Help:      "Nagiosspoolfilelines parsed",
		})
	prometheus.MustRegister(SpoolFilesParsedSize)
	BytesSend := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nagflux",
			Subsystem: "target",
			Name:      "sent_bytes",
			Help:      "Bytes send to database",
		}, []string{"type"})
	prometheus.MustRegister(BytesSend)
	SendDuration := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "nagflux",
			Subsystem: "target",
			Name:      "sent_duration_milliseconds",
			Help:      "Time per package to sent to database",
		}, []string{"type"})
	prometheus.MustRegister(SendDuration)

	return PrometheusServer{bufferLength: bufferLength, SpoolFilesOnDisk:spoolFilesOnDisk,
		SpoolFilesInQueue:SpoolFilesInQueue, SpoolFilesParsedDuration:SpoolFilesParsedDuration,
		SpoolFilesLines:SpoolFilesParsedSize, SpoolFilesParsed:SpoolFilesParsed,
		BytesSend:BytesSend, SendDuration:SendDuration}
}

func NewPrometheusServer(address string) (PrometheusServer) {
	p_mutex.Lock()
	server = initServerConfig()
	p_mutex.Unlock()
	if address != "" {
		go func() {
			http.Handle("/metrics", prometheus.Handler())
			if err := http.ListenAndServe(address, nil); err != nil {
				logging.GetLogger().Warn(err.Error())
			}
		}()
		logging.GetLogger().Infof("serving prometheus metrics at %s/metrics", address)
	}
	return server
}

func GetPrometheusServer() PrometheusServer {
	return server
}

func (s PrometheusServer) WatchResultQueueLength(channels map[data.Datatype]chan collector.Printable) {
	go func() {
		for {
			for k, c := range channels {
				s.bufferLength.WithLabelValues(string(k)).Set(float64(len(c)))

			}
			time.Sleep(time.Duration(100 * time.Millisecond))
		}
	}()
}

func (s PrometheusServer) SetBufferLength(length float64) {
	server.bufferLength.WithLabelValues("commen").Set(length)
}

func (s PrometheusServer) Stop() {
	prometheusListener.Close()
}