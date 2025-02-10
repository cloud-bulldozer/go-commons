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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	k8sconnector "github.com/cloud-bulldozer/go-commons/k8s-connector"
)

func GetDataImportCronSourceFormatForStorageClass(k8sConnector k8sconnector.K8SConnector, storageClassName string) (cdiv1beta1.DataImportCronSourceFormat, error) {
	storageProfile, err := GetStorageProfileForStorageClass(k8sConnector, storageClassName)
	if err != nil {
		return "", err
	}

	status, found, err := unstructured.NestedMap(storageProfile.Object, "status")
	if err != nil {
		return "", err
	}
	if !found {
		return "", fmt.Errorf("StorageProfile object does not have a status field")
	}

	sourceFormat, ok := status["dataImportCronSourceFormat"]
	if !ok {
		return "", fmt.Errorf("status field of StorageProfile object does not have a dataImportCronSourceFormat field")
	}
	return cdiv1beta1.DataImportCronSourceFormat(sourceFormat.(string)), nil
}

func GetStorageProfileForStorageClass(k8sConnector k8sconnector.K8SConnector, storageClassName string) (*unstructured.Unstructured, error) {
	storageProfileGVR := schema.GroupVersionResource{Group: "cdi.kubevirt.io", Version: "v1beta1", Resource: "storageprofiles"}
	return k8sConnector.DynamicClient().Resource(storageProfileGVR).Get(context.TODO(), storageClassName, metav1.GetOptions{})
}
