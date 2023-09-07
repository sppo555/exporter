# Prometheus Custom Exporter


## 主要功能

1. 從不同的資料源（如 API、文件和 JSON URL）收集指標。
2. 提供 `/metrics` 端點，以供 Prometheus 抓取。

## 文件結構

- `main.go`: 主應用程式入口，初始化 Prometheus 指標和 HTTP 伺服器。
- `fetchMetrics.go`: 從 API 端點收集指標。
- `fileMetrics.go`: 從文件讀取和解析指標。
- `JsonUrlMetrics.go`: 從 JSON URL 讀取和解析指標。

## 環境變數

1. `RESOURCE_URL`: API 的 URL，用於 `fetchMetrics.go`。
2. `RESOURCE_FILE`: 指標文件的路徑，用於 `fileMetrics.go`。
3. `RESOURCE_JSON`: JSON URL 的路徑，用於 `JsonUrlMetrics.go`。

## 如何運行

1. 設置上述環境變數。
2. 運行應用程式：
   ```bash
   export RESOURCE_URL=XXX RESOURCE_FILE=/XXX/XXX
   go run *.go
   ```
3. Docker-Compose:
   ```bash
   version: '3.8'
   services:
     custom-exporter:
       image: sppo555/exporter:latest
       container_name: custom-exporter
       restart: unless-stopped
       volumes:
         - /tmp/test.txt:/app/test/txt
       environment:
         - RESOURCE_URL=http://_IP_/api/metrics/connection-numbers
         - RESOURCE_FILE=/app/test.txt
         - RESOURCE_JSON=http://_IP_/api/json
       ports:
         - 8080:8080
    ```
