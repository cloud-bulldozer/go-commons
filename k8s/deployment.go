package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Common params shared across the class
type DeploymentResource struct {
	clientSet *kubernetes.Clientset
	// Deployment resource attributes and metadata
}

// Deployment specific params
type DeploymentParams struct {
	// Name of the deployment
	Name string
	// Namespace of the deployment
	Namespace string
	// Replicas of the deployment
	Replicas int32
	// Selector labels deployment.Spec.Selector
	// semicolon separated string with labels
	// Example: "app1=label1;app2=label2"
	SelectorLabels string
	// Metadata labels deployment.Spec.Template.ObjectSpec.Labels
	// semicolon separated string with labels
	// Example: "app1=label1;app2=label2"
	MetadataLabels string
	// NodeAffinity rules to go in RequiredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	NodeAffinityPreferred string
	// NodeAffinity rules to go in PreferredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	NodeAffinityRequired string
	// PodAffinity rules to go in RequiredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	PodAffinityPreferred string
	// PodAffinity rules to go in PreferredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	PodAffinityRequired string
	// PodAntiAffinityPreferred rules to go in RequiredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	PodAntiAffinityPreferred string
	// PodAntiAffinityRequired rules to go in PreferredDuringSchedulingIgnoredDuringExecution section
	// semicolon separated string with labels
	// Example : `app1=["a","b","c"];app2!=["d","e","f"];app3=;app4!=`
	// in the above example labels fall under `OpIn;OpOut;Exists;DoesNotExists`
	// cases respectively based on their equality signs. Parsing is done automatically.
	PodAntiAffinityRequired string
	// NodeSelector labels deployment.Spec.Template.Spec.NodeSelector
	// semicolon separated string with labels
	// Example: "app1=label1;app2=label2"
	NodeSelectorLabels string
	// Flag to enable hostnetwork
	HostNetwork bool
	// Service account name
	ServiceAccountName string
	// List of container specs
	Containers []v1.Container
	// Deployment object if you already have on in place.
	Deployment *appsv1.Deployment
}

// Method to create a deployment.
func (d *DeploymentResource) Create(ctx context.Context, deployParams interface{}, dryRun bool) (interface{}, error) {
	if d.clientSet == nil {
		d.clientSet = getClientSet()
		if d.clientSet == nil {
			return nil, fmt.Errorf("Error while connecting to the cluster")
		}
	}
	params := deployParams.(DeploymentParams)
	if params.Name != "" && params.Deployment != nil {
		return nil, fmt.Errorf("Invalid params. deployment.Name and deployment.Deployment are mutually exclusive")
	}
	if params.Name != "" {
		deployment := setupDeployment(deployParams.(DeploymentParams))
		params.Deployment = deployment
	}
	if !dryRun {
		// Create the deployment
		_, err := d.clientSet.AppsV1().Deployments("default").Create(ctx, params.Deployment, metav1.CreateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Error creating Deployment: %v", err)
		}
	}
	return params, nil
}

// Method to update a deployment.
func (d *DeploymentResource) Update(ctx context.Context, deployParams interface{}, dryRun bool) (interface{}, error) {
	if d.clientSet == nil {
		d.clientSet = getClientSet()
		if d.clientSet == nil {
			return nil, fmt.Errorf("Error while connecting to the cluster")
		}
	}
	params := deployParams.(DeploymentParams)
	if params.Name != "" && params.Deployment != nil {
		return nil, fmt.Errorf("Invalid params. deployment.Name and deployment.Deployment are mutually exclusive")
	}
	var deploymentToUpdate *appsv1.Deployment
	if params.Name != "" {
		currentDeployment, err := d.Get(ctx, params)
		if err != nil {
			return nil, fmt.Errorf("Error invalid operation. Deployment not found to update")
		}
		deployment := currentDeployment.(*appsv1.Deployment)
		*deployment.Spec.Replicas = params.Replicas
		nodeAffinity, podAffinity, podAntiAffinity := getAffinityRules(params)
		deployment.Spec.Template.Spec.Affinity = &v1.Affinity{
			NodeAffinity:    nodeAffinity,
			PodAffinity:     podAffinity,
			PodAntiAffinity: podAntiAffinity,
		}
		deployment.Spec.Template.Spec.Containers = params.Containers
		deployment.Spec.Template.Spec.NodeSelector = getLabels(params.NodeSelectorLabels)
		deployment.Spec.Template.Spec.HostNetwork = params.HostNetwork
		deployment.Spec.Template.Spec.ServiceAccountName = params.ServiceAccountName
		deploymentToUpdate = deployment
	} else {
		deploymentToUpdate = params.Deployment
	}
	params.Deployment = deploymentToUpdate
	if !dryRun {
		// Patch the deployment
		_, err := d.clientSet.AppsV1().Deployments(params.Namespace).Update(ctx, deploymentToUpdate, metav1.UpdateOptions{})
		if err != nil {
			return nil, fmt.Errorf("Error patching deployment: %s\n", err.Error())
		}
	}
	return params, nil
}

