# Prometheus Custom Exporter

## 此功能主要使用在rocketws監控,亦可在其他相同需求上使用
## 主要功能

1. 從不同的資料源（如 API、Local文件和 JSON URL）收集指標。
2. 針對rocketws的connectionNumbers抓取數值傳送到GCP Monitor
3. 提供 `/metrics` 端點，以供 Prometheus 抓取。

## 文件結構

- `main.go`: 主應用程式入口，初始化 Prometheus 指標和 HTTP 伺服器。
- `fetchMetrics.go`: 專門為Rocketws 的URL /api/metrics/connection-numbers 收集指標。
- `fileMetrics.go`: 從Local文件讀取和解析指標。
- `JsonUrlMetrics.go`: 從 HTTP URL  API JSON格式讀取和解析指標。
- `sd_dummy_exporter.go`: 只針對metrics-url抓取數值傳送到GCP Monitor。
- `merge`: 抓取mq server比對local client跟 server兩邊的消費狀況,並產出/tmp/mq_consumer.log

## 環境變數

1. 開關環境變數 METRICS_JSON, METRICS_FILE, METRICS_URL, SD_EXPORTER, MQ_CONSUMER
   設定成disable則關閉功能。
2. `RESOURCE_URL`: Rocketws URL，用於 `fetchMetrics.go`。
3. `RESOURCE_FILE`: 指標文件的路徑，用於 `fileMetrics.go`。
4. `RESOURCE_JSON`: URL 的路徑，用於 `JsonUrlMetrics.go`。
5. `MQ_NAMESERVER`: 用於merge.go

## sd_dummy_exporter 使用須知
### GCE VM
可直接開啟 Stackdriver Monitoring API 讀寫
### k8s
需安裝 

https://github.com/GoogleCloudPlatform/k8s-stackdriver/tree/master/custom-metrics-stackdriver-adapter

安裝的custom-metrics-stackdriver-adapter請按照說明授予workload identity的Monitoring API 讀寫
#### pod 
使用的deployment or other必須有授權gcp workload identity的Monitoring API 讀寫

## 如何運行

1. 設置上述環境變數。
2. 運行應用程式：
   ```bash
   export RESOURCE_URL=XXX RESOURCE_FILE=/XXX/XXX
   go run *.go  --pod-id=123 --namespace=ws --pod-name=ws --metrics-url=http://1.1.1.1/api/metrics/connection-numbers
   ```
3. binary run
   ```sh
   CGO_ENABLED=0 GOOS=linux go install sd-dummy-exporter
   ./sd-dummy-exporter --pod-id=123 --namespace=ws --pod-name=ws --metrics-url=http://_URL_/api/metrics/connection-numbers
   ```
4. Docker-Compose:
   ```sh
   docker build -t custom-exporter:latest .
   ```
   ```bash
   version: '3'
   services:
     custom-exporter:
       image: custom-exporter:latest
       container_name: custom-exporter
       restart: unless-stopped
       volumes:
         - /tmp/test.txt:/app/test/txt
       environment:
        # - METRICS_JSON=disable
        # - METRICS_FILE=disable
        # - METRICS_URL=disable
        # - SD_EXPORTER=disable
        # - MQ_CONSUMER=disable
         - RESOURCE_URL=http://_URL_/api/metrics/connection-numbers
         - RESOURCE_FILE=/app/test.txt
         - RESOURCE_JSON=http://_URL_/url_api_json
         - MQ_NAMESERVER=localhost:8888
       command: ["./main", "--pod-id=test", "--namespace=ws", "--pod-name=ws", "--metrics-url=http://localhost/api/metrics/connection-numbers"]
       ports:
         - 8080:8080
   ```

#### 開啟http://localhost:8080/metrics 就可以看到prometheus的metrics監控

### fileMetrics 的File 範例
```sh
aaa:123
{"power":"999", "house":true, "car":null}
abc:abc
#.
#...等
# 以上就會有5個metrics 與 5 個 value 
# 可以無限多行
```