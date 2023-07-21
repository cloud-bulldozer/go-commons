package k8scommons

import (
	"context"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KClient interface {
	GetVersion() string
	GetPod(ctx context.Context, namespace string, name string, version string) (*apiv1.Pod, error)
	GetPods(ctx context.Context, namespace string, podParams PodParams) (*apiv1.PodList, error)
}

func NewKClient(configOverrides clientcmd.ConfigOverrides) (KClient, error) {
	kconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		clientcmd.NewDefaultClientConfigLoadingRules(),
		&configOverrides)
	rconfig, err := kconfig.ClientConfig()
	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(rconfig)
	if err != nil {
		return nil, err
	}
	return &K8SClient{
		client: client,
	}, nil
}
