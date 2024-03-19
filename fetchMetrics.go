package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

var resourceURL string

type apiResponse struct {
	ConnectionNumbers float64 `json:"connectionNumbers"`
}

func init() {
	log.SetFlags(0)

	if os.Getenv("METRICS_URL") == "disable" {
		log.Printf("[%s] [INFO] METRICS_URL is set to disable, exiting initialization.", time.Now().Format(time.RFC3339))
		resourceURL = "" // 將 resourceURL 設定為空字符串，後續邏輯中將跳過執行
		return           // 提前返回，不進行後續的初始化操作
	}

	resourceURL = os.Getenv("RESOURCE_URL")
	if resourceURL == "false" || resourceURL == "" {
		log.Printf("[%s] [WARNING] RESOURCE_URL environment variable is not set or invalid, fetchMetrics will not run", time.Now().Format(time.RFC3339))
		resourceURL = "" // 设置 resourceURL 为空字符串，后续逻辑中将跳过执行
	}
}

// isValidURL 检查给定的字符串是否为有效的 URL
func isValidURL(toTest string) bool {
	u, err := url.Parse(toTest)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func fetchMetrics() {
	// 如果 resourceURL 为空，直接返回而不执行任何操作
	if resourceURL == "" {
		return
	}

	// 检查 RESOURCE_URL 是否为有效的 URL
	if !isValidURL(resourceURL) {
		log.Printf("[%s] [ERROR] Invalid RESOURCE_URL: %s", time.Now().Format(time.RFC3339), resourceURL)
		return
	}

	for {
		resp, err := http.Get(resourceURL)
		if err != nil {
			log.Printf("[%s] [ERROR] Failed to fetch metrics: %s", time.Now().Format(time.RFC3339), err)
			time.Sleep(10 * time.Second)
			continue // 出错后等待一段时间后重试
		}

		var result apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			log.Printf("[%s] [ERROR] Failed to decode metrics: %s", time.Now().Format(time.RFC3339), err)
			resp.Body.Close()
			time.Sleep(10 * time.Second)
			continue
		}
		resp.Body.Close()

		connectionNumbers.Set(result.ConnectionNumbers)
		log.Printf("[%s] [INFO] ConnectionNumbers updated: %f", time.Now().Format(time.RFC3339), result.ConnectionNumbers)
		time.Sleep(10 * time.Second) // 等待一段时间后再次执行
	}
}
