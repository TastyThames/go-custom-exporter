package proc

import (
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type LoadCollector struct {
	load1 prometheus.Gauge
}

func NewLoadCollector() *LoadCollector {
	return &LoadCollector{
		load1: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_load1",
			Help: "1-minute load average from /proc/loadavg.",
		}),
	}
}

func (l *LoadCollector) Describe(ch chan<- *prometheus.Desc) { l.load1.Describe(ch) }

func (l *LoadCollector) Collect(ch chan<- prometheus.Metric) {
	b, err := os.ReadFile("/proc/loadavg")
	if err == nil {
		fields := strings.Fields(string(b))
		if len(fields) >= 1 {
			v, _ := strconv.ParseFloat(fields[0], 64)
			l.load1.Set(v)
		}
	}
	l.load1.Collect(ch)
}
