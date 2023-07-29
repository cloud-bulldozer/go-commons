package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Common params shared across the class
type PodResource struct {
	clientSet *kubernetes.Clientset
	// Pod resource attributes and metadata
}

// Pod specific params
type PodParams struct {
}

func (p *PodResource) Create(ctx context.Context, podParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Create Pod here")
	return nil, nil
}

func (p *PodResource) Update(ctx context.Context, podParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Updating Pod here")
	return nil, nil
}

func (p *PodResource) Get(ctx context.Context, podParams interface{}) (interface{}, error) {
	fmt.Println("Getting Pod here")
	return nil, nil
}

func (p *PodResource) Delete(ctx context.Context, podParams interface{}) error {
	fmt.Println("Deleting Pod here")
	return nil
}
