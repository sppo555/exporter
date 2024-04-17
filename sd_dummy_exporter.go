package main

import (
        "encoding/json"
        "flag"
        "fmt"
        "io/ioutil"
        "log"
        "net/http"
        "os"
        "strings"
        "time"

        gce "cloud.google.com/go/compute/metadata"
        "golang.org/x/net/context"
        "golang.org/x/oauth2"
        "golang.org/x/oauth2/google"
        monitoring "google.golang.org/api/monitoring/v3"
)

// func main(){
//      sdExporter()
// }

func sdExporter() {

        if os.Getenv("SD_EXPORTER") == "disable" {
                log.Printf("[%s] [INFO] SD_EXPORTER is set to disable, exiting initialization.", time.Now().Format(time.RFC3339))
                return // 提前返回，不進行後續初始化
        }
        // Gather pod information
        metricsURL := flag.String("metrics-url", "", "URL to fetch the metric name from")
        podId := flag.String("pod-id", "", "pod id")
        namespace := flag.String("namespace", "", "namespace")
        podName := flag.String("pod-name", "", "pod name")
        metricName := flag.String("metric-name", "foo", "custom metric name")
        // metricValue := flag.Int64("metric-value", 0, "custom metric value")
        metricLabelsArg := flag.String("metric-labels", "bar=1", "custom metric labels")
        // Whether to use old Stackdriver resource model - use monitored resource "gke_container"
        // For old resource model, podId flag has to be set.
        useOldResourceModel := flag.Bool("use-old-resource-model", true, "use old stackdriver resource model")
        // Whether to use new Stackdriver resource model - use monitored resource "k8s_pod"
        // For new resource model, podName and namespace flags have to be set.
        useNewResourceModel := flag.Bool("use-new-resource-model", false, "use new stackdriver resource model")
        flag.Parse()

        if *metricsURL == "" {
                log.Fatalf("Metrics URL not specified.")
        }

        if *podId == "" && *useOldResourceModel {
                log.Fatalf("No pod id specified.")
        }

        if *podName == "" && *useNewResourceModel {
                log.Fatalf("No pod name specified.")
        }

        if *namespace == "" && *useNewResourceModel {
                log.Fatalf("No pod namespace specified.")
        }

        stackdriverService, err := getStackDriverService()
        if err != nil {
                log.Fatalf("Error getting Stackdriver service: %v", err)
        }

        metricLabels := make(map[string]string)
        for _, label := range strings.Split(*metricLabelsArg, ",") {
                labelParts := strings.Split(label, "=")
                metricLabels[labelParts[0]] = labelParts[1]
        }

        oldModelLabels := getResourceLabelsForOldModel(*podId)
        newModelLabels := getResourceLabelsForNewModel(*namespace, *podName)
        for {
                metricValue, err := fetchMetricValue(*metricsURL)
                if err != nil {
                        log.Printf("[%s] [ERROR] Error fetching metric value: %v", time.Now().Format(time.RFC3339), err)
                        time.Sleep(5000 * time.Millisecond)
                        continue
                }
                if *useOldResourceModel {
                        err := exportMetric(stackdriverService, *metricName, metricValue, metricLabels, "gke_container", oldModelLabels)
                        if err != nil {
                                log.Printf("[%s] [ERROR] Failed to write time series data for old resource model: %v\n", time.Now().Format(time.RFC3339), err)
                        } else {
                                log.Printf("[%s] [INFO] Successfully wrote time series for old resource model with value %s: %d\n", time.Now().Format(time.RFC3339), *metricName, metricValue)
                        }
                }
                if *useNewResourceModel {
                        err := exportMetric(stackdriverService, *metricName, metricValue, metricLabels, "k8s_pod", newModelLabels)
                        if err != nil {
                                log.Printf("[%s] [ERROR] Failed to write time series data for new resource model: %v\n", time.Now().Format(time.RFC3339), err)
                        } else {
                                log.Printf("[%s] [INFO] Successfully wrote time series for new resource model with value %s: %d\n", time.Now().Format(time.RFC3339), *metricName, metricValue)
                        }
                }
                time.Sleep(30000 * time.Millisecond)
        }
}

func getStackDriverService() (*monitoring.Service, error) {
        oauthClient := oauth2.NewClient(context.Background(), google.ComputeTokenSource(""))
        return monitoring.New(oauthClient)
}

func fetchMetricValue(metricsURL string) (int64, error) {
        resp, err := http.Get(metricsURL)
        if err != nil {
                return 0, err
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
                return 0, err
        }

        var result struct {
                ConnectionNumbers int64 `json:"connectionNumbers"`
        }

        err = json.Unmarshal(body, &result)
        if err != nil {
                return 0, err
        }

        return result.ConnectionNumbers, nil
}

// getResourceLabelsForOldModel returns resource labels needed to correctly label metric data
// exported to StackDriver. Labels contain details on the cluster (project id, name)
// and pod for which the metric is exported (zone, id).
func getResourceLabelsForOldModel(podId string) map[string]string {
        projectId, _ := gce.ProjectID()
        zone, _ := gce.Zone()
        clusterName, _ := gce.InstanceAttributeValue("cluster-name")
        clusterName = strings.TrimSpace(clusterName)
        return map[string]string{
                "project_id":   projectId,
                "zone":         zone,
                "cluster_name": clusterName,
                // container name doesn't matter here, because the metric is exported for
                // the pod, not the container
                "container_name": "",
                "pod_id":         podId,
                // namespace_id and instance_id don't matter
                "namespace_id": "default",
                "instance_id":  "",
        }
}

// getResourceLabelsForNewModel returns resource labels needed to correctly label metric data
// exported to StackDriver. Labels contain details on the cluster (project id, location, name)
// and pod for which the metric is exported (namespace, name).
func getResourceLabelsForNewModel(namespace, name string) map[string]string {
        projectId, _ := gce.ProjectID()
        location, _ := gce.InstanceAttributeValue("cluster-location")
        location = strings.TrimSpace(location)
        clusterName, _ := gce.InstanceAttributeValue("cluster-name")
        clusterName = strings.TrimSpace(clusterName)
        return map[string]string{
                "project_id":     projectId,
                "location":       location,
                "cluster_name":   clusterName,
                "namespace_name": namespace,
                "pod_name":       name,
        }
}

func exportMetric(stackdriverService *monitoring.Service, metricName string,
        metricValue int64, metricLabels map[string]string, monitoredResource string, resourceLabels map[string]string) error {
        dataPoint := &monitoring.Point{
                Interval: &monitoring.TimeInterval{
                        EndTime: time.Now().Format(time.RFC3339),
                },
                Value: &monitoring.TypedValue{
                        Int64Value: &metricValue,
                },
        }
        // Write time series data.
        request := &monitoring.CreateTimeSeriesRequest{
                TimeSeries: []*monitoring.TimeSeries{
                        {
                                Metric: &monitoring.Metric{
                                        Type:   "custom.googleapis.com/" + metricName,
                                        Labels: metricLabels,
                                },
                                Resource: &monitoring.MonitoredResource{
                                        Type:   monitoredResource,
                                        Labels: resourceLabels,
                                },
                                Points: []*monitoring.Point{
                                        dataPoint,
                                },
                        },
                },
        }
        projectName := fmt.Sprintf("projects/%s", resourceLabels["project_id"])
        _, err := stackdriverService.Projects.TimeSeries.Create(projectName, request).Do()
        return err
}