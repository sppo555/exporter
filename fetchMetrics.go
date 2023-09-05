package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

var resourceURL string

type apiResponse struct {
	ConnectionNumbers float64 `json:"connectionNumbers"`
}

func init() {
	log.SetFlags(0)
	// export RESOURCE_URL=http://35.185.155.194/api/metrics/connection-numbers
	resourceURL = os.Getenv("RESOURCE_URL")
	if resourceURL == "" {
		// Provide a default value or handle the error if the environment variable is not set
		log.Fatalf("[%s] [ERROR] RESOURCE_URL environment variable is not set", time.Now().Format(time.RFC3339))
	}
}

func fetchMetrics() {
	for {
		resp, err := http.Get(resourceURL)
		if err != nil {
			log.Fatalf("[%s] [ERROR] Failed to fetch metrics: %s", time.Now().Format(time.RFC3339), err)
		}

		var result apiResponse
		json.NewDecoder(resp.Body).Decode(&result)
		resp.Body.Close()

		connectionNumbers.Set(result.ConnectionNumbers)
		log.Printf("[%s] [INFO] ConnectionNumbers Refresh content", time.Now().Format(time.RFC3339))
		time.Sleep(10 * time.Second)
	}
}
