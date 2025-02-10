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
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/fake"
	cdiv1beta1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

	"github.com/cloud-bulldozer/go-commons/mocks"
)

var _ = Describe("Tests for K8S Storage Profile", func() {
	var (
		mockCtrl         *gomock.Controller
		mockK8SConnector *mocks.MockK8SConnector
		scheme           *runtime.Scheme
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()

		// Add CDI objects to the scheme
		metav1.AddToGroupVersion(scheme, metav1.SchemeGroupVersion)
		scheme.AddKnownTypes(cdiv1beta1.SchemeGroupVersion, &cdiv1beta1.StorageProfile{})

		mockCtrl = gomock.NewController(GinkgoT())
		mockK8SConnector = mocks.NewMockK8SConnector(mockCtrl)
	})

	Context("GetStorageProfileForStorageClass", func() {
		var (
			storageProfileName = "main-sp"
		)
		BeforeEach(func() {
			dynamicClient := fake.NewSimpleDynamicClient(
				scheme,
				&cdiv1beta1.StorageProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name: storageProfileName,
					},
				},
			)
			mockK8SConnector.EXPECT().DynamicClient().Return(dynamicClient)
		})
		It("Should return the StorageProfile when it exists", func() {
			ret, err := GetStorageProfileForStorageClass(mockK8SConnector, storageProfileName)
			Expect(err).To(BeNil())
			Expect(ret.GetName()).To(Equal(storageProfileName))
		})

		It("Should return an error when it does not exist", func() {
			_, err := GetStorageProfileForStorageClass(mockK8SConnector, "foo")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal("storageprofiles.cdi.kubevirt.io \"foo\" not found"))
		})
	})

	Context("GetDataImportCronSourceFormatForStorageClass", func() {
		var (
			storageProfileName = "main-sp"
		)

		It("Should return the dataImportCronSourceFormat", func() {
			expectedSourceFormat := cdiv1beta1.DataImportCronSourceFormatPvc
			dynamicClient := fake.NewSimpleDynamicClient(
				scheme,
				&cdiv1beta1.StorageProfile{
					ObjectMeta: metav1.ObjectMeta{
						Name: storageProfileName,
					},
					Status: cdiv1beta1.StorageProfileStatus{
						DataImportCronSourceFormat: &expectedSourceFormat,
					},
				},
			)
			mockK8SConnector.EXPECT().DynamicClient().Return(dynamicClient)
			sourceFormat, err := GetDataImportCronSourceFormatForStorageClass(mockK8SConnector, storageProfileName)
			Expect(err).To(BeNil())
			Expect(sourceFormat).To(Equal(expectedSourceFormat))
		})

		Context("Error cases", func() {
			BeforeEach(func() {
				dynamicClient := fake.NewSimpleDynamicClient(
					scheme,
					&cdiv1beta1.StorageProfile{
						ObjectMeta: metav1.ObjectMeta{
							Name: storageProfileName,
						},
					},
				)
				mockK8SConnector.EXPECT().DynamicClient().Return(dynamicClient)
			})

			It("Should return error is status does not contain dataImportCronSourceFormat", func() {
				_, err := GetDataImportCronSourceFormatForStorageClass(mockK8SConnector, storageProfileName)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("status field of StorageProfile object does not have a dataImportCronSourceFormat field"))
			})

			It("Should return an error if the StorageProfile does not exist", func(){
				_, err := GetDataImportCronSourceFormatForStorageClass(mockK8SConnector, "foo")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal("storageprofiles.cdi.kubevirt.io \"foo\" not found"))
			})
		})
	})
})
