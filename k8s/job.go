package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Common params shared across the class
type JobResource struct {
	clientSet *kubernetes.Clientset
	// Job resource attributes and metadata
}

// Job specific params here
type JobParams struct {
}

func (p *JobResource) Create(ctx context.Context, JobParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Create Job here")
	return nil, nil
}

func (p *JobResource) Update(ctx context.Context, JobParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Updating Job here")
	return nil, nil
}

func (p *JobResource) Get(ctx context.Context, JobParams interface{}) (interface{}, error) {
	fmt.Println("Getting Job here")
	return nil, nil
}

func (p *JobResource) Delete(ctx context.Context, JobParams interface{}) error {
	fmt.Println("Deleting Job here")
	return nil
}
