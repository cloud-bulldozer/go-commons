// Copyright 2023 The go-commons Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ocpmetadata

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"

	"gopkg.in/yaml.v3"
	authenticationv1 "k8s.io/api/authentication/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/utils/ptr"

	k8sconnector "github.com/cloud-bulldozer/go-commons/v2/k8s-connector"
)

// Metadata object
type Metadata struct {
	connector k8sconnector.K8SConnector
}

// NewMetadata instantiates a new OCP metadata discovery agent
func NewMetadata(restConfig *rest.Config) (Metadata, error) {
	k8sConnector, err := k8sconnector.NewK8SConnector(restConfig)
	if err != nil {
		return Metadata{}, err
	}
	return Metadata{
		connector: k8sConnector,
	}, err
}

// GetClusterMetadata returns a clusterMetadata object from the given OCP cluster
func (meta *Metadata) GetClusterMetadata() (ClusterMetadata, error) {
	metadata := ClusterMetadata{}
	infra, err := meta.getInfraDetails()
	if err != nil {
		return metadata, nil
	}
	version, err := meta.getVersionInfo()
	if err != nil {
		return metadata, err
	}
	metadata.OCPVersion, metadata.OCPMajorVersion, metadata.K8SVersion = version.ocpVersion, version.ocpMajorVersion, version.k8sVersion
	if meta.getNodesInfo(&metadata) != nil {
		return metadata, err
	}
	if infra != nil {
		metadata.ClusterName, metadata.Platform, metadata.Region = infra.Status.InfrastructureName, infra.Status.Platform, infra.Status.PlatformStatus.Aws.Region
		metadata.ClusterType = "self-managed"
		for _, v := range infra.Status.PlatformStatus.Aws.ResourceTags {
			if v.Key == "red-hat-clustertype" {
				metadata.ClusterType = v.Value
			}
		}
		metadata.SDNType, err = meta.getSDNInfo()
		if err != nil {
			return metadata, err
		}

		// Get InstallConfig to use in multiple methods
		installConfig, err := meta.getClusterConfig()
		if err != nil {
			return metadata, err
		}

		metadata.Fips, err = meta.getFips(installConfig)
		if err != nil {
			return metadata, err
		}
		metadata.Publish, err = meta.getPublish(installConfig)
		if err != nil {
			return metadata, err
		}
		metadata.WorkerArch, err = meta.getComputeWorkerArch(installConfig)
		if err != nil {
			return metadata, err
		}
		metadata.ControlPlaneArch, err = meta.getControlPlaneArch(installConfig)
		if err != nil {
			return metadata, err
		}
		metadata.Ipsec, metadata.IpsecMode, err = meta.getIPSec()
		if err != nil {
			return metadata, err
		}
	}
	return metadata, err
}

// GetPrometheus Returns Prometheus URL and a valid Bearer token
func (meta *Metadata) GetPrometheus() (string, string, error) {
	prometheusURL, err := getPrometheusURL(meta.connector.DynamicClient())
	if err != nil {
		return prometheusURL, "", err
	}
	prometheusToken, err := getBearerToken(meta.connector.ClientSet())
	return prometheusURL, prometheusToken, err
}

// GetCurrentPodCount returns the number of current running pods across all worker nodes
func (meta *Metadata) GetCurrentPodCount() (int, error) {
	var podCount int
	nodeList, err := meta.connector.ClientSet().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: workerNodeSelector})
	if err != nil {
		return podCount, err
	}
	podList, err := meta.connector.ClientSet().CoreV1().Pods(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{FieldSelector: "status.phase=" + running})
	if err != nil {
		return podCount, err
	}
	for _, pod := range podList.Items {
		for _, node := range nodeList.Items {
			if pod.Spec.NodeName == node.Name {
				podCount++
				break
			}
		}
	}
	return podCount, nil
}

// Returns the number of current running VMIs in the cluster
func (meta *Metadata) GetCurrentVMICount() (int, error) {
	var vmiCount int
	vmis, err := meta.connector.DynamicClient().Resource(vmiGVR).Namespace(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return vmiCount, err
	}
	for _, vmi := range vmis.Items {
		status, found, err := unstructured.NestedString(vmi.UnstructuredContent(), "status", "phase")
		if !found {
			return vmiCount, fmt.Errorf("phase field not found in kubevirt.io/v1/namespaces/%s/virtualmachineinstances/%s status", vmi.GetNamespace(), vmi.GetName())
		}
		if err != nil {
			return vmiCount, err
		}
		if status == running {
			vmiCount++
		}
	}
	return vmiCount, nil
}

