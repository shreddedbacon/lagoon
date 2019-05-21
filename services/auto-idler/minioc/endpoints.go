package minioc

import (
	"errors"
)

// GetEndpointsInNamespace get the endpoints in a namespace
func (openshift *OpenShift) GetEndpointsInNamespace(openshiftURL string, openshiftToken string, openshiftNamespace string) (string, error) {
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "api/v1/namespaces/"+openshiftNamespace+"/endpoints", "")
	if err != nil {
		return "", errors.New("Error performing endpoints check")
	}
	return apiResponse, nil
}

// PatchServiceEndpoint patch endpoints in a namespace
func (openshift *OpenShift) PatchServiceEndpoint(openshiftURL string, openshiftToken string, openshiftNamespace string, serviceName string, endPointPatch string) (string, error) {
	apiResponse, err := openshift.RunHTTPRequest("PATCH", openshiftURL, openshiftToken, "api/v1/namespaces/"+openshiftNamespace+"/endpoints/"+serviceName, endPointPatch)
	if err != nil {
		return "", errors.New("Error performing patch operation")
	}
	return apiResponse, nil
}
