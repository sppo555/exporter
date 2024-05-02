package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type OffsetData struct {
	BrokerName string
	Topic      string
	QueueId    int
	Cnt        int
}

type SourceContent struct {
	Status int          `json:"status"`
	Data   []SourceData `json:"data"`
	ErrMsg string       `json:"errMsg"`
}

type SourceData struct {
	Topic             string          `json:"topic"`
	QueueStatInfoList []QueueStatInfo `json:"queueStatInfoList"`
}

type QueueStatInfo struct {
	BrokerOffset int    `json:"brokerOffset"`
	ClientInfo   string `json:"clientInfo"`
}

// func main() {
//     MqConsumer()
// }

func MqConsumer() {
	log.SetFlags(0)
	for {
		time.Sleep(30 * time.Second)
		if os.Getenv("MQ_CONSUMER") == "disable" {
			log.Println("[%s] [INFO] MQ_CONSUMER is set to disable, exiting initialization. %s\n", time.Now().Format(time.RFC3339))
			return // 提前返回，不進行後續的初始化操作
		}
		basePath := "/root/.rocketmq_offsets/"

		csOffsets := []string{"ws-msg-consumer", "room-consumer"}
		result := make(map[string]int)

		// Find the latest client_id folder
		clientID, err := findLatestClientID(basePath)
		if err != nil {
			log.Fatalf("[%s] [ERROR] Error finding latest client_id folder: %v", time.Now().Format(time.RFC3339), err)
			time.Sleep(30 * time.Second) // 休眠一段時間後再重新執行
			continue
		}

		for _, csOffset := range csOffsets {
			// Find matching group_name folder
			groupName, err := findGroupName(csOffset, clientID, basePath)
			if err != nil {
				log.Fatalf("[%s] [ERROR] Error finding matching group_name folder: %s\n", time.Now().Format(time.RFC3339), err)
				continue
			}

			// Compose target path
			targetPath := filepath.Join(basePath, clientID, groupName, "offsets.json")

			// Read offsets.json
			fileBytes, err := ioutil.ReadFile(targetPath)
			if err != nil {
				log.Fatalf("[%s] [ERROR] Error reading offsets.json: %s\n", time.Now().Format(time.RFC3339), err)
				continue
			}

			// Use regex to match data
			re := regexp.MustCompile(`\{\s*"brokerName":"(.*?)",\s*"queueId":(\d+),\s*"topic":"(.*?)"\s*\}:(\d+)`)
			matches := re.FindAllStringSubmatch(string(fileBytes), -1)

			var totalCntValue int
			for _, match := range matches {
				queueCnt, err := strconv.Atoi(match[4])
				if err != nil {
					log.Fatalf("[%s] [ERROR] Error converting Cnt to int: %s\n", time.Now().Format(time.RFC3339), err)
					return
				}
				totalCntValue += queueCnt
			}

			key := ""
			if csOffset == "ws-msg-consumer" {
				key = "client-ws-msg-consumer"
			} else if csOffset == "room-consumer" {
				key = "client-room-consumer"
			}
			result[key] = totalCntValue
		}

		// Combine with server.go results
		serverResult := make(map[string]int)

		for _, csOffset := range csOffsets {
			groupName, err := findGroupName(csOffset, clientID, basePath)
			if err != nil {
				log.Fatalf("[%s] [ERROR] Error:", time.Now().Format(time.RFC3339), err)
				continue
			}

			sourceContentURL := getSourceContentURL(groupName)
			var sourceContent SourceContent
			var totalBrokerOffset int

			for totalBrokerOffset == 0 {
				resp, err := http.Get(sourceContentURL)
				if err != nil {
					log.Fatalf("[%s] [ERROR] Error fetching source content: %s\n", time.Now().Format(time.RFC3339), err)
					return
				}
				defer resp.Body.Close()

				err = json.NewDecoder(resp.Body).Decode(&sourceContent)
				if err != nil {
					log.Fatalf("[%s] [ERROR] Error decoding source content: %s\n", time.Now().Format(time.RFC3339), err)
					return
				}

				totalBrokerOffset = processSourceContent(csOffset, clientID, groupName, sourceContent)

				time.Sleep(1 * time.Second) // Sleep for 1 second before retrying
			}

			serverResult[csOffset] = totalBrokerOffset
		}

		roomTotal := serverResult["room-consumer"] - result["client-room-consumer"]
		msgTotal := serverResult["ws-msg-consumer"] - result["client-ws-msg-consumer"]

		finalResult := map[string]int{
			"room_total":             roomTotal,
			"msg_total":              msgTotal,
			"client_room_consumer":   result["client-room-consumer"],
			"client_ws_msg_consumer": result["client-ws-msg-consumer"],
			"server_room_consumer":   serverResult["room-consumer"],
			"server_ws_msg_consumer": serverResult["ws-msg-consumer"],
		}

		finalJSON, err := json.Marshal(finalResult)
		if err != nil {
			log.Fatalf("[%s] [ERROR] Error marshalling final result to JSON: %s\n", time.Now().Format(time.RFC3339), err)
			return
		}

		// 定义一个新的正则表达式来匹配 JSON 键，并且仅替换键中的"-"
		reKey := regexp.MustCompile(`"(\w+)-(\w+)":`)
		modifiedJSON := reKey.ReplaceAllString(string(finalJSON), `"${1}_${2}":`)

		logFilePath := "/tmp/mq_consumer.log"
		err = ioutil.WriteFile(logFilePath, []byte(modifiedJSON), 0644)
		if err != nil {
			log.Fatalf("[%s] [ERROR] Error writing result to log file: %s\n", time.Now().Format(time.RFC3339), err)
			return
		}

		log.Printf("[%s] [INFO] Execution completed. Result has been written to %s\n", time.Now().Format(time.RFC3339), logFilePath)
	}
}

