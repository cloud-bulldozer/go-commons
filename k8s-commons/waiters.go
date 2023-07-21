// Shamelessly based on the waiters of kube-burner

// Copyright 2020 The Kube-burner Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package k8scommons

import (
	"context"
	"fmt"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
)

func (c *K8SClient) WaitForObject(ctx context.Context, namespace string, resourceType ResourceType, name string, timeout time.Duration) (bool, error) {
	switch resourceType {
	case Deployment:
		return c.waitForDeployments(ctx, namespace, name, timeout)
	case ReplicaSet:
		return c.waitForRS(ctx, namespace, name, timeout)
	case ReplicationController:
		return c.waitForRC(ctx, namespace, name, timeout)
	case StatefulSet:
		return c.waitForStatefulSet(ctx, namespace, name, timeout)
	case DaemonSet:
		return c.waitForDS(ctx, namespace, name, timeout)
	case Pod:
		return c.waitForPod(ctx, namespace, name, timeout)
	// TODO: Don't know how to impllement this one need to understand more the original code
	// case Build, BuildConfig:
	// 	c.waitForBuild(ctx, namespace, name, timeout, obj.Replicas)

	// TODO: Finish converting this methods
	// case VirtualMachine:
	// 	c.waitForVM(ctx, namespace, name, timeout)
	// case VirtualMachineInstance:
	// 	c.waitForVMI(ctx, namespace, name, timeout)
	// case VirtualMachineInstanceReplicaSet:
	// 	c.waitForVMIRS(ctx, namespace, name, timeout)
	// case Job:
	// 	c.waitForJob(ctx, namespace, name, timeout)
	case PersistentVolumeClaim:
		return c.waitForPVC(ctx, namespace, name, timeout)
	default:
		return false, fmt.Errorf("invalid resource type: %v", resourceType)
	}
}

func (c *K8SClient) waitForDeployments(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	// TODO handle errors such as timeouts
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		dep, err := c.client.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if *dep.Spec.Replicas != dep.Status.ReadyReplicas {
			return false, nil
		}
		return true, nil
	})
	return false, nil
}

func (c *K8SClient) waitForRS(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		rs, err := c.client.AppsV1().ReplicaSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if *rs.Spec.Replicas != rs.Status.ReadyReplicas {
			return false, nil
		}
		return true, nil
	})
	return false, nil
}

func (c *K8SClient) waitForStatefulSet(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		sts, err := c.client.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if *sts.Spec.Replicas != sts.Status.ReadyReplicas {
			return false, nil
		}
		return true, nil
	})
	return false, nil
}

func (c *K8SClient) waitForPVC(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		pvc, err := c.client.CoreV1().PersistentVolumeClaims(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pvc.Status.Phase == v1.ClaimBound {
			return true, nil
		}
		return false, nil
	})
	return false, nil
}

func (c *K8SClient) waitForRC(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		rc, err := c.client.CoreV1().ReplicationControllers(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if *rc.Spec.Replicas != rc.Status.ReadyReplicas {
			return false, nil
		}
		return true, nil
	})
	return false, nil
}

func (c *K8SClient) waitForDS(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {
	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		ds, err := c.client.AppsV1().DaemonSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}

		if ds.Status.DesiredNumberScheduled != ds.Status.NumberReady {
			return false, nil
		}

		return true, nil
	})
	return false, nil
}

func (c *K8SClient) waitForPod(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {

	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
		pod, err := c.client.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return false, err
		}
		if pod.Status.Phase == v1.PodRunning {
			return true, nil
		}
		return false, nil
	})
	return false, nil
}

// func (c *K8SClient) waitForBuild(ctx context.Context, ns string, maxWaitTimeout time.Duration, expected int) {
//
// 	buildStatus := []string{"New", "Pending", "Running"}
// 	var build types.UnstructuredContent
// 	gvr := schema.GroupVersionResource{
// 		Group:    types.OpenShiftBuildGroup,
// 		Version:  types.OpenShiftBuildAPIVersion,
// 		Resource: types.OpenShiftBuildResource,
// 	}
// 	wait.PollUntilContextTimeout(context.TODO(), time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
// 		builds, err := waitDynamicClient.Resource(gvr).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
// 		if err != nil {
// 			return false, err
// 		}
// 		if len(builds.Items) < expected {
// 			log.Debugf("Waiting for Builds in ns %s to be completed", ns)
// 			return false, err
// 		}
// 		for _, b := range builds.Items {
// 			jsonBuild, err := b.MarshalJSON()
// 			if err != nil {
// 				log.Errorf("Error decoding Build object: %s", err)
// 			}
// 			_ = json.Unmarshal(jsonBuild, &build)
// 			for _, bs := range buildStatus {
// 				if build.Status.Phase == "" || build.Status.Phase == bs {
// 					log.Debugf("Waiting for Builds in ns %s to be completed", ns)
// 					return false, err
// 				}
// 			}
// 		}
// 		return true, nil
// 	})
// }

