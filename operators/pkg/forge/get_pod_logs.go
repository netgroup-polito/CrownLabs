package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func getAllNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]string, error) {
	namespaces, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	var nsNames []string
	for _, ns := range namespaces.Items {
		nsNames = append(nsNames, ns.Name)
	}
	return nsNames, nil
}

// LogEntry is a struct to hold a log entry and its timestamp.
type LogEntry struct {
	Timestamp float64 `json:"ts"`
	Entry     string  `json:"entry"`
}

// GetPodLogsForResource returns logs for a given resource from the CAPI controllers containing the provider name label.
func GetPodLogsForResource(ctx context.Context, c client.Client, restConfig *rest.Config, kind string, namespace string, name string) ([]string, error) {
	log := ctrl.LoggerFrom(ctx)
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	namespaces, err := getAllNamespaces(ctx, clientset)
	if err != nil {
		return nil, err
	}

	logEntries := make([]LogEntry, 0)

	selector := metav1.LabelSelector{
		MatchExpressions: []metav1.LabelSelectorRequirement{
			{
				Key:      clusterv1.ProviderNameLabel,
				Operator: metav1.LabelSelectorOpExists,
			},
		},
	}
	labelselector, err := metav1.LabelSelectorAsSelector(&selector)
	if err != nil {
		return nil, err
	}

	for _, podNamespace := range namespaces {
		listOpts := []client.ListOption{
			client.MatchingLabelsSelector{Selector: labelselector},
			client.InNamespace(podNamespace),
		}

		podList := &corev1.PodList{}
		if err := c.List(ctx, podList, listOpts...); err != nil {
			return nil, err
		}

		for _, pod := range podList.Items {
			logs, err := getPodLogs(ctx, clientset, pod)
			if err != nil {
				return nil, err
			}
			logs = strings.TrimSuffix(logs, "\n")
			res := strings.Split(logs, "\n")
			for _, line := range res {
				timeStampString := "{\"ts\":"
				kindNamespaceNameString := fmt.Sprintf("%s=\"%s/%s\"", kind, namespace, name)
				controllerKindString := fmt.Sprintf("controllerKind=\"%s\"", kind)
				namespaceString := fmt.Sprintf("namespace=\"%s\"", namespace)
				nameString := fmt.Sprintf("name=\"%s\"", name)
				if strings.Contains(line, timeStampString) { // Try to parse string as a JSON object with a "ts" key.
					jsonMap := make(map[string]interface{})
					if err := json.Unmarshal([]byte(line), &jsonMap); err != nil {
						log.Error(err, "Failed to parse log line", "line", line)
						continue
					} else {
						ts, ok := jsonMap["ts"].(float64)
						if !ok {
							log.Error(errors.Errorf("Failed to parse timestamp"), "line", line)
							continue
						}
						if logMatchesResource(jsonMap, kind, namespace, name) {
							logEntries = append(logEntries, LogEntry{Timestamp: ts, Entry: line})
						}
					}
				} else if strings.Contains(line, kindNamespaceNameString) ||
					(strings.Contains(line, controllerKindString) && strings.Contains(line, namespaceString) && strings.Contains(line, nameString)) { // Otherwise try to parse as non-JSON log line.
					// We expect the format to look like this:
					// I1126 21:28:39.622231       1 some_controller.go:12] "some message" controllerKind="TestKind" TestKind="test-namespace/test-name" namespace="test-namespace" name="test-name"
					re := regexp.MustCompile(`\d\d:\d\d:\d\d.\d*`)
					timestamp := re.Find([]byte(line))
					tm, err := time.Parse("15:04:05.000000", string(timestamp))
					if err != nil {
						log.Error(err, "Failed to parse timestamp", "line", line)
						continue
					}
					ts := tm.UnixMicro()
					logEntries = append(logEntries, LogEntry{Timestamp: float64(ts), Entry: line})
				}
			}
		}
	}

	// Note: we could probably optimize this using a merge-sort strategy since each of the log entries for each pod are already sorted.
	sort.Slice(logEntries, func(i, j int) bool {
		// Return with most recent at top
		return logEntries[i].Timestamp > logEntries[j].Timestamp
	})

	allLogs := make([]string, len(logEntries))
	for i, entry := range logEntries {
		allLogs[i] = entry.Entry
	}

	return allLogs, nil
}

// logMatchesResource returns true if the log entry matches the given resource by having a keys in the format of kind.namespace and kind.name.
// For example, if there is a Cluster with "name: test-cluster" and "namespace: test-namespace", then the log entry should have the following keys:
// "Cluster": {"name": "test-cluster", "namespace": "test-namespace"}
func logMatchesResource(jsonMap map[string]interface{}, kind string, namespace string, name string) bool {
	if val, ok := GetNestedValue(jsonMap, []string{kind, "name"}); ok && val == name {
		if val, ok := GetNestedValue(jsonMap, []string{kind, "namespace"}); ok && val == namespace {
			return true
		}
	}
	return false
}

// GetNestedValue returns the value of a nested key in a map.
func GetNestedValue(data map[string]any, path []string) (result string, found bool) {
	if len(path) == 0 {
		return "", false
	}
	if len(path) == 1 {
		if val, ok := data[path[0]]; ok {
			if s, ok := val.(string); ok {
				return s, true
			}
		}
		return "", false
	}

	if val, ok := data[path[0]]; ok {
		if m, ok := val.(map[string]any); ok {
			return GetNestedValue(m, path[1:])
		}
	}

	return "", false
}

// getPodLogs returns logs for a given pod.
func getPodLogs(ctx context.Context, clientset *kubernetes.Clientset, pod corev1.Pod) (string, error) {
	allLogs := ""
	for _, c := range pod.Spec.InitContainers {
		if logs, err := getPodContainerLogs(ctx, clientset, pod, c.Name); err != nil {
			return "", err
		} else {
			allLogs = allLogs + logs
		}
	}
	for _, c := range pod.Spec.Containers {
		if logs, err := getPodContainerLogs(ctx, clientset, pod, c.Name); err != nil {
			return "", err
		} else {
			allLogs = allLogs + logs
		}
	}
	for _, c := range pod.Spec.EphemeralContainers {
		if logs, err := getPodContainerLogs(ctx, clientset, pod, c.Name); err != nil {
			return "", err
		} else {
			allLogs = allLogs + logs
		}
	}

	return allLogs, nil
}

// getPodContainerLogs returns logs for a given pod and container.
func getPodContainerLogs(ctx context.Context, clientset *kubernetes.Clientset, pod corev1.Pod, containerName string) (string, error) {
	podLogOpts := corev1.PodLogOptions{
		Container: containerName,
	}

	req := clientset.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &podLogOpts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer podLogs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, podLogs)
	if err != nil {
		return "", err
	}
	str := buf.String()

	return str, nil
}
