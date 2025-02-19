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
	"fmt"

	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	k8sconnector "github.com/cloud-bulldozer/go-commons/v2/k8s-connector"
)

const (
	defaultStorageClassAnnotation     = "storageclass.kubernetes.io/is-default-class"
	defaultVirtStorageClassAnnotation = "storageclass.kubevirt.io/is-default-virt-class"
)

// StorageClassExists Check if a storageClass with this name exists.
func StorageClassExists(k8sConnector k8sconnector.K8SConnector, storageClassName string) (bool, error) {
	_, err := k8sConnector.ClientSet().StorageV1().StorageClasses().Get(context.Background(), storageClassName, metav1.GetOptions{})
	if err != nil {
		if kerrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// GetDefaultStorageClassName returns the name of the default storage class
// If preferVirt is false look for the storageclass.kubernetes.io/is-default-class annotation
// If preferVirt is true look for the storageclass.kubevirt.io/is-default-virt-class annotation
func GetDefaultStorageClassName(k8sConnector k8sconnector.K8SConnector, preferVirt bool) (string, error) {
	storageClasses, err := k8sConnector.ClientSet().StorageV1().StorageClasses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return "", err
	}

	var defaultStorageClassName, defaultVirtStorageClassName, storageClassName string
	for _, storageClass := range storageClasses.Items {
		if val, ok := storageClass.Annotations[defaultStorageClassAnnotation]; ok && val == "true" {
			defaultStorageClassName = storageClass.GetName()
		}
		if val, ok := storageClass.Annotations[defaultVirtStorageClassAnnotation]; ok && val == "true" {
			defaultVirtStorageClassName = storageClass.GetName()
		}
	}
	if preferVirt {
		storageClassName = defaultVirtStorageClassName
	}
	if storageClassName == "" {
		storageClassName = defaultStorageClassName
	}
	if storageClassName == "" {
		return "", fmt.Errorf("no default StorageClass was set")
	}
	return storageClassName, nil
}

// StorageClassSupportsVolumeExpansion returns true if the storageClass allows Volume Expansion
func StorageClassSupportsVolumeExpansion(k8sConnector k8sconnector.K8SConnector, storageClassName string) (bool, error) {
	storageClass, err := k8sConnector.ClientSet().StorageV1().StorageClasses().Get(context.Background(), storageClassName, metav1.GetOptions{})
	if err != nil {
		return false, err
	}
	return storageClass.AllowVolumeExpansion != nil && *storageClass.AllowVolumeExpansion, nil
}

// If storageClassName was provided, verify that it exists. Otherwise, get the default
func GetStorageClassName(k8sConnector k8sconnector.K8SConnector, storageClassName string, preferVirt bool) (string, error) {
	if storageClassName != "" {
		exists, err := StorageClassExists(k8sConnector, storageClassName)
		if err != nil || !exists {
			return "", err
		}
		return storageClassName, nil
	}
	return GetDefaultStorageClassName(k8sConnector, preferVirt)
}

func getStorageClassProvisioner(k8sConnector k8sconnector.K8SConnector, storageClassName string) (string, error) {
	storageClass, err := k8sConnector.ClientSet().StorageV1().StorageClasses().Get(context.Background(), storageClassName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}

	return storageClass.Provisioner, nil
}
