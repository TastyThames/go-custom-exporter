package main

import (
	"fmt"
	"log"
	"net/http"
)

const (
	listenAddr = ":9200"
	version    = "dev"
)

func metricsHandler(w http.ResponseWriter, r *http.Request) {
	// Prometheus text exposition format
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	// Exporter health metric
	fmt.Fprintln(w, "# HELP os_exporter_up 1 if the exporter is running")
	fmt.Fprintln(w, "# TYPE os_exporter_up gauge")
	fmt.Fprintln(w, "os_exporter_up 1")

	// Build information metric (useful for debugging deployments)
	fmt.Fprintln(w, "# HELP os_exporter_build_info Build information")
	fmt.Fprintln(w, "# TYPE os_exporter_build_info gauge")
	fmt.Fprintf(w, "os_exporter_build_info{version=%q} 1\n", version)
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func main() {
	log.Printf("os-exporter starting on %s\n", listenAddr)

	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/health", healthHandler)

	// ListenAndServe blocks forever (good for long-running services like exporters)
	log.Fatal(http.ListenAndServe(listenAddr, nil))
}