// GetDefaultIngressDomain returns default ingress domain of the default ingress controller
func (meta *Metadata) GetDefaultIngressDomain() (string, error) {
	ingressController, err := meta.connector.DynamicClient().Resource(ingressControllerGRV).
		Namespace("openshift-ingress-operator").Get(context.TODO(), "default", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	ingressDomain, found, err := unstructured.NestedString(ingressController.UnstructuredContent(), "status", "domain")
	if !found {
		return "", fmt.Errorf("domain field not found in operator.openshift.io/v1/namespaces/openshift-ingress-operator/ingresscontrollers/default status")
	}
	return ingressDomain, err
}

// getPrometheusURL Returns a valid prometheus endpoint from the openshift-monitoring/prometheus-k8s route
func getPrometheusURL(dynamicClient dynamic.Interface) (string, error) {
	route, err := dynamicClient.Resource(routeGVR).Namespace(monitoringNs).Get(context.TODO(), "prometheus-k8s", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	prometheusHost, found, err := unstructured.NestedString(route.UnstructuredContent(), "spec", "host")
	if !found {
		return "", fmt.Errorf("host field not found in %s/prometheus-k8s route spec", monitoringNs)
	}
	if err != nil {
		return "", err
	}
	return "https://" + prometheusHost, nil
}

// getBearerToken returns a valid bearer token from the openshift-monitoring/prometheus-k8s service account
func getBearerToken(clientset kubernetes.Interface) (string, error) {
	request := authenticationv1.TokenRequest{
		Spec: authenticationv1.TokenRequestSpec{
			ExpirationSeconds: ptr.To(int64(tokenExpiration.Seconds())),
		},
	}
	response, err := clientset.CoreV1().ServiceAccounts(monitoringNs).CreateToken(context.TODO(), "prometheus-k8s", &request, metav1.CreateOptions{})
	return response.Status.Token, err
}

// getInfraDetails returns a pointer to an infrastructure object or nil
func (meta *Metadata) getInfraDetails() (*infraObj, error) {
	var infraJSON infraObj
	infra, err := meta.connector.DynamicClient().Resource(schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "infrastructures",
	}).Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		// If the infrastructure resource is not found we assume this is not an OCP cluster
		if errors.IsNotFound(err) {
			return nil, nil
		}
		return &infraJSON, err
	}
	infraData, _ := infra.MarshalJSON()
	err = json.Unmarshal(infraData, &infraJSON)
	return &infraJSON, err
}

// getVersionInfo obtains OCP and k8s version information
func (meta *Metadata) getVersionInfo() (versionObj, error) {
	var cv clusterVersion
	var versionInfo versionObj
	version, err := meta.connector.ClientSet().Discovery().ServerVersion()
	if err != nil {
		return versionInfo, err
	}
	versionInfo.k8sVersion = version.GitVersion
	clusterVersion, err := meta.connector.DynamicClient().Resource(
		schema.GroupVersionResource{
			Group:    "config.openshift.io",
			Version:  "v1",
			Resource: "clusterversions",
		}).Get(context.TODO(), "version", metav1.GetOptions{})
	if err != nil {
		// If the clusterversion resource is not found we assume this is not an OCP cluster
		if errors.IsNotFound(err) {
			return versionInfo, nil
		}
		return versionInfo, err
	}
	clusterVersionBytes, _ := clusterVersion.MarshalJSON()
	err = json.Unmarshal(clusterVersionBytes, &cv)
	if err != nil {
		return versionInfo, err
	}
	for _, update := range cv.Status.History {
		if update.State == completedUpdate {
			// obtain the version from the last completed update
			versionInfo.ocpVersion = update.Version
			break
		}
	}
	shortReg, _ := regexp.Compile(`([0-9]\.[0-9]+)-*`)
	versionInfo.ocpMajorVersion = shortReg.FindString(versionInfo.ocpVersion)
	return versionInfo, err
}

// getNodesInfo returns node information
func (meta *Metadata) getNodesInfo(clusterMetadata *ClusterMetadata) error {
	nodes, err := meta.connector.ClientSet().CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	clusterMetadata.TotalNodes = len(nodes.Items)
	// When the master label is found, the node is considered a master, regardless of other labels the node could have
	for _, node := range nodes.Items {
		if _, ok := node.Labels["node-role.kubernetes.io/master"]; ok { // Check for master role
			clusterMetadata.MasterNodesCount++
			clusterMetadata.MasterNodesType = node.Labels["node.kubernetes.io/instance-type"]
			if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok {
				if len(node.Spec.Taints) == 0 { // When mastersSchedulable is false, master nodes have at least one taint
					clusterMetadata.WorkerNodesCount++
				}
			}
		} else if _, ok := node.Labels["node-role.kubernetes.io/control-plane"]; ok { // Check for control-plane role
			clusterMetadata.MasterNodesCount++
			clusterMetadata.MasterNodesType = node.Labels["node.kubernetes.io/instance-type"]
		} else if _, ok := node.Labels["node-role.kubernetes.io/infra"]; ok { // Check for infra role
			clusterMetadata.InfraNodesCount++
			clusterMetadata.InfraNodesType = node.Labels["node.kubernetes.io/instance-type"]
		} else if _, ok := node.Labels["node-role.kubernetes.io/worker"]; ok { // Check for worker role
			clusterMetadata.WorkerNodesCount++
			clusterMetadata.WorkerNodesType = node.Labels["node.kubernetes.io/instance-type"]
		} else {
			clusterMetadata.OtherNodesCount++
		}
	}
	return err
}

