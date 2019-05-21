package minioc

import (
	"encoding/json"
	"errors"
	"regexp"
)

// GetPodsInNamespaceByRegex get pods in a namespace using a regex pattern
func (openshift *OpenShift) GetPodsInNamespaceByRegex(openshiftURL string, openshiftToken string, openshiftNamespace string, matchRegex string) (string, error) {
	// get if there are any pods in the project
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "api/v1/namespaces/"+openshiftNamespace+"/pods", "")
	if err != nil {
		return "", errors.New("Error performing check of pods in namespace")
	}
	podsInNamespace := NamespacePods{}
	json.Unmarshal([]byte(apiResponse), &podsInNamespace)
	for _, pods := range podsInNamespace.Items {
		re, _ := regexp.Compile(matchRegex)
		if re.Match([]byte(pods.Metadata.Name)) {
			return pods.Metadata.Name, nil
		}
	}
}

// GetPodOwner get pod information
func (openshift *OpenShift) GetPodOwner(openshiftURL string, openshiftToken string, openshiftNamespace string, podName string) (string, error) {
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "api/v1/namespaces/"+openshiftNamespace+"/pods/"+podName, "")
	if err != nil {
		return "", errors.New("Error performing check of pod owner")
	}
	return apiResponse, nil	
}
