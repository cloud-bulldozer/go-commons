package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Implement individual Go structs for each Kubernetes resource
type ReplicaSetResource struct {
	clientSet *kubernetes.Clientset
	// ReplicaSet resource attributes and metadata
}

// ReplicaSet specific params
type ReplicaSetParams struct {
}

func (p *ReplicaSetResource) Create(ctx context.Context, ReplicaSetParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Create ReplicaSet here")
	return nil, nil
}

func (p *ReplicaSetResource) Update(ctx context.Context, ReplicaSetParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Updating ReplicaSet here")
	return nil, nil
}

func (p *ReplicaSetResource) Get(ctx context.Context, ReplicaSetParams interface{}) (interface{}, error) {
	fmt.Println("Getting ReplicaSet here")
	return nil, nil
}

func (p *ReplicaSetResource) Delete(ctx context.Context, ReplicaSetParams interface{}) error {
	fmt.Println("Deleting ReplicaSet here")
	return nil
}
