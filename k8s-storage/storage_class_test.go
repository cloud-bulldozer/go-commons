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
	"fmt"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
	fakestoragev1 "k8s.io/client-go/kubernetes/typed/storage/v1/fake"
	"k8s.io/client-go/testing"
	"k8s.io/utils/ptr"

	"github.com/cloud-bulldozer/go-commons/v2/mocks"
)

var _ = Describe("Tests for K8S Storage Class", func() {
	var (
		mockCtrl         *gomock.Controller
		mockK8SConnector *mocks.MockK8SConnector
	)

	BeforeEach(func() {
		mockCtrl = gomock.NewController(GinkgoT())
		mockK8SConnector = mocks.NewMockK8SConnector(mockCtrl)
	})

	Context("StorageClassExists", func() {

		Context("Success", func() {
			var (
				storageClassName = "test-sc"
			)
			BeforeEach(func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: storageClassName,
						},
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			})
			It("Should return true when storage class exists", func() {
				exists, err := StorageClassExists(mockK8SConnector, storageClassName)
				Expect(err).To(BeNil())
				Expect(exists).To(BeTrue())
			})

			It("Should return false when storage class exists", func() {
				nonExistingStorageClassName := "no-sc"
				exists, err := StorageClassExists(mockK8SConnector, nonExistingStorageClassName)
				Expect(err).To(BeNil())
				Expect(exists).To(BeFalse())
			})
		})

		Context("Failure", func() {
			It("Should return an error when get fails", func() {
				errorMessage := "Error getting storage classes"
				clientSet := fake.NewSimpleClientset()
				clientSet.StorageV1().(*fakestoragev1.FakeStorageV1).PrependReactor(
					"get",
					"storageclasses",
					func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New(errorMessage)
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
				_, err := StorageClassExists(mockK8SConnector, "foo")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(errorMessage))
			})
		})
	})

	Context("GetDefaultStorageClassName", func() {

		Context("Success", func() {
			var (
				defaultStorageClassName     = "default-sc"
				virtDefaultStorageClassName = "default-virt-sc"
				nonDefaultStorageClassName  = "another-sc"
			)
			BeforeEach(func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: defaultStorageClassName,
							Annotations: map[string]string{
								defaultStorageClassAnnotation: "true",
							},
						},
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: virtDefaultStorageClassName,
							Annotations: map[string]string{
								defaultVirtStorageClassAnnotation: "true",
							},
						},
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: nonDefaultStorageClassName,
						},
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			})

			It("Should return the default storage class when virt is not prefered", func() {
				name, err := GetDefaultStorageClassName(mockK8SConnector, false)
				Expect(err).To(BeNil())
				Expect(name).To(Equal(defaultStorageClassName))
			})

			It("Should return the virt default storage class when virt is prefered", func() {
				name, err := GetDefaultStorageClassName(mockK8SConnector, true)
				Expect(err).To(BeNil())
				Expect(name).To(Equal(virtDefaultStorageClassName))
			})
		})

		Context("Failure", func() {
			It("Should return error when no default StorageClass was set", func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: "foo",
						},
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
				_, err := GetDefaultStorageClassName(mockK8SConnector, false)
				Expect(err).To(Equal(fmt.Errorf("no default StorageClass was set")))
			})

			It("Should return an error when list fails", func() {
				errorMessage := "Error listing storage classes"
				clientSet := fake.NewSimpleClientset()
				clientSet.StorageV1().(*fakestoragev1.FakeStorageV1).PrependReactor(
					"list",
					"storageclasses",
					func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New(errorMessage)
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
				_, err := GetDefaultStorageClassName(mockK8SConnector, false)
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(errorMessage))
			})
		})
	})

	Context("StorageClassSupportsVolumeExpansion", func() {

		Context("Success", func() {
			var (
				scSupportsVolumeExpansionTrue  = "yes"
				scSupportsVolumeExpansionFalse = "no"
				scSupportsVolumeExpansionNone  = "none"
			)
			BeforeEach(func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: scSupportsVolumeExpansionTrue,
						},
						AllowVolumeExpansion: ptr.To(true),
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: scSupportsVolumeExpansionFalse,
						},
						AllowVolumeExpansion: ptr.To(false),
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: scSupportsVolumeExpansionNone,
						},
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			})

			It("Should return true when the StorageClass supports VolumeExpansion", func() {
				supported, err := StorageClassSupportsVolumeExpansion(mockK8SConnector, scSupportsVolumeExpansionTrue)
				Expect(err).To(BeNil())
				Expect(supported).To(BeTrue())
			})

			It("Should return false when the StorageClass does not support VolumeExpansion", func() {
				supported, err := StorageClassSupportsVolumeExpansion(mockK8SConnector, scSupportsVolumeExpansionFalse)
				Expect(err).To(BeNil())
				Expect(supported).To(BeFalse())
			})

			It("Should return false when the VolumeExpansion is not set", func() {
				supported, err := StorageClassSupportsVolumeExpansion(mockK8SConnector, scSupportsVolumeExpansionNone)
				Expect(err).To(BeNil())
				Expect(supported).To(BeFalse())
			})
		})

		Context("Failure", func() {
			It("Should return an error when get fails", func() {
				errorMessage := "Error getting storage classes"
				clientSet := fake.NewSimpleClientset()
				clientSet.StorageV1().(*fakestoragev1.FakeStorageV1).PrependReactor(
					"get",
					"storageclasses",
					func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New(errorMessage)
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
				_, err := StorageClassSupportsVolumeExpansion(mockK8SConnector, "foo")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(errorMessage))
			})
		})
	})

	Context("getStorageClassProvisioner", func() {

		Context("Success", func() {
			var (
				scName          = "foo"
				provisionerName = "foo.example.com"
			)
			BeforeEach(func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: scName,
						},
						Provisioner: provisionerName,
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			})

			It("Should return the name of the provisioner", func() {
				name, err := getStorageClassProvisioner(mockK8SConnector, scName)
				Expect(err).To(BeNil())
				Expect(name).To(Equal(provisionerName))
			})
		})

		Context("Failure", func() {
			It("Should return an error when get fails", func() {
				errorMessage := "Error getting storage classes"
				clientSet := fake.NewSimpleClientset()
				clientSet.StorageV1().(*fakestoragev1.FakeStorageV1).PrependReactor(
					"get",
					"storageclasses",
					func(action testing.Action) (handled bool, ret runtime.Object, err error) {
						return true, nil, errors.New(errorMessage)
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
				_, err := getStorageClassProvisioner(mockK8SConnector, "foo")
				Expect(err).ToNot(BeNil())
				Expect(err.Error()).To(Equal(errorMessage))
			})
		})
	})

	Context("GetStorageClassName", func() {

		Context("Success", func() {
			var (
				defaultStorageClassName     = "default-sc"
				virtDefaultStorageClassName = "default-virt-sc"
				nonDefaultStorageClassName  = "another-sc"
			)
			BeforeEach(func() {
				clientSet := fake.NewSimpleClientset(
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: defaultStorageClassName,
							Annotations: map[string]string{
								defaultStorageClassAnnotation: "true",
							},
						},
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: virtDefaultStorageClassName,
							Annotations: map[string]string{
								defaultVirtStorageClassAnnotation: "true",
							},
						},
					},
					&storagev1.StorageClass{
						ObjectMeta: metav1.ObjectMeta{
							Name: nonDefaultStorageClassName,
						},
					},
				)
				mockK8SConnector.EXPECT().ClientSet().Return(clientSet)
			})

			Context("No StorageClassName was provided", func() {
				It("Should return the default storage class when virt is not prefered", func() {
					name, err := GetStorageClassName(mockK8SConnector, "", false)
					Expect(err).To(BeNil())
					Expect(name).To(Equal(defaultStorageClassName))
				})

				It("Should return the virt default storage class when virt is prefered", func() {
					name, err := GetStorageClassName(mockK8SConnector, "", true)
					Expect(err).To(BeNil())
					Expect(name).To(Equal(virtDefaultStorageClassName))
				})
			})

			Context("With StorageClassName", func() {
				It("Should return the same name when the StorageClass exists", func() {
					name, err := GetStorageClassName(mockK8SConnector, nonDefaultStorageClassName, false)
					Expect(err).To(BeNil())
					Expect(name).To(Equal(nonDefaultStorageClassName))
				})

				It("Should return an empty string if the StorageClass does not exist", func() {
					name, err := GetStorageClassName(mockK8SConnector, "foo", false)
					Expect(err).To(BeNil())
					Expect(name).To(Equal(""))
				})
			})
		})
	})
})
