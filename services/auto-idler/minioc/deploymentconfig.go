package minioc

import (
	"encoding/json"
	"errors"
)

// GetDeploymentScale get the deployment scale information
func (openshift *OpenShift) GetDeploymentScale(openshiftURL string, openshiftToken string, openshiftNamespace string, deploymentName string) (string, error) {
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/deploymentconfigs/"+deploymentName+"/scale", "")
	if err != nil {
		return "", errors.New("Error performing deployment config scale check")
	}
	return apiResponse, nil
}

// PutDeploymentScale put the new deployment scale
func (openshift *OpenShift) PutDeploymentScale(openshiftURL string, openshiftToken string, openshiftNamespace string, deploymentName string, newScale string) (string, error) {
	apiResponse, err := openshift.RunHTTPRequest("PUT", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/deploymentconfigs/"+deploymentName+"/scale", newScale)
	if err != nil {
		return "", errors.New("Error performing deployment config put operation")
	}
	return apiResponse, nil
}

//GetDeploymentConfig get deployment config for a deployment
func (openshift *OpenShift) GetDeploymentConfig(openshiftURL string, openshiftToken string, openshiftNamespace string, deploymentName string) (string, error) {
	// get if there is any deployment config for cli
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/deploymentconfigs/"+deploymentName, "")
	if err != nil {
		return "", err
	}
	return apiResponse, nil
}

// IdleDeployment idle a deployment config
func (openshift *OpenShift) IdleDeployment(openshiftURL string, openshiftToken string, openshiftNamespace string, deploymentName string) (string, error) {
	// get the current replica config, set the replicas to 0 and put the new config back
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/deploymentconfigs/"+deploymentName+"/scale", "")
	if err != nil {
		return apiResponse, err
	}
	depReplicas := DeploymentReplicas{}
	json.Unmarshal([]byte(apiResponse), &depReplicas)
	depReplicas.Spec.Replicas = 0
	updateDepReplicas, err := json.Marshal(depReplicas)
	apiResponse2, err2 := openshift.RunHTTPRequest("PUT", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/deploymentconfigs/"+deploymentName+"/scale", string(updateDepReplicas))
	if err2 != nil {
		return apiResponse2, err2
	}
	return apiResponse2, nil
}
