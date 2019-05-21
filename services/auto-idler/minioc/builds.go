package minioc

import (
	"errors"
)

// GetRunningBuilds get running builds in a namespace
func (openshift *OpenShift) GetRunningBuilds(openshiftURL string, openshiftToken string, openshiftNamespace string) (string, error) {
	// get if there are any currently running builds in openshift for thet project
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "oapi/v1/namespaces/"+openshiftNamespace+"/builds", "")
	if err != nil {
		return "", errors.New("Error performing running builds check")
	}
	return apiResponse, nil
}
