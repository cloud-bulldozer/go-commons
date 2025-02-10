// Copyright 2025 The go-commons Authors.
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

package k8sstorage

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	k8sconnector "github.com/cloud-bulldozer/go-commons/k8s-connector"
)

func getVolumeSnapshotClassNameForProvisioner(k8sConnector k8sconnector.K8SConnector, provisioner string) (string, error) {
	volumeSnapshotClassGVR := schema.GroupVersionResource{Group: "snapshot.storage.k8s.io", Version: "v1", Resource: "volumesnapshotclasses"}
	itemList, err := k8sConnector.DynamicClient().Resource(volumeSnapshotClassGVR).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}
	var snapshotClassName string
	for _, item := range itemList.Items {
		driver, found, err := unstructured.NestedString(item.Object, "driver")
		if err != nil {
			return "", err
		}
		if !found {
			continue
		}
		if driver == provisioner {
			snapshotClassName = item.GetName()
			break
		}
	}
	return snapshotClassName, nil
}

// GetVolumeSnapshotClassNameForStorageClass returns the name VolumeSnapshotClass with the same provisioner as that of the storageClass
func GetVolumeSnapshotClassNameForStorageClass(k8sConnector k8sconnector.K8SConnector, storageClassName string) (string, error) {
	provisioner, err := getStorageClassProvisioner(k8sConnector, storageClassName)
	if err != nil {
		return "", err
	}
	return getVolumeSnapshotClassNameForProvisioner(k8sConnector, provisioner)
}
