package k8scommons

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GetZone returns a map with all the zones an the number of
// worker nodes on each one
func (c *K8SClient) GetZones(ctx context.Context) (map[string]int, error) {
	return c.getZones(ctx)
}

// IsClusterMultiAZ returns true if the cluster has more than
// one zone configured in their worker nodes
func (c *K8SClient) IsClusterMultiAZ(ctx context.Context) (bool, error) {
	zones, err := c.getZones(ctx)
	if err != nil {
		return false, err
	}
	return len(zones) > 1, nil
}

func (c *K8SClient) getZones(ctx context.Context) (map[string]int, error) {
	zones := make(map[string]int)

	n, err := c.client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{LabelSelector: "node-role.kubernetes.io/worker="})
	if err != nil {
		return zones, err
	}

	for _, l := range n.Items {
		if len(l.GetLabels()["topology.kubernetes.io/zone"]) < 1 {
			return zones, fmt.Errorf("no zone label on Node: %s", l.Name)
		}
		zones[l.GetLabels()["topology.kubernetes.io/zone"]]++
	}
	return zones, nil
}

func (c *K8SClient) GetZoneForNode(ctx context.Context, nodeName string) (string, error) {
	n, err := c.client.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	if len(n.GetLabels()["topology.kubernetes.io/zone"]) < 1 {
		return "", fmt.Errorf("no zone label on Node: %s", nodeName)
	}
	return n.GetLabels()["topology.kubernetes.io/zone"], nil
}
