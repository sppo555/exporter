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
)

func init() {
    log.SetFlags(0)
    prometheus.MustRegister(connectionNumbers)
}

func main() {
    log.SetFlags(0)
    log.Printf("[%s] [INFO] Starting the application...", time.Now().Format(time.RFC3339))
    METRICS_JSON := os.Getenv("METRICS_JSON")
    METRICS_URL := os.Getenv("METRICS_URL")
    METRICS_FILE := os.Getenv("METRICS_FILE")
    SD_EXPORTER := os.Getenv("SD_EXPORTER")
    MQ_CONSUMER := os.Getenv("MQ_CONSUMER")
    if METRICS_JSON != "disable" {
        go fetchJsonUrlMetrics()
    }

    if METRICS_URL != "disable" {
        go fetchMetrics()
    }

    if METRICS_FILE != "disable" {
        go fetchFileMetrics()
    }

    if SD_EXPORTER != "disable" {
        go sdExporter()
    }

    if MQ_CONSUMER != "disable" {
        go MqConsumer()
    }

    port := ":8080"
    if portEnv := os.Getenv("SERVER_PORT"); portEnv != "" {
        port = ":" + portEnv
    }

    http.Handle("/metrics", promhttp.Handler())
    log.Printf("[%s] [INFO] HTTP server started on %s", time.Now().Format(time.RFC3339), port)
    http.ListenAndServe(port, nil)
}