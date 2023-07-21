package k8scommons

import (
	"k8s.io/client-go/kubernetes"
)

type K8SClient struct {
	client *kubernetes.Clientset
}

type PodParams struct {
	Name            string
	Deployment      string
	All             bool // true if you want non running pods to be returned
	LabelSelector   string
	FieldSelector   string
	ResourceVersion string
}

func (c *K8SClient) GetVersion() string {
	return c.client.RESTClient().APIVersion().Version
}