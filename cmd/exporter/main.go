package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"go-custom-exporter/internal/app"
	"go-custom-exporter/internal/proc"
	"go-custom-exporter/internal/redis"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	listenAddr := getenv("EXPORTER_LISTEN", ":9200")

	sentinelAddrs := getenv("REDIS_SENTINELS", "")
	masterName := getenv("REDIS_MASTER_NAME", "mymaster")
	redisPassword := getenv("REDIS_PASSWORD", "") // optional

	log.Printf("config: listen=%q sentinels=%q masterName=%q\n", listenAddr, sentinelAddrs, masterName)

	// OS metrics
	prometheus.MustRegister(proc.NewCPUCollector(1 * time.Second))
	prometheus.MustRegister(proc.NewMemCollector())
	prometheus.MustRegister(proc.NewLoadCollector())
	prometheus.MustRegister(proc.NewUptimeCollector())
	prometheus.MustRegister(proc.NewNetCollector())

	// Sentinel-aware Redis metrics
	if sentinelAddrs == "" {
		log.Println("WARN: REDIS_SENTINELS is empty; sentinel metrics disabled")
	} else {
		prometheus.MustRegister(
			redis.NewSentinelMasterInfoCollector(sentinelAddrs, masterName, redisPassword, 5*time.Second),
		)
		log.Println("sentinel collector registered")
	}

	// App metrics (stub for now)
	appMetrics := app.NewAppMetrics()
	appMetrics.Register()

	// HTTP
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	log.Printf("go-custom-exporter listening on %s\n", listenAddr)
	log.Fatal(http.ListenAndServe(listenAddr, mux))
}
