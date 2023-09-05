package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var urlJsonMetrics = make(map[string]prometheus.Gauge)
var RESOURCE_JSON string

func init() {
	log.SetFlags(0)
	// export RESOURCE_JSON=http://35.185.155.194/api/metrics/connection-numbers
	RESOURCE_JSON = os.Getenv("RESOURCE_JSON")
	if RESOURCE_JSON == "" {
		log.Fatalf("[%s] [ERROR] RESOURCE_JSON environment variable is not set", time.Now().Format(time.RFC3339))
	}
}

func fetchJsonUrlMetrics() {
	for {
		resp, err := http.Get(RESOURCE_JSON)
		if err != nil {
			log.Printf("[%s] [ERROR] Failed to fetch JSON data: %s", time.Now().Format(time.RFC3339), err)
			time.Sleep(10 * time.Second)
			continue
		}

		var data map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&data)
		resp.Body.Close()
		if err != nil {
			log.Printf("[%s] [ERROR] Failed to decode JSON data: %s", time.Now().Format(time.RFC3339), err)
			time.Sleep(10 * time.Second)
			continue
		}

		processJsonData("", data)
		log.Printf("[%s] [INFO] Json URL Refresh content", time.Now().Format(time.RFC3339))
		time.Sleep(10 * time.Second)
	}
}

func processJsonData(prefix string, data map[string]interface{}) {
	for key, value := range data {
		newPrefix := key
		if prefix != "" {
			newPrefix = prefix + "_" + key
		}

		switch v := value.(type) {
		case float64, string:
			setUrlJsonMetricValue(newPrefix, v)
		case map[string]interface{}:
			processJsonData(newPrefix, v)
		default:
			// Unsupported type, you can handle or log it here if needed
		}
	}
}

func setUrlJsonMetricValue(key string, value interface{}) {
	if _, exists := urlJsonMetrics[key]; !exists {
		urlJsonMetrics[key] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: key,
			Help: "Metric from JSON URL",
		})
		prometheus.MustRegister(urlJsonMetrics[key])
	}

	switch v := value.(type) {
	case float64:
		urlJsonMetrics[key].Set(v)
	case string:
		if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
			urlJsonMetrics[key].Set(floatValue)
		} else {
			urlJsonMetrics[key].Set(1)
		}
	default:
	}
}
