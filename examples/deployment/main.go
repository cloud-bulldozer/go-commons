package main

import (
	"context"
	"fmt"

	"github.com/cloud-bulldozer/go-commons/k8s"
	log "github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
)

func main() {
	// Usage on common Kubernetes Repository
	kubeRepo, err := k8s.NewKubernetesRepository()
	if err != nil {
		log.Info("Some error occured")
	}
	log.Infof("kube Repo object : %+v", kubeRepo)

	var res k8s.Resource = &kubeRepo.Deployment
	if _, ok := res.(*k8s.DeploymentResource); ok {
		fmt.Printf("DeploymentResource implements Resource interface. %+v", res)
	} else {
		fmt.Printf("DeploymentResource does not implement Resource interface. %+v", res)
	}
	deploymentParams := k8s.DeploymentParams{
		Name:               "testdeployment",
		Namespace:          "default",
		Replicas:           2,
		SelectorLabels:     "app=test",
		MetadataLabels:     "app=test",
		NodeSelectorLabels: "app=test",
		Containers: []v1.Container{
			{
				Name:            "sleep",
				Image:           "gcr.io/google_containers/pause-amd64:3.0",
				ImagePullPolicy: v1.PullAlways,
			},
		},
	}
	deployParams, err := kubeRepo.Deployment.Create(context.Background(), deploymentParams, false)
	if err != nil {
		fmt.Println("Error creating Deployment:", err)
	}
	log.Infof("Created Deployment: %+v", deployParams.(k8s.DeploymentParams).Name)

	deploymentParams = k8s.DeploymentParams{
		Name:           "testdeployment",
		Namespace:      "default",
		SelectorLabels: "app=test",
		MetadataLabels: "app=test",
		Replicas:       4,
		Containers: []v1.Container{
			{
				Name:            "sleep",
				Image:           "gcr.io/google_containers/pause-amd64:3.0",
				ImagePullPolicy: v1.PullAlways,
			},
		},
	}
	deployParams, err = kubeRepo.Deployment.Update(context.Background(), deploymentParams, false)
	if err != nil {
		fmt.Println("Error Updating Deployment:", err)
	}
	log.Infof("Updated Deployment: %+v", deployParams.(k8s.DeploymentParams).Name)

	deploymentParams = k8s.DeploymentParams{
		Deployment: deployParams.(k8s.DeploymentParams).Deployment,
	}

	deployment, error := kubeRepo.Deployment.Get(context.Background(), deploymentParams)
	if error != nil {
		fmt.Println("Error Getting Deployment:", error)
	}
	log.Infof("Got Deployment: %+v", deployment.(*appsv1.Deployment))

	deploymentParams = k8s.DeploymentParams{
		Name:      "testdeployment",
		Namespace: "default",
	}
	err = kubeRepo.Deployment.Delete(context.Background(), deploymentParams)
	if err != nil {
		fmt.Println("Error Deleting Deployment:", err)
	}
	log.Infof("Deleted Deployment: %+v", deploymentParams.Name)

	// Usage of individual component explicitly
	deploymentResource := &k8s.DeploymentResource{}
	deployment_dup := k8s.DeploymentParams{
		Name:               "testdeployment",
		Namespace:          "default",
		Replicas:           2,
		SelectorLabels:     "app=test",
		MetadataLabels:     "app=test",
		NodeSelectorLabels: "app=test",
		Containers: []v1.Container{
			{
				Name:            "sleep",
				Image:           "gcr.io/google_containers/pause-amd64:3.0",
				ImagePullPolicy: v1.PullAlways,
			},
		},
	}
	deployParams, err = deploymentResource.Create(context.Background(), deployment_dup, false)
	if err != nil {
		fmt.Println("Error creating Deployment:", err)
	}
	log.Infof("Created Deployment: %+v", deployParams.(k8s.DeploymentParams).Name)

	deploymentParams = k8s.DeploymentParams{
		Name:      "testdeployment",
		Namespace: "default",
	}
	err = deploymentResource.Delete(context.Background(), deploymentParams)
	if err != nil {
		fmt.Println("Error Deleting Deployment:", err)
	}
	log.Infof("Deleted Deployment: %+v", deploymentParams.Name)
}
