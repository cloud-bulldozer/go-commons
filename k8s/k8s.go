package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
)

// Define the Resource Interface for CRUD operations
type Resource interface {
	Create(context.Context, interface{}, bool) (interface{}, error)
	Update(context.Context, interface{}, bool) (interface{}, error)
	Get(context.Context, interface{}) (interface{}, error)
	Delete(context.Context, interface{}) error
	// Other resource-specific methods...
}

// Implement a single repository struct containing all resource-specific structs
type KubernetesRepository struct {
	Pod        PodResource
	Deployment DeploymentResource
	ReplicaSet ReplicaSetResource
	DeamonSet  DeamonSetResource
	Job        JobResource
	clientSet  *kubernetes.Clientset
	// Other resource-specific structs...
}

func NewKubernetesRepository() (*KubernetesRepository, error) {
	clientSet := getClientSet()
	if clientSet == nil {
		return nil, fmt.Errorf("Error while connecting to the cluster")
	}
	return &KubernetesRepository{
		Pod:        PodResource{clientSet: clientSet},
		Deployment: DeploymentResource{clientSet: clientSet},
		ReplicaSet: ReplicaSetResource{clientSet: clientSet},
		DeamonSet:  DeamonSetResource{clientSet: clientSet},
		Job:        JobResource{clientSet: clientSet},
		clientSet:  clientSet,
	}, nil
}
