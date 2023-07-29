package k8s

import (
	"encoding/json"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func getClientSet() *kubernetes.Clientset {
	restConfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&clientcmd.ConfigOverrides{ClusterInfo: clientcmdapi.Cluster{Server: ""}}).ClientConfig()
	if err != nil {
		return nil
	}
	return kubernetes.NewForConfigOrDie(restConfig)
}

func getLabels(labelString string) map[string]string {
	labelMap := make(map[string]string)
	labelPairs := strings.Split(labelString, ";")

	for _, labelPair := range labelPairs {
		labelPair = strings.TrimSpace(labelPair)
		parts := strings.Split(labelPair, "=")
		if len(parts) != 2 {
			return nil
		}
		labelMap[parts[0]] = parts[1]
	}
	return labelMap
}

func parseAffinityRules(rulesString string) (map[string][]string, map[string][]string, []string, []string) {
	inOps := make(map[string][]string)
	notInOps := make(map[string][]string)
	var exists []string
	var doesNotExists []string

	parts := strings.Split(rulesString, ";")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if strings.Contains(part, "!=") {
			key, value := extractKeyValue(part, "!=")
			if value == "" {
				doesNotExists = append(doesNotExists, key)
			} else {
				fillMap(&notInOps, key, value)
			}
		} else if strings.Contains(part, "=") {
			key, value := extractKeyValue(part, "=")
			if value == "" {
				exists = append(exists, key)
			} else {
				fillMap(&inOps, key, value)
			}
		}
	}
	return inOps, notInOps, exists, doesNotExists
}

func extractKeyValue(s, separator string) (string, string) {
	parts := strings.SplitN(s, separator, 2)
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
}

func fillMap(m *map[string][]string, key, value string) {
	var temp []string
	err := json.Unmarshal([]byte(value), &temp)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	(*m)[key] = temp
}

func generateNodeSelectorRequirement(inOps, notInOps map[string][]string, exists, doesNotExists []string) []v1.NodeSelectorRequirement {

	var nodeSelectorRequirements []v1.NodeSelectorRequirement
	for key, value := range inOps {
		nodeSelectorRequirements = append(nodeSelectorRequirements, v1.NodeSelectorRequirement{
			Key:      key,
			Operator: v1.NodeSelectorOpIn,
			Values:   value,
		})
	}

	for key, value := range notInOps {
		nodeSelectorRequirements = append(nodeSelectorRequirements, v1.NodeSelectorRequirement{
			Key:      key,
			Operator: v1.NodeSelectorOpNotIn,
			Values:   value,
		})
	}

	for _, value := range exists {
		nodeSelectorRequirements = append(nodeSelectorRequirements, v1.NodeSelectorRequirement{
			Key:      value,
			Operator: v1.NodeSelectorOpExists,
		})
	}

	for _, value := range doesNotExists {
		nodeSelectorRequirements = append(nodeSelectorRequirements, v1.NodeSelectorRequirement{
			Key:      value,
			Operator: v1.NodeSelectorOpDoesNotExist,
		})
	}

	return nodeSelectorRequirements
}

func generateLabelSelectorRequirement(inOps, notInOps map[string][]string, exists, doesNotExists []string) []metav1.LabelSelectorRequirement {

	var labelSelectorRequirements []metav1.LabelSelectorRequirement
	for key, value := range inOps {
		labelSelectorRequirements = append(labelSelectorRequirements, metav1.LabelSelectorRequirement{
			Key:      key,
			Operator: metav1.LabelSelectorOpIn,
			Values:   value,
		})
	}

	for key, value := range notInOps {
		labelSelectorRequirements = append(labelSelectorRequirements, metav1.LabelSelectorRequirement{
			Key:      key,
			Operator: metav1.LabelSelectorOpNotIn,
			Values:   value,
		})
	}

	for _, value := range exists {
		labelSelectorRequirements = append(labelSelectorRequirements, metav1.LabelSelectorRequirement{
			Key:      value,
			Operator: metav1.LabelSelectorOpExists,
		})
	}

	for _, value := range doesNotExists {
		labelSelectorRequirements = append(labelSelectorRequirements, metav1.LabelSelectorRequirement{
			Key:      value,
			Operator: metav1.LabelSelectorOpDoesNotExist,
		})
	}

	return labelSelectorRequirements
}
