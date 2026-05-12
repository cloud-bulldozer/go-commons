// Copyright 2026 The go-commons Authors.
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
	"errors"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/fake"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes"
	k8sfake "k8s.io/client-go/kubernetes/fake"
	"k8s.io/client-go/rest"
	k8stesting "k8s.io/client-go/testing"
)

type fakeConnector struct {
	clientSet     kubernetes.Interface
	dynamicClient dynamic.Interface
	restConfig    *rest.Config
}

func (f fakeConnector) ClientSet() kubernetes.Interface {
	return f.clientSet
}

func (f fakeConnector) RestConfig() *rest.Config {
	return f.restConfig
}

func (f fakeConnector) DynamicClient() dynamic.Interface {
	return f.dynamicClient
}

func TestDetectDistribution(t *testing.T) {
	tests := []struct {
		name             string
		apiGroups        []string
		configMapData    map[string]string
		discoveryError   bool
		wantDistribution string
		wantVersion      string
		wantMajorMinor   string
		wantErr          bool
	}{
		{
			name:      "microshift configmap with version fields",
			apiGroups: []string{apiGroupOpenShiftRoute},
			configMapData: map[string]string{
				"version": "4.22.0~rc.2",
				"major":   "4",
				"minor":   "22",
			},
			wantDistribution: DistributionMicroShift,
			wantVersion:      "4.22.0~rc.2",
			wantMajorMinor:   "4.22",
		},
		{
			name: "microshift configmap with version only",
			configMapData: map[string]string{
				"version": "4.22.0~rc.2",
			},
			wantDistribution: DistributionMicroShift,
			wantVersion:      "4.22.0~rc.2",
			wantMajorMinor:   "4.22",
		},
		{
			name:             "malformed microshift configmap still detects microshift",
			configMapData:    map[string]string{"major": "4", "minor": "22"},
			wantDistribution: DistributionMicroShift,
		},
		{
			name:             "microshift route without config heuristic",
			apiGroups:        []string{apiGroupOpenShiftRoute},
			wantDistribution: DistributionMicroShift,
		},
		{
			name:             "openshift",
			apiGroups:        []string{apiGroupOpenShiftConfig, apiGroupOpenShiftRoute},
			wantDistribution: DistributionOpenShift,
		},
		{
			name:             "microshift configmap wins over openshift api",
			apiGroups:        []string{apiGroupOpenShiftConfig, apiGroupOpenShiftRoute},
			configMapData:    map[string]string{"version": "4.22.0~rc.2", "major": "4", "minor": "22"},
			wantDistribution: DistributionMicroShift,
			wantVersion:      "4.22.0~rc.2",
			wantMajorMinor:   "4.22",
		},
		{
			name:             "microshift configmap works when discovery fails",
			configMapData:    map[string]string{"version": "4.22.0~rc.2", "major": "4", "minor": "22"},
			discoveryError:   true,
			wantDistribution: DistributionMicroShift,
			wantVersion:      "4.22.0~rc.2",
			wantMajorMinor:   "4.22",
		},
		{
			name:           "discovery error without microshift configmap returns error",
			discoveryError: true,
			wantErr:        true,
		},
		{
			name:             "kubernetes",
			wantDistribution: DistributionKubernetes,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			meta := newTestMetadata(t, tt.apiGroups, tt.configMapData)
			if tt.discoveryError {
				injectDiscoveryError(t, meta)
			}
			gotDistribution, gotVersion, err := meta.detectDistribution()
			if tt.wantErr {
				if err == nil {
					t.Fatal("detectDistribution returned nil error, want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("detectDistribution returned unexpected error: %v", err)
			}
			if gotDistribution != tt.wantDistribution {
				t.Fatalf("distribution = %q, want %q", gotDistribution, tt.wantDistribution)
			}
			if gotVersion.version != tt.wantVersion {
				t.Fatalf("version = %q, want %q", gotVersion.version, tt.wantVersion)
			}
			if gotVersion.majorMinor != tt.wantMajorMinor {
				t.Fatalf("majorMinor = %q, want %q", gotVersion.majorMinor, tt.wantMajorMinor)
			}
		})
	}
}

func injectDiscoveryError(t *testing.T, meta Metadata) {
	t.Helper()

	clientSet, ok := meta.connector.ClientSet().(*k8sfake.Clientset)
	if !ok {
		t.Fatal("unexpected clientset type")
	}
	clientSet.PrependReactor("get", "group", func(k8stesting.Action) (bool, runtime.Object, error) {
		return true, nil, errors.New("discovery failed")
	})
}

func TestGetClusterMetadataMicroShift(t *testing.T) {
	meta := newTestMetadata(t, []string{apiGroupOpenShiftRoute}, map[string]string{
		"version": "4.22.0~rc.2",
		"major":   "4",
		"minor":   "22",
	})

	clusterMetadata, err := meta.GetClusterMetadata()
	if err != nil {
		t.Fatalf("GetClusterMetadata returned unexpected error: %v", err)
	}
	if clusterMetadata.Distribution != DistributionMicroShift {
		t.Fatalf("Distribution = %q, want %q", clusterMetadata.Distribution, DistributionMicroShift)
	}
	if !clusterMetadata.MicroShift {
		t.Fatal("MicroShift = false, want true")
	}
	if clusterMetadata.MicroShiftVersion != "4.22.0~rc.2" {
		t.Fatalf("MicroShiftVersion = %q, want %q", clusterMetadata.MicroShiftVersion, "4.22.0~rc.2")
	}
	if clusterMetadata.MicroShiftMajorVersion != "4.22" {
		t.Fatalf("MicroShiftMajorVersion = %q, want %q", clusterMetadata.MicroShiftMajorVersion, "4.22")
	}
	if clusterMetadata.K8SVersion != "v1.35.3" {
		t.Fatalf("K8SVersion = %q, want %q", clusterMetadata.K8SVersion, "v1.35.3")
	}
	if clusterMetadata.TotalNodes != 1 {
		t.Fatalf("TotalNodes = %d, want %d", clusterMetadata.TotalNodes, 1)
	}
}

func TestGetClusterMetadataOpenShift(t *testing.T) {
	meta := newOpenShiftTestMetadata(t)

	clusterMetadata, err := meta.GetClusterMetadata()
	if err != nil {
		t.Fatalf("GetClusterMetadata returned unexpected error: %v", err)
	}
	if clusterMetadata.Distribution != DistributionOpenShift {
		t.Fatalf("Distribution = %q, want %q", clusterMetadata.Distribution, DistributionOpenShift)
	}
	if clusterMetadata.MicroShift {
		t.Fatal("MicroShift = true, want false")
	}
	if clusterMetadata.OCPVersion != "4.18.9" {
		t.Fatalf("OCPVersion = %q, want %q", clusterMetadata.OCPVersion, "4.18.9")
	}
	if clusterMetadata.OCPMajorVersion != "4.18" {
		t.Fatalf("OCPMajorVersion = %q, want %q", clusterMetadata.OCPMajorVersion, "4.18")
	}
	if clusterMetadata.K8SVersion != "v1.31.6" {
		t.Fatalf("K8SVersion = %q, want %q", clusterMetadata.K8SVersion, "v1.31.6")
	}
	if clusterMetadata.ClusterName != "ocp-test-abcde" {
		t.Fatalf("ClusterName = %q, want %q", clusterMetadata.ClusterName, "ocp-test-abcde")
	}
	if clusterMetadata.Platform != "AWS" {
		t.Fatalf("Platform = %q, want %q", clusterMetadata.Platform, "AWS")
	}
	if clusterMetadata.ClusterType != "rosa" {
		t.Fatalf("ClusterType = %q, want %q", clusterMetadata.ClusterType, "rosa")
	}
	if clusterMetadata.Region != "us-east-1" {
		t.Fatalf("Region = %q, want %q", clusterMetadata.Region, "us-east-1")
	}
	if clusterMetadata.SDNType != "OVNKubernetes" {
		t.Fatalf("SDNType = %q, want %q", clusterMetadata.SDNType, "OVNKubernetes")
	}
	if !clusterMetadata.Fips {
		t.Fatal("Fips = false, want true")
	}
	if clusterMetadata.Publish != "External" {
		t.Fatalf("Publish = %q, want %q", clusterMetadata.Publish, "External")
	}
	if clusterMetadata.WorkerArch != "amd64" {
		t.Fatalf("WorkerArch = %q, want %q", clusterMetadata.WorkerArch, "amd64")
	}
	if clusterMetadata.ControlPlaneArch != "amd64" {
		t.Fatalf("ControlPlaneArch = %q, want %q", clusterMetadata.ControlPlaneArch, "amd64")
	}
	if !clusterMetadata.Ipsec {
		t.Fatal("Ipsec = false, want true")
	}
	if clusterMetadata.IpsecMode != "Full" {
		t.Fatalf("IpsecMode = %q, want %q", clusterMetadata.IpsecMode, "Full")
	}
	if clusterMetadata.TotalNodes != 2 {
		t.Fatalf("TotalNodes = %d, want %d", clusterMetadata.TotalNodes, 2)
	}
	if clusterMetadata.MasterNodesCount != 1 {
		t.Fatalf("MasterNodesCount = %d, want %d", clusterMetadata.MasterNodesCount, 1)
	}
	if clusterMetadata.WorkerNodesCount != 1 {
		t.Fatalf("WorkerNodesCount = %d, want %d", clusterMetadata.WorkerNodesCount, 1)
	}
}

func newTestMetadata(t *testing.T, apiGroups []string, configMapData map[string]string) Metadata {
	t.Helper()

	objects := []runtime.Object{
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "microshift",
				Labels: map[string]string{
					"node-role.kubernetes.io/master": "",
					"node-role.kubernetes.io/worker": "",
				},
			},
		},
	}
	if configMapData != nil {
		objects = append(objects, newMicroShiftVersionConfigMap(configMapData))
	}

	clientSet := k8sfake.NewClientset(objects...)
	discovery, ok := clientSet.Discovery().(*fake.FakeDiscovery)
	if !ok {
		t.Fatal("unexpected discovery type")
	}
	discovery.FakedServerVersion = &version.Info{GitVersion: "v1.35.3"}
	for _, group := range apiGroups {
		discovery.Resources = append(discovery.Resources, &metav1.APIResourceList{
			GroupVersion: group + "/v1",
		})
	}

	return Metadata{
		connector: fakeConnector{
			clientSet:     clientSet,
			dynamicClient: dynamicfake.NewSimpleDynamicClient(runtime.NewScheme()),
			restConfig:    &rest.Config{},
		},
	}
}

