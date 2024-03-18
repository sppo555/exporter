package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	connectionNumbers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connection_numbers",
		Help: "Number of connections from API",
	})
	// alex_test = prometheus.NewGauge(prometheus.GaugeOpts{
	// 	Name: "alex_numbers",
	// 	Help: "Alex test metrics",
	// })
)

func init() {
	log.SetFlags(0)
	prometheus.MustRegister(connectionNumbers)
	// prometheus.MustRegister(alex_test)
}

func main() {
	log.Printf("[%s] [INFO] Starting the application...", time.Now().Format(time.RFC3339))
	METRICS_JSON := os.Getenv("METRICS_JSON")
	METRICS_URL := os.Getenv("METRICS_URL")
	METRICS_FILE := os.Getenv("METRICS_FILE")
	if METRICS_JSON != "disable" {
		go fetchJsonUrlMetrics()
	}

	if METRICS_URL != "disable" {
		go fetchMetrics()
	}

	if METRICS_FILE != "disable" {
		go fetchFileMetrics()
	}
	// alex_test.Set(123456)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("[%s] [INFO] HTTP server started on :8080", time.Now().Format(time.RFC3339))
	http.ListenAndServe(":8080", nil)
}