// getSDNInfo returns SDN type
func (meta *Metadata) getSDNInfo() (string, error) {
	networkData, err := meta.connector.DynamicClient().Resource(schema.GroupVersionResource{
		Group:    "config.openshift.io",
		Version:  "v1",
		Resource: "networks",
	}).Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	networkType, found, err := unstructured.NestedString(networkData.UnstructuredContent(), "status", "networkType")
	if !found {
		return "", fmt.Errorf("networkType field not found in config.openshift.io/v1/network/networks/cluster status")
	}
	return networkType, err
}

func (meta *Metadata) getPublish(installConfig map[string]interface{}) (string, error) {
	if val, ok := installConfig["publish"]; ok {
		return val.(string), nil
	}
	return "", nil
}

func (meta *Metadata) getFips(installConfig map[string]interface{}) (bool, error) {
	if val, ok := installConfig["fips"]; ok {
		return val.(bool), nil
	}
	return false, nil
}

func (meta *Metadata) getComputeWorkerArch(installConfig map[string]interface{}) (string, error) {
	if val, ok := installConfig["compute"]; ok {
		for _, val := range val.([]interface{}) {
			comConfig := val.(map[string]interface{})
			if v, ok := comConfig["name"].(string); ok {
				if v == "worker" {
					return comConfig["architecture"].(string), nil
				}
			}
		}
	}
	return "", nil
}

func (meta *Metadata) getControlPlaneArch(installConfig map[string]interface{}) (string, error) {
	if val, ok := installConfig["controlPlane"]; ok {
		cpConfig := val.(map[string]interface{})
		if v, ok := cpConfig["architecture"].(string); ok {
			return v, nil
		}
	}
	return "", nil
}

// getIPSec returns if the cluster has IPSec enabled
func (meta *Metadata) getIPSec() (bool, string, error) {
	ipsecType := "Disabled"
	networks, err := meta.connector.DynamicClient().Resource(schema.GroupVersionResource{
		Group:    "operator.openshift.io",
		Version:  "v1",
		Resource: "networks",
	}).Get(context.TODO(), "cluster", metav1.GetOptions{})
	if err != nil {
		return false, ipsecType, err
	}
	ipsecMode, found, err := unstructured.NestedString(networks.UnstructuredContent(), "spec", "defaultNetwork", "ovnKubernetesConfig", "ipsecConfig", "mode")
	if !found {
		_, found, _ := unstructured.NestedMap(networks.UnstructuredContent(), "spec", "defaultNetwork", "ovnKubernetesConfig", "ipsecConfig")
		if !found {
			return false, ipsecType, nil
		} else {
			ipsecType = "Full"
			return true, ipsecType, nil
		}
	}
	if err != nil {
		return false, ipsecType, err
	}
	if ipsecMode != "Disabled" {
		return true, ipsecMode, nil
	}
	return false, ipsecMode, nil
}

// getClusterConfig returns cluster configuration yaml
func (meta *Metadata) getClusterConfig() (map[string]interface{}, error) {
	config, err := meta.connector.ClientSet().CoreV1().ConfigMaps("kube-system").Get(context.TODO(), "cluster-config-v1", metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	installConfigRaw := config.Data["install-config"]
	installConfig, err := toMap(installConfigRaw)
	if err != nil {
		return nil, err
	}
	return installConfig, nil
}

func toMap(str string) (map[string]interface{}, error) {
	config := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(str), &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func (meta *Metadata) GetOCPVirtualizationVersion() (string, error) {
	virtOp, err := meta.connector.ClientSet().AppsV1().Deployments("openshift-cnv").Get(context.TODO(), "virt-operator", metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if virtOpVers, ok := virtOp.Labels["app.kubernetes.io/version"]; ok {
		return virtOpVers, nil
	} else {
		return "", fmt.Errorf("label app.kubernetes.io/version not found in virt-operator deployment")
	}
}
