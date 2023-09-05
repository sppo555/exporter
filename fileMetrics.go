package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var METRICS_FILE string
var fileMetrics = make(map[string]prometheus.Gauge)

type jsonMetrics struct {
	Data map[string]interface{} `json:"-"`
}

func init() {
	log.SetFlags(0)
	// export RESOURCE_FILE=/tmp/test.txt
	METRICS_FILE = os.Getenv("RESOURCE_FILE")
	if METRICS_FILE == "" {
		// Handle the error if the environment variable is not set
		log.Fatalf("[%s] [ERROR] RESOURCE_URL environment variable is not set", time.Now().Format(time.RFC3339))
	}
}

func (jm *jsonMetrics) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &jm.Data)
}

func fetchFileMetrics() {
	for {
		activeMetrics := make(map[string]bool)

		content, err := ioutil.ReadFile(METRICS_FILE)
		if err != nil {
			log.Fatalf("[%s] [ERROR] Failed to fetch metrics: %s", time.Now().Format(time.RFC3339), err)
		}

		lines := strings.Split(string(content), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			if strings.Contains(line, "{") && strings.Contains(line, "}") {
				var jm jsonMetrics
				err := json.Unmarshal([]byte(line), &jm)
				if err != nil {
					panic(err)
				}

				for key, value := range jm.Data {
					setMetricValue(key, value)
					activeMetrics[key] = true
				}
			} else {
				parts := strings.Split(line, ":")
				if len(parts) == 2 {
					setMetricValue(parts[0], parts[1])
					activeMetrics[parts[0]] = true
				}
			}
		}

		// Remove inactive metrics
		for key := range fileMetrics {
			if _, isActive := activeMetrics[key]; !isActive {
				prometheus.Unregister(fileMetrics[key])
				delete(fileMetrics, key)
			}
		}

		log.Printf("[%s] [INFO] FILE Refresh content", time.Now().Format(time.RFC3339))
		time.Sleep(10 * time.Second)
	}
}

func setMetricValue(key string, value interface{}) {
	if _, exists := fileMetrics[key]; !exists {
		fileMetrics[key] = prometheus.NewGauge(prometheus.GaugeOpts{
			Name: key,
			Help: "Metric from file",
		})
		prometheus.MustRegister(fileMetrics[key])
	}

	switch v := value.(type) {
	case float64:
		fileMetrics[key].Set(v)
	case string:
		if floatValue, err := strconv.ParseFloat(v, 64); err == nil {
			fileMetrics[key].Set(floatValue)
		} else {
			fileMetrics[key].Set(1)
		}
	default:
	}
}
