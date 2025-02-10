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
	"errors"

	"github.com/golang/mock/gomock"
	volumesnapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v4/apis/volumesnapshot/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/kubernetes/fake"
	fakestoragev1 "k8s.io/client-go/kubernetes/typed/storage/v1/fake"
	"k8s.io/client-go/testing"

	"github.com/cloud-bulldozer/go-commons/mocks"
)

var _ = Describe("Tests for K8S Volume Snapshot", func() {
	var (
		mockCtrl                               *gomock.Controller
		mockK8SConnector                       *mocks.MockK8SConnector
		scheme                                 *runtime.Scheme
		expectedVolumeSnapshotClassName        = "expected-vsc"
		expectedVolumeSnapshotClassProvisioner = "expected.provisioner.example.com"
		otherVolumeSnapshotClassName           = "other-vsc"
		otherVolumeSnapshotClassProvisioner    = "other.provisioner.example.com"
		expectedStorageClassName               = "expected-sc"

		getAndRegisterDynamicClient = func() *dynamicfake.FakeDynamicClient {
			dynamicClient := dynamicfake.NewSimpleDynamicClient(
				scheme,
				&volumesnapshotv1.VolumeSnapshotClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: expectedVolumeSnapshotClassName,
					},
					Driver: expectedVolumeSnapshotClassProvisioner,
				},
				&volumesnapshotv1.VolumeSnapshotClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: otherVolumeSnapshotClassName,
					},
					Driver: otherVolumeSnapshotClassProvisioner,
				},
			)
			mockK8SConnector.EXPECT().DynamicClient().Return(dynamicClient)
			return dynamicClient
		}

		getAndRegisterClientSet = func() *fake.Clientset {
			clientSet := fake.NewSimpleClientset(
				&storagev1.StorageClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: expectedStorageClassName,
					},
					Provisioner: expectedVolumeSnapshotClassProvisioner,
				},
			)
			mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			return clientSet
		}
	)

	BeforeEach(func() {
		scheme = runtime.NewScheme()

		// Add VolumeSnapshotClass objects to the scheme
		scheme.AddKnownTypes(
			volumesnapshotv1.SchemeGroupVersion,
			&volumesnapshotv1.VolumeSnapshotClass{},
			&volumesnapshotv1.VolumeSnapshotClassList{},
		)

		mockCtrl = gomock.NewController(GinkgoT())
		mockK8SConnector = mocks.NewMockK8SConnector(mockCtrl)
	})

	Context("getVolumeSnapshotClassNameForProvisioner", func() {
		var (
			dynamicClient *dynamicfake.FakeDynamicClient
		)
		BeforeEach(func() {
			dynamicClient = getAndRegisterDynamicClient()
		})

		It("Should return the expected volume snapshot class", func() {
			name, err := getVolumeSnapshotClassNameForProvisioner(mockK8SConnector, expectedVolumeSnapshotClassProvisioner)
			Expect(err).To(BeNil())
			Expect(name).To(Equal(expectedVolumeSnapshotClassName))
		})

		It("Should return an empty string if not found", func() {
			name, err := getVolumeSnapshotClassNameForProvisioner(mockK8SConnector, "foo")
			Expect(err).To(BeNil())
			Expect(name).To(Equal(""))
		})

		It("Should return an error if list failed", func() {
			errorMessage := "Failed listing volumesnapshotclasses"
			dynamicClient.PrependReactor(
				"list",
				"volumesnapshotclasses",
				func(action testing.Action) (handled bool, ret runtime.Object, err error) {
					return true, nil, errors.New(errorMessage)
				},
			)
			_, err := getVolumeSnapshotClassNameForProvisioner(mockK8SConnector, "foo")
			Expect(err).ToNot(BeNil())
			Expect(err.Error()).To(Equal(errorMessage))
		})
	})

	Context("GetVolumeSnapshotClassNameForStorageClass", func() {
		Context("Success", func(){
			BeforeEach(func() {
				getAndRegisterDynamicClient()
				getAndRegisterClientSet()
			})

			It("Should return the volume snapshot class with the matching provisioner", func() {
				name, err := GetVolumeSnapshotClassNameForStorageClass(mockK8SConnector, expectedStorageClassName)
				Expect(err).To(BeNil())
				Expect(name).To(Equal(expectedVolumeSnapshotClassName))
			})
		})
		Context("Failure", func(){
			var (
				clientSet *fake.Clientset
			)
			BeforeEach(func() {
				clientSet = getAndRegisterClientSet()
			})
			It("Should return an error if failed to get the storage class", func() {
				errorMessage := "Error getting storage classes"
				clientSet.StorageV1().(*fakestoragev1.FakeStorageV1).PrependReactor(
					"get",
					"storageclasses",
					func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New(errorMessage)
					},
				)
				_, err := GetVolumeSnapshotClassNameForStorageClass(mockK8SConnector, "foo")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(errorMessage))
			})
		})
	})
})