// Method to get a deployment
func (d *DeploymentResource) Get(ctx context.Context, deployParams interface{}) (interface{}, error) {
	if d.clientSet == nil {
		d.clientSet = getClientSet()
		if d.clientSet == nil {
			return nil, fmt.Errorf("Error while connecting to the cluster")
		}
	}
	params := deployParams.(DeploymentParams)
	if params.Name != "" && params.Deployment != nil {
		return nil, fmt.Errorf("Invalid params. deployment.Name and deployment.Deployment are mutually exclusive")
	}
	if params.Name == "" {
		params.Name = params.Deployment.ObjectMeta.Name
		params.Namespace = params.Deployment.ObjectMeta.Namespace
	}
	// Get the deployment object
	deployment, err := d.clientSet.AppsV1().Deployments(params.Namespace).Get(ctx, params.Name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("Error getting deployment: %s\n", err.Error())
	}
	return deployment, nil
}

// Method to delete a deployment
func (d *DeploymentResource) Delete(ctx context.Context, deployParams interface{}) error {
	if d.clientSet == nil {
		d.clientSet = getClientSet()
		if d.clientSet == nil {
			return fmt.Errorf("Error while connecting to the cluster")
		}
	}
	params := deployParams.(DeploymentParams)
	if params.Name != "" && params.Deployment != nil {
		return fmt.Errorf("Invalid params. deployment.Name and deployment.Deployment are mutually exclusive")
	}
	if params.Name == "" {
		params.Name = params.Deployment.ObjectMeta.Name
		params.Namespace = params.Deployment.ObjectMeta.Namespace
	}
	// Delete the deployment
	err := d.clientSet.AppsV1().Deployments(params.Namespace).Delete(ctx, params.Name, metav1.DeleteOptions{})
	if err != nil {
		return fmt.Errorf("Error getting deployment: %s\n", err.Error())
	}
	return nil
}

// Private helper method to handle operations with a deployment.

func getNodeAffinityRequired(nodeAffinityString string) *v1.NodeAffinity {
	return &v1.NodeAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: generateNodeSelectorRequirement(parseAffinityRules(nodeAffinityString)),
				},
			},
		},
	}
}

func getPodAffinityRequired(podAffinityString string) *v1.PodAffinity {
	return &v1.PodAffinity{
		RequiredDuringSchedulingIgnoredDuringExecution: []v1.PodAffinityTerm{
			{
				LabelSelector: &metav1.LabelSelector{
					MatchExpressions: generateLabelSelectorRequirement(parseAffinityRules(podAffinityString)),
				},
				TopologyKey: "kubernetes.io/hostname",
			},
		},
	}
}

func getNodeAffinityPreferred(nodeAffinityString string) *v1.NodeAffinity {
	return &v1.NodeAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.PreferredSchedulingTerm{
			{
				Weight: 100,
				Preference: v1.NodeSelectorTerm{
					MatchExpressions: generateNodeSelectorRequirement(parseAffinityRules(nodeAffinityString)),
				},
			},
		},
	}
}

func getPodAffinityPreferred(podAffinityString string) *v1.PodAffinity {
	return &v1.PodAffinity{
		PreferredDuringSchedulingIgnoredDuringExecution: []v1.WeightedPodAffinityTerm{
			{
				Weight: 100,
				PodAffinityTerm: v1.PodAffinityTerm{
					LabelSelector: &metav1.LabelSelector{
						MatchExpressions: generateLabelSelectorRequirement(parseAffinityRules(podAffinityString)),
					},
					TopologyKey: "kubernetes.io/hostname",
				},
			},
		},
	}
}