func findLatestClientID(basePath string) (string, error) {
	infos, err := ioutil.ReadDir(basePath)
	if err != nil {
		return "", err
	}

	var latestTime time.Time
	var latestClientID string

	for _, info := range infos {
		if info.IsDir() && info.ModTime().After(latestTime) {
			latestTime = info.ModTime()
			latestClientID = info.Name()
		}
	}

	if latestClientID == "" {
		return "", fmt.Errorf("no client_id folder found")
	}

	return latestClientID, nil
}

func findGroupName(csOffset, clientID, basePath string) (string, error) {
	clientDir := filepath.Join(basePath, clientID)
	groupNames, err := ioutil.ReadDir(clientDir)
	if err != nil {
		return "", err
	}

	var matchingGroupName string
	for _, groupName := range groupNames {
		if groupName.IsDir() && strings.Contains(groupName.Name(), csOffset) {
			matchingGroupName = groupName.Name()
			break
		}
	}

	if matchingGroupName == "" {
		return "", fmt.Errorf("[%s] [ERROR] no matching group_name folder found for CSOffset: %s", csOffset)
	}

	return matchingGroupName, nil
}

func getSourceContentURL(groupName string) string {
	mqNameserver := os.Getenv("MQ_NAMESERVER")
	return fmt.Sprintf("http://%s/consumer/queryTopicByConsumer.query?consumerGroup=%s", mqNameserver, groupName)
}

func processSourceContent(csOffset, clientID, groupName string, sourceContent SourceContent) int {
	var totalBrokerOffset int

	for _, data := range sourceContent.Data {
		for _, queueStat := range data.QueueStatInfoList {
			if csOffset == "ws-msg-consumer" {
				if strings.Contains(queueStat.ClientInfo, clientID) {
					totalBrokerOffset += queueStat.BrokerOffset
				}
			} else if csOffset == "room-consumer" {
				totalBrokerOffset += queueStat.BrokerOffset
			}
		}
	}

	return totalBrokerOffset
}