func newOpenShiftTestMetadata(t *testing.T) Metadata {
	t.Helper()

	clientSet := k8sfake.NewClientset(
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-0",
				Labels: map[string]string{
					"node-role.kubernetes.io/master":   "",
					"node.kubernetes.io/instance-type": "m6i.2xlarge",
				},
			},
		},
		&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: "worker-0",
				Labels: map[string]string{
					"node-role.kubernetes.io/worker":   "",
					"node.kubernetes.io/instance-type": "m6i.xlarge",
				},
			},
		},
		&corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "kube-system",
				Name:      "cluster-config-v1",
			},
			Data: map[string]string{
				"install-config": `
publish: External
fips: true
compute:
- name: worker
  architecture: amd64
controlPlane:
  architecture: amd64
`,
			},
		},
	)
	discovery, ok := clientSet.Discovery().(*fake.FakeDiscovery)
	if !ok {
		t.Fatal("unexpected discovery type")
	}
	discovery.FakedServerVersion = &version.Info{GitVersion: "v1.31.6"}
	for _, group := range []string{apiGroupOpenShiftConfig, apiGroupOpenShiftRoute} {
		discovery.Resources = append(discovery.Resources, &metav1.APIResourceList{
			GroupVersion: group + "/v1",
		})
	}

	return Metadata{
		connector: fakeConnector{
			clientSet: clientSet,
			dynamicClient: dynamicfake.NewSimpleDynamicClient(
				runtime.NewScheme(),
				newClusterVersion("4.18.9"),
				newInfrastructure(),
				newConfigNetwork(),
				newOperatorNetwork(),
			),
			restConfig: &rest.Config{},
		},
	}
}

