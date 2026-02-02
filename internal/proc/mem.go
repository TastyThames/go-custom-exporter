package proc

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type MemCollector struct {
	total prometheus.Gauge
	avail prometheus.Gauge
}

func NewMemCollector() *MemCollector {
	return &MemCollector{
		total: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_mem_total_bytes",
			Help: "Total memory in bytes from /proc/meminfo.",
		}),
		avail: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_mem_available_bytes",
			Help: "Available memory in bytes from /proc/meminfo.",
		}),
	}
}

func readMeminfo() (totalBytes, availBytes uint64, err error) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	var totalKB, availKB uint64
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "MemTotal:") {
			totalKB = parseKB(line)
		}
		if strings.HasPrefix(line, "MemAvailable:") {
			availKB = parseKB(line)
		}
	}
	if err := sc.Err(); err != nil {
		return 0, 0, err
	}

	return totalKB * 1024, availKB * 1024, nil
}

func parseKB(line string) uint64 {
	parts := strings.Fields(line)
	if len(parts) < 2 {
		return 0
	}
	v, _ := strconv.ParseUint(parts[1], 10, 64)
	return v
}

func (m *MemCollector) Describe(ch chan<- *prometheus.Desc) {
	m.total.Describe(ch)
	m.avail.Describe(ch)
}

func (m *MemCollector) Collect(ch chan<- prometheus.Metric) {
	total, avail, err := readMeminfo()
	if err == nil && total > 0 {
		m.total.Set(float64(total))
		m.avail.Set(float64(avail))
	}
	m.total.Collect(ch)
	m.avail.Collect(ch)
}
