package k8scommons

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (c *K8SClient) GetPod(ctx context.Context, namespace string, name string, version string) (*apiv1.Pod, error) {
	pod, err := c.getPods(ctx, namespace, PodParams{Name: name, ResourceVersion: version})
	if err != nil {
		return nil, err
	}
	return pod.(*apiv1.Pod), nil
}

func (c *K8SClient) GetPods(ctx context.Context, namespace string, podParams PodParams) (*apiv1.PodList, error) {
	pods, err := c.getPods(ctx, namespace, podParams)
	if err != nil {
		return nil, err
	}
	return pods.(*apiv1.PodList), nil
}

func (c *K8SClient) getPods(ctx context.Context, namespace string, podParams PodParams) (interface{}, error) {
	fieldSelector := "status.phase=Running"
	labelSelector := podParams.LabelSelector
	if podParams.Name != "" {
		return c.client.CoreV1().Pods(namespace).Get(ctx, podParams.Name, v1.GetOptions{ResourceVersion: podParams.ResourceVersion})
	}
	if podParams.All {
		fieldSelector = ""
	}
	if podParams.FieldSelector != "" {
		fieldSelector = fmt.Sprintf("%s,%s", fieldSelector, podParams.FieldSelector)
	}
	if podParams.Deployment != "" {
		d, err := c.client.AppsV1().Deployments(namespace).Get(ctx, podParams.Deployment, v1.GetOptions{ResourceVersion: podParams.ResourceVersion})
		if err != nil {
			return nil, err
		}
		selector, err := metav1.LabelSelectorAsSelector(d.Spec.Selector)
		if err != nil {
			return nil, err
		}
		labelSelector = fmt.Sprintf("%s,%s", labelSelector, selector.String())
	}

	return c.client.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{LabelSelector: labelSelector, FieldSelector: fieldSelector, ResourceVersion: podParams.ResourceVersion})
}