func newMicroShiftVersionConfigMap(data map[string]string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: microShiftVersionNamespace,
			Name:      microShiftVersionConfigMap,
		},
		Data: data,
	}
}

func newClusterVersion(version string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "config.openshift.io/v1",
			"kind":       "ClusterVersion",
			"metadata": map[string]interface{}{
				"name": "version",
			},
			"status": map[string]interface{}{
				"history": []interface{}{
					map[string]interface{}{
						"state":   completedUpdate,
						"version": version,
					},
				},
			},
		},
	}
}

func newInfrastructure() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "config.openshift.io/v1",
			"kind":       "Infrastructure",
			"metadata": map[string]interface{}{
				"name": "cluster",
			},
			"status": map[string]interface{}{
				"infrastructureName": "ocp-test-abcde",
				"platform":           "AWS",
				"platformStatus": map[string]interface{}{
					"aws": map[string]interface{}{
						"region": "us-east-1",
						"resourceTags": []interface{}{
							map[string]interface{}{
								"key":   "red-hat-clustertype",
								"value": "rosa",
							},
						},
					},
					"type": "AWS",
				},
			},
		},
	}
}

func newConfigNetwork() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "config.openshift.io/v1",
			"kind":       "Network",
			"metadata": map[string]interface{}{
				"name": "cluster",
			},
			"status": map[string]interface{}{
				"networkType": "OVNKubernetes",
			},
		},
	}
}

func newOperatorNetwork() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "operator.openshift.io/v1",
			"kind":       "Network",
			"metadata": map[string]interface{}{
				"name": "cluster",
			},
			"spec": map[string]interface{}{
				"defaultNetwork": map[string]interface{}{
					"ovnKubernetesConfig": map[string]interface{}{
						"ipsecConfig": map[string]interface{}{
							"mode": "Full",
						},
					},
				},
			},
		},
	}
}