func getAffinityRules(deployParams DeploymentParams) (*v1.NodeAffinity, *v1.PodAffinity, *v1.PodAntiAffinity) {
	NodeAffinity := &v1.NodeAffinity{}
	PodAffinity := &v1.PodAffinity{}
	PodAntiAffinity := &v1.PodAntiAffinity{}

	if deployParams.NodeAffinityPreferred != "" && deployParams.NodeAffinityRequired != "" {
		NodeAffinity = &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  getNodeAffinityRequired(deployParams.NodeAffinityRequired).RequiredDuringSchedulingIgnoredDuringExecution,
			PreferredDuringSchedulingIgnoredDuringExecution: getNodeAffinityPreferred(deployParams.NodeAffinityPreferred).PreferredDuringSchedulingIgnoredDuringExecution,
		}
	} else {
		if deployParams.NodeAffinityRequired != "" {
			NodeAffinity = getNodeAffinityRequired(deployParams.NodeAffinityRequired)
		} else if deployParams.NodeAffinityPreferred != "" {
			NodeAffinity = getNodeAffinityPreferred(deployParams.NodeAffinityPreferred)
		}
	}

	if deployParams.PodAffinityRequired != "" && deployParams.PodAffinityPreferred != "" {
		PodAffinity = &v1.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  getPodAffinityRequired(deployParams.PodAffinityRequired).RequiredDuringSchedulingIgnoredDuringExecution,
			PreferredDuringSchedulingIgnoredDuringExecution: getPodAffinityPreferred(deployParams.PodAffinityPreferred).PreferredDuringSchedulingIgnoredDuringExecution,
		}
	} else {
		if deployParams.PodAffinityRequired != "" {
			PodAffinity = getPodAffinityRequired(deployParams.PodAffinityRequired)
		} else if deployParams.PodAffinityPreferred != "" {
			PodAffinity = getPodAffinityPreferred(deployParams.PodAffinityPreferred)
		}

	}

	if deployParams.PodAntiAffinityRequired != "" && deployParams.PodAntiAffinityPreferred != "" {
		PodAntiAffinity = &v1.PodAntiAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution:  getPodAffinityRequired(deployParams.PodAntiAffinityRequired).RequiredDuringSchedulingIgnoredDuringExecution,
			PreferredDuringSchedulingIgnoredDuringExecution: getPodAffinityPreferred(deployParams.PodAntiAffinityPreferred).PreferredDuringSchedulingIgnoredDuringExecution,
		}
	} else {
		if deployParams.PodAntiAffinityRequired != "" {
			PodAntiAffinity = &v1.PodAntiAffinity{RequiredDuringSchedulingIgnoredDuringExecution: getPodAffinityRequired(deployParams.PodAntiAffinityRequired).RequiredDuringSchedulingIgnoredDuringExecution}
		} else if deployParams.PodAntiAffinityPreferred != "" {
			PodAntiAffinity = &v1.PodAntiAffinity{PreferredDuringSchedulingIgnoredDuringExecution: getPodAffinityPreferred(deployParams.PodAntiAffinityRequired).PreferredDuringSchedulingIgnoredDuringExecution}
		}
	}
	return NodeAffinity, PodAffinity, PodAntiAffinity
}

func setupDeployment(deployParams DeploymentParams) *appsv1.Deployment {

	NodeAffinity, PodAffinity, PodAntiAffinity := getAffinityRules(deployParams)
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deployParams.Name,
			Namespace: deployParams.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &deployParams.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: getLabels(deployParams.SelectorLabels),
			},
			Template: v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: getLabels(deployParams.MetadataLabels),
				},
				Spec: v1.PodSpec{
					ServiceAccountName: deployParams.ServiceAccountName,
					HostNetwork:        deployParams.HostNetwork,
					Containers:         deployParams.Containers,
					Affinity: &v1.Affinity{
						NodeAffinity:    NodeAffinity,
						PodAffinity:     PodAffinity,
						PodAntiAffinity: PodAntiAffinity,
					},
					NodeSelector: getLabels(deployParams.NodeSelectorLabels),
				},
			},
		},
	}
}
