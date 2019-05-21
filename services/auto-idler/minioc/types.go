package minioc

import (
	"time"
)

// IdleDeploymentReplicas type
type IdleDeploymentReplicas struct {
	Services []DeploymentReplicas
}

// DeploymentReplicas type
type DeploymentReplicas struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		SelfLink          string    `json:"selfLink"`
		UID               string    `json:"uid"`
		ResourceVersion   string    `json:"resourceVersion"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Annotations       struct {
			IdlingAlphaOpenshiftIoIdledAt       time.Time `json:"idling.alpha.openshift.io/idled-at"`
			IdlingAlphaOpenshiftIoPreviousScale string    `json:"idling.alpha.openshift.io/previous-scale"`
		} `json:"annotations"`
	} `json:"metadata"`
	Spec struct {
		Replicas int `json:"replicas"`
	} `json:"spec"`
	Status struct {
		Replicas int `json:"replicas"`
		Selector struct {
			Deploymentconfig string `json:"deploymentconfig"`
			Service          string `json:"service"`
		} `json:"selector"`
		TargetSelector string `json:"targetSelector"`
	} `json:"status"`
}

// NamespacePods type
type NamespacePods struct {
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			GenerateName      string    `json:"generateName"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			UID               string    `json:"uid"`
			ResourceVersion   string    `json:"resourceVersion"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
			Labels            struct {
				Branch           string `json:"branch"`
				Deployment       string `json:"deployment"`
				Deploymentconfig string `json:"deploymentconfig"`
				Project          string `json:"project"`
				Service          string `json:"service"`
			} `json:"labels"`
		} `json:"metadata,omitempty"`
	} `json:"items"`
}
