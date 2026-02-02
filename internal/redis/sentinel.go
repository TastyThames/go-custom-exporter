package redis

import (
	"context"
	"strings"
	"time"

	rds "github.com/go-redis/redis/v8"
	"github.com/prometheus/client_golang/prometheus"
)

type SentinelMasterInfoCollector struct {
	sentinels  []string
	masterName string
	password   string
	timeout    time.Duration

	masterInfo *prometheus.GaugeVec
	up         prometheus.Gauge

	// per-endpoint
	endpointUp *prometheus.GaugeVec
	reachable  prometheus.Gauge
}

func NewSentinelMasterInfoCollector(sentinelsCSV, masterName, redisPassword string, timeout time.Duration) *SentinelMasterInfoCollector {
	return &SentinelMasterInfoCollector{
		sentinels:  splitCSV(sentinelsCSV),
		masterName: masterName,
		password:   redisPassword,
		timeout:    timeout,

		masterInfo: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lab_redis_sentinel_master_info",
			Help: "Master address discovered via Sentinel. Value is 1 for the current master labels.",
		}, []string{"master_ip", "master_port"}),

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_redis_sentinel_up",
			Help: "1 if exporter can query Sentinel (at least one endpoint reachable), else 0.",
		}),

		endpointUp: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "lab_redis_sentinel_endpoint_up",
			Help: "1 if exporter can query this Sentinel endpoint, else 0.",
		}, []string{"sentinel"}),

		reachable: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "lab_redis_sentinel_reachable_count",
			Help: "Number of Sentinel endpoints reachable by the exporter.",
		}),
	}
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func (c *SentinelMasterInfoCollector) Describe(ch chan<- *prometheus.Desc) {
	c.up.Describe(ch)
	c.masterInfo.Describe(ch)
	c.endpointUp.Describe(ch)
	c.reachable.Describe(ch)
}

func (c *SentinelMasterInfoCollector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
	defer cancel()

	reachableCount := 0

	// Track first master we can discover
	var masterIP, masterPort string
	masterFound := false

	for _, addr := range c.sentinels {
		ok, ip, port := c.queryMasterFromOne(ctx, addr)
		if ok {
			reachableCount++
			c.endpointUp.WithLabelValues(addr).Set(1)

			if !masterFound {
				masterFound = true
				masterIP = ip
				masterPort = port
			}
		} else {
			c.endpointUp.WithLabelValues(addr).Set(0)
		}
	}

	// overall up/down
	if reachableCount > 0 {
		c.up.Set(1)
	} else {
		c.up.Set(0)
	}
	c.reachable.Set(float64(reachableCount))

	// master info + redis ping/role metrics
	if masterFound {
		c.masterInfo.WithLabelValues(masterIP, masterPort).Set(1)

		role, reachable := queryRedisRoleAndPing(masterIP+":"+masterPort, c.password, c.timeout)
		if reachable {
			redisMasterReachable.Set(1)
		} else {
			redisMasterReachable.Set(0)
		}

		if role == "" {
			role = "unknown"
		}
		redisLocalRole.WithLabelValues(role).Set(1)

		expected := "master"
		if role != expected {
			redisRoleMismatch.WithLabelValues(expected, role).Set(1)
		} else {
			redisRoleMismatch.WithLabelValues(expected, role).Set(0)
		}
	} else {
		redisMasterReachable.Set(0)
	}

	// IMPORTANT: emit all metrics
	c.up.Collect(ch)
	c.masterInfo.Collect(ch)
	c.endpointUp.Collect(ch)
	c.reachable.Collect(ch)
}

func (c *SentinelMasterInfoCollector) queryMasterFromOne(ctx context.Context, addr string) (ok bool, ip string, port string) {
	rdb := rds.NewClient(&rds.Options{Addr: addr})
	res, err := rdb.Do(ctx, "SENTINEL", "get-master-addr-by-name", c.masterName).Result()
	_ = rdb.Close()

	if err != nil {
		return false, "", ""
	}

	arr, ok2 := res.([]interface{})
	if !ok2 || len(arr) < 2 {
		return false, "", ""
	}

	ip, _ = arr[0].(string)
	port, _ = arr[1].(string)

	if ip == "" || port == "" {
		return false, "", ""
	}
	return true, ip, port
}

// We export role / mismatch / reachable here to keep Phase 2 minimal but useful.
var (
	redisLocalRole = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lab_redis_local_role",
		Help: "Role of the current Redis master as reported by INFO replication (exported as a label).",
	}, []string{"role"})

	redisRoleMismatch = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "lab_redis_role_mismatch",
		Help: "1 if the Redis role is not as expected.",
	}, []string{"expected", "actual"})

	redisMasterReachable = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "lab_redis_master_reachable",
		Help: "1 if exporter can PING the current Redis master, else 0.",
	})
)

func init() {
	prometheus.MustRegister(redisLocalRole, redisRoleMismatch, redisMasterReachable)
}
