package main

import (
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	connectionNumbers = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "connection_numbers",
		Help: "Number of connections from API",
	})
	alex_test = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "alex_numbers",
		Help: "Alex test metrics",
	})
)

func init() {
	log.SetFlags(0)
	prometheus.MustRegister(connectionNumbers)
	prometheus.MustRegister(alex_test)
}

func main() {
	log.Printf("[%s] [INFO] Starting the application...", time.Now().Format(time.RFC3339))
	go fetchMetrics()
	go fetchJsonUrlMetrics()
	go fetchFileMetrics()
	alex_test.Set(123456)
	http.Handle("/metrics", promhttp.Handler())
	log.Printf("[%s] [INFO] HTTP server started on :8080", time.Now().Format(time.RFC3339))
	http.ListenAndServe(":8080", nil)
}
