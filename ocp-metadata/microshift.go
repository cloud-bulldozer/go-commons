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
	"context"
	"fmt"
	"regexp"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	microShiftVersionNamespace = "kube-public"
	microShiftVersionConfigMap = "microshift-version"

	apiGroupOpenShiftConfig = "config.openshift.io"
	apiGroupOpenShiftRoute  = "route.openshift.io"

	// DistributionKubernetes identifies a vanilla Kubernetes cluster.
	DistributionKubernetes = "kubernetes"
	// DistributionOpenShift identifies an OpenShift cluster.
	DistributionOpenShift = "openshift"
	// DistributionMicroShift identifies a MicroShift cluster.
	DistributionMicroShift = "microshift"
)

type microShiftVersionInfo struct {
	version    string
	majorMinor string
}

var microShiftMajorMinorRe = regexp.MustCompile(`^[0-9]+\.[0-9]+`)

func (meta *Metadata) detectDistribution() (string, microShiftVersionInfo, error) {
	cm, err := meta.readMicroShiftVersionConfigMap()
	switch {
	case err == nil:
		info, _ := parseMicroShiftVersion(cm)
		return DistributionMicroShift, info, nil
	case apierrors.IsNotFound(err):
		// Expected on OpenShift and vanilla Kubernetes.
	default:
		// Keep metadata collection best-effort when kube-public is unreadable.
	}

	groups, err := meta.connector.ClientSet().Discovery().ServerGroups()
	if err != nil {
		return DistributionKubernetes, microShiftVersionInfo{}, err
	}
	apiGroups := make(map[string]bool, len(groups.Groups))
	for _, group := range groups.Groups {
		apiGroups[group.Name] = true
	}

	if apiGroups[apiGroupOpenShiftConfig] {
		return DistributionOpenShift, microShiftVersionInfo{}, nil
	}
	if apiGroups[apiGroupOpenShiftRoute] {
		return DistributionMicroShift, microShiftVersionInfo{}, nil
	}
	return DistributionKubernetes, microShiftVersionInfo{}, nil
}

func (meta *Metadata) readMicroShiftVersionConfigMap() (*corev1.ConfigMap, error) {
	return meta.connector.ClientSet().CoreV1().ConfigMaps(microShiftVersionNamespace).Get(context.TODO(), microShiftVersionConfigMap, metav1.GetOptions{})
}

func parseMicroShiftVersion(cm *corev1.ConfigMap) (microShiftVersionInfo, error) {
	version := cm.Data["version"]
	if version == "" {
		return microShiftVersionInfo{}, fmt.Errorf("%s/%s ConfigMap has no version field", microShiftVersionNamespace, microShiftVersionConfigMap)
	}
	info := microShiftVersionInfo{version: version}
	major := cm.Data["major"]
	minor := cm.Data["minor"]
	if major != "" && minor != "" {
		info.majorMinor = major + "." + minor
	} else {
		info.majorMinor = microShiftMajorMinorRe.FindString(version)
	}
	return info, nil
}
