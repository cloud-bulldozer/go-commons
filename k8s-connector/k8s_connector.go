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

package k8sconnector

import (
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type K8SConnector interface {
	ClientSet() kubernetes.Interface
	RestConfig() *rest.Config
	DynamicClient() dynamic.Interface
}

// K8SConnectorImpl object
type K8SConnectorImpl struct {
	clientSet     kubernetes.Interface
	restConfig    *rest.Config
	dynamicClient dynamic.Interface
}

// NewMetadata instantiates a new OCP metadata discovery agent
func NewK8SConnector(restConfig *rest.Config) (K8SConnector, error) {
	cs, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	dc, err := dynamic.NewForConfig(restConfig)
	return &K8SConnectorImpl{
		clientSet:     cs,
		dynamicClient: dc,
		restConfig:    restConfig,
	}, err
}

func (c *K8SConnectorImpl) ClientSet() kubernetes.Interface {
	return c.clientSet
}

func (c *K8SConnectorImpl) RestConfig() *rest.Config {
	return c.restConfig
}

func (c *K8SConnectorImpl) DynamicClient() dynamic.Interface {
	return c.dynamicClient
}
