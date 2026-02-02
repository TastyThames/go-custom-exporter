package proc

import (
	"bufio"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

type NetCollector struct {
	rx *prometheus.GaugeVec
	tx *prometheus.GaugeVec
}

func NewNetCollector() *NetCollector {
	return &NetCollector{
		rx: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lab_net_rx_bytes_total",
			Help: "RX bytes by interface from /proc/net/dev (counter value as gauge).",
		}, []string{"iface"}),
		tx: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lab_net_tx_bytes_total",
			Help: "TX bytes by interface from /proc/net/dev (counter value as gauge).",
		}, []string{"iface"}),
	}
}

func (n *NetCollector) Describe(ch chan<- *prometheus.Desc) {
	n.rx.Describe(ch)
	n.tx.Describe(ch)
}

func (n *NetCollector) Collect(ch chan<- prometheus.Metric) {
	f, err := os.Open("/proc/net/dev")
	if err == nil {
		defer f.Close()
		sc := bufio.NewScanner(f)
		for sc.Scan() {
			line := strings.TrimSpace(sc.Text())
			if !strings.Contains(line, ":") || strings.HasPrefix(line, "Inter-") || strings.HasPrefix(line, "face") {
				continue
			}

			parts := strings.Fields(strings.ReplaceAll(line, ":", " "))
			if len(parts) < 10 {
				continue
			}

			iface := parts[0]
			rxBytes, _ := strconv.ParseFloat(parts[1], 64)
			txBytes, _ := strconv.ParseFloat(parts[9], 64)

			n.rx.WithLabelValues(iface).Set(rxBytes)
			n.tx.WithLabelValues(iface).Set(txBytes)
		}
	}

	n.rx.Collect(ch)
	n.tx.Collect(ch)
}
