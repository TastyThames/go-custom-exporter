package proc

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// cpuTimes stores aggregated CPU jiffies from /proc/stat
type cpuTimes struct {
	user, nice, system, idle, iowait, irq, softirq, steal uint64
}

func (c cpuTimes) total() uint64 {
	return c.user + c.nice + c.system + c.idle + c.iowait + c.irq + c.softirq + c.steal
}

func readProcStatCPU() (cpuTimes, error) {
	f, err := os.Open("/proc/stat")
	if err != nil {
		return cpuTimes{}, err
	}
	defer f.Close()

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "cpu ") {
			parts := strings.Fields(line)

			parse := func(i int) uint64 {
				if i >= len(parts) {
					return 0
				}
				v, _ := strconv.ParseUint(parts[i], 10, 64)
				return v
			}

			return cpuTimes{
				user:    parse(1),
				nice:    parse(2),
				system:  parse(3),
				idle:    parse(4),
				iowait:  parse(5),
				irq:     parse(6),
				softirq: parse(7),
				steal:   parse(8),
			}, nil
		}
	}
	return cpuTimes{}, sc.Err()
}

// CPUCollector computes CPU usage percent using two /proc/stat samples.
type CPUCollector struct {
	usage prometheus.Gauge

	mu   sync.RWMutex
	last cpuTimes
}

func NewCPUCollector(sampleEvery time.Duration) *CPUCollector {
	c := &CPUCollector{
		usage: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_cpu_usage_percent",
			Help: "CPU usage percent computed from /proc/stat (all CPUs aggregated).",
		}),
	}

	base, _ := readProcStatCPU()
	c.last = base

	go func() {
		t := time.NewTicker(sampleEvery)
		defer t.Stop()

		for range t.C {
			now, err := readProcStatCPU()
			if err != nil {
				continue
			}

			prev := c.last
			totalDelta := float64(now.total() - prev.total())
			idleDelta := float64((now.idle + now.iowait) - (prev.idle + prev.iowait))

			var usage float64
			if totalDelta > 0 {
				usage = (1.0 - (idleDelta / totalDelta)) * 100.0
				if usage < 0 {
					usage = 0
				}
				if usage > 100 {
					usage = 100
				}
			}

			c.mu.Lock()
			c.last = now
			c.mu.Unlock()

			c.usage.Set(usage)
		}
	}()

	return c
}

func (c *CPUCollector) Describe(ch chan<- *prometheus.Desc) { c.usage.Describe(ch) }
func (c *CPUCollector) Collect(ch chan<- prometheus.Metric) { c.usage.Collect(ch) }