// func (c *K8SClient) waitForJob(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {

// 	gvr := schema.GroupVersionResource{
// 		Group:    "batch",
// 		Version:  "v1",
// 		Resource: "jobs",
// 	}
// 	verifyCondition(gvr, ns, "Complete", maxWaitTimeout)
// 	return false, nil
// }

// func (c *K8SClient) waitForCondition(gvr schema.GroupVersionResource, ns, condition string, maxWaitTimeout time.Duration,
// 	wg *sync.WaitGroup) {

// 	verifyCondition(gvr, ns, condition, maxWaitTimeout)
// }

// func verifyCondition(gvr schema.GroupVersionResource, ns, condition string, maxWaitTimeout time.Duration) {
// 	var uObj types.UnstructuredContent
// 	wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
// 		var objs *unstructured.UnstructuredList
// 		if ns != "" {
// 			objs, err = waitDynamicClient.Resource(gvr).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
// 		} else {
// 			objs, err = waitDynamicClient.Resource(gvr).List(context.TODO(), metav1.ListOptions{})
// 		}
// 		if err != nil {
// 			return false, err
// 		}
// 	VERIFY:
// 		for _, obj := range objs.Items {
// 			jsonBuild, err := obj.MarshalJSON()
// 			if err != nil {
// 				log.Errorf("Error decoding object: %s", err)
// 				return false, err
// 			}
// 			_ = json.Unmarshal(jsonBuild, &uObj)
// 			for _, c := range uObj.Status.Conditions {
// 				if c.Status == "True" && c.Type == condition {
// 					continue VERIFY
// 				}
// 			}
// 			if ns != "" {
// 				log.Debugf("Waiting for %s in ns %s to be ready", gvr.Resource, ns)
// 			} else {
// 				log.Debugf("Waiting for %s to be ready", gvr.Resource)
// 			}
// 			return false, err
// 		}
// 		return true, nil
// 	})
// }

// func (c *K8SClient) waitForVM(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {

// 	vmGVR := schema.GroupVersionResource{
// 		Group:    types.KubevirtGroup,
// 		Version:  types.KubevirtAPIVersion,
// 		Resource: types.VirtualMachineResource,
// 	}
// 	verifyCondition(vmGVR, ns, "Ready", maxWaitTimeout)
// 	return false, nil
// }

// func (c *K8SClient) waitForVMI(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {

// 	vmiGVR := schema.GroupVersionResource{
// 		Group:    types.KubevirtGroup,
// 		Version:  types.KubevirtAPIVersion,
// 		Resource: types.VirtualMachineInstanceResource,
// 	}
// 	verifyCondition(vmiGVR, ns, "Ready", maxWaitTimeout)
// 	return false, nil
// }

// func (c *K8SClient) waitForVMIRS(ctx context.Context, ns string, name string, maxWaitTimeout time.Duration) (bool, error) {

// 	var rs types.UnstructuredContent
// 	vmiGVRRS := schema.GroupVersionResource{
// 		Group:    types.KubevirtGroup,
// 		Version:  types.KubevirtAPIVersion,
// 		Resource: types.VirtualMachineInstanceReplicaSetResource,
// 	}
// 	wait.PollUntilContextTimeout(context.TODO(), 10*time.Second, maxWaitTimeout, true, func(ctx context.Context) (done bool, err error) {
// 		objs, err := waitDynamicClient.Resource(vmiGVRRS).Namespace(ns).List(context.TODO(), metav1.ListOptions{})
// 		if err != nil {
// 			log.Debugf("VMIRS error %v", err)
// 			return false, err
// 		}
// 		for _, obj := range objs.Items {
// 			jsonBuild, err := obj.MarshalJSON()
// 			if err != nil {
// 				log.Errorf("Error decoding VMIRS object: %s", err)
// 				return false, err
// 			}
// 			_ = json.Unmarshal(jsonBuild, &rs)
// 			if rs.Spec.Replicas != rs.Status.ReadyReplicas {
// 				log.Debugf("Waiting for replicas from VMIRS in ns %s to be running", ns)
// 				return false, nil
// 			}
// 		}
// 		return true, nil
// 	})
// 	return false, nil
// }
