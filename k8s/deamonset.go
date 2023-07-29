package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Common params shared across the class
type DeamonSetResource struct {
	clientSet *kubernetes.Clientset
	// DeamonSet resource attributes and metadata
}

// DeamonSet specific params
type DeamonSetParams struct {
}

func (p *DeamonSetResource) Create(ctx context.Context, DeamonSetParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Create DeamonSet here")
	return nil, nil
}

func (p *DeamonSetResource) Update(ctx context.Context, DeamonSetParams interface{}, dryRun bool) (interface{}, error) {
	fmt.Println("Updating DeamonSet here")
	return nil, nil
}

func (p *DeamonSetResource) Get(ctx context.Context, DeamonSetParams interface{}) (interface{}, error) {
	fmt.Println("Getting DeamonSet here")
	return nil, nil
}

func (p *DeamonSetResource) Delete(ctx context.Context, DeamonSetParams interface{}) error {
	fmt.Println("Deleting DeamonSet here")
	return nil
}
