package proc

import (
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type UptimeCollector struct {
	uptime prometheus.Gauge
}

func NewUptimeCollector() *UptimeCollector {
	return &UptimeCollector{
		uptime: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_uptime_seconds",
			Help: "System uptime seconds from /proc/uptime.",
		}),
	}
}

func readUptimeSeconds() (float64, error) {
	b, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return 0, err
	}
	parts := strings.Fields(string(b))
	if len(parts) < 1 {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseFloat(parts[0], 64)
}

func (u *UptimeCollector) Describe(ch chan<- *prometheus.Desc) { u.uptime.Describe(ch) }

func (u *UptimeCollector) Collect(ch chan<- prometheus.Metric) {
	sec, err := readUptimeSeconds()
	if err == nil {
		u.uptime.Set(sec)
	}
	u.uptime.Collect(ch)
}
