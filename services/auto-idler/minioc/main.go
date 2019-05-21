package minioc

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

type openShift interface {
	RunHTTPRequest(string, string, string, string, string) (string, error)
	GetPodProcessCount(string, string, string, string) (string, error)
	GetOpenshiftProject(string, string, string) (string, error)
	GetDeploymentConfig(string, string, string, string) (string, error)
	GetRunningBuilds(string, string, string) (string, error)
	GetPodsInNamespaceByRegex(string, string, string, string) (string, error)
	IdleDeployment(string, string, string, string) (string, error)
	GetEndpointsInNamespace(string, string, string) (string, error)
	GetPodOwner(string, string, string, string) (string, error)
	GetDeploymentScale(string, string, string, string) (string, error)
	PutDeploymentScale(string, string, string, string, string) (string, error)
	PatchServiceEndpoint(string, string, string, string, string) (string, error)
	IdleServices(string, string, string, []string) (string, error)
}

// OpenShift type
type OpenShift struct {
	netClient *http.Client
	logPrefix string
}

// NewOpenshift set parts for openshift
func NewOpenshift(netClient *http.Client, logPrefix string) openShift {
	return &OpenShift{
		netClient: netClient,
		logPrefix: logPrefix,
	}
}

// RunHTTPRequest make generic http requests to an api
func (openshift *OpenShift) RunHTTPRequest(httpMethod string, openshiftURL string, openshiftToken string, apiPath string, queryParam string) (string, error) {
	apiRequest, apiRequestErr := http.NewRequest(httpMethod, openshiftURL+apiPath, bytes.NewBuffer([]byte(queryParam)))
	if apiRequestErr != nil {
		return "", apiRequestErr
	}
	apiRequest.Header.Add("Authorization", "Bearer "+openshiftToken)
	apiRequest.Header.Set("Accept", "application/json")
	if httpMethod == "PATCH" {
		apiRequest.Header.Set("Content-Type", "application/strategic-merge-patch+json")
	} else {
		apiRequest.Header.Set("Content-Type", "application/json")
	}

	apiResponse, apiResponseErr := openshift.netClient.Do(apiRequest)
	if apiResponseErr != nil {
		return "", apiResponseErr
	}
	defer apiResponse.Body.Close()
	apiResponseBody, _ := ioutil.ReadAll(apiResponse.Body)
	apiReturnBody := string(apiResponseBody)
	if apiResponse.StatusCode != 200 {
		return apiReturnBody, errors.New("error performing check or connecting to openshift")
	}
	return apiReturnBody, nil
}

// GetPodProcessCount get the processes count from a running pod
func (openshift *OpenShift) GetPodProcessCount(openshiftURL string, openshiftToken string, openshiftNamespace string, podName string) (string, error) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	uhost, _ := url.Parse(openshiftURL)
	// `command=/bin/bash -c pgrep -P 0 | tail -n +3 | wc -l | tr -d ' '`
	queryParam := "command=/bin/bash&command=-c&command=pgrep%20-P%200|tail%20-n%20%2B3|wc%20-l|tr%20-d%20'%20'&stdin=true&stderr=true&stdout=true&tty=false"
	u := url.URL{Scheme: "wss", Host: uhost.Host, Path: "api/v1/namespaces/" + openshiftNamespace + "/pods/" + podName + "/exec", RawQuery: queryParam}
	d := websocket.Dialer{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}

	upstreamHeader := http.Header{}
	upstreamHeader.Add("Authorization", "Bearer "+openshiftToken)
	c, resp, err := d.Dial(u.String(), upstreamHeader)
	if err != nil {
		if err == websocket.ErrBadHandshake {
			errMsg := fmt.Sprintf("handshake failed with status %d", resp.StatusCode)
			return "", errors.New(errMsg)
		}
		log.Println(err)
		return "", err
	}
	defer c.Close()
	done := make(chan struct{})
	processCount := func() string {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				return ""
			}
			if len(message) > 1 {
				processCount := strings.TrimSpace(string(message))
				return processCount
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			return processCount, nil
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				return "", err
			}
		case <-interrupt:
			// close the connection by sending a close message and then waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				return "", err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return processCount, nil
		}
	}
	return processCount, nil
}

// GetOpenshiftProject get openshift project
func (openshift *OpenShift) GetOpenshiftProject(openshiftURL string, openshiftToken string, openshiftNamespace string) (string, error) {
	// get the project to check if it exists
	apiResponse, err := openshift.RunHTTPRequest("GET", openshiftURL, openshiftToken, "oapi/v1/projects/"+openshiftNamespace, "")
	if err != nil {
		return "", err
	}
	return apiResponse, nil
}

// IdleServices run through the services provided and idle them, but make sure to update the unidle-targets so it unidles properly https://access.redhat.com/documentation/en-us/openshift_container_platform/3.5/html-single/using_the_openshift_rest_api/index#idling-a-deploymentconfig
func (openshift *OpenShift) IdleServices(openshiftURL string, openshiftToken string, openshiftNamespace string, idledServices []string) (string, error) {
	// create a new idledeploymentreplicas type
	idleDepReplicas := IdleDeploymentReplicas{}
	// create a unidletargets json string
	unidleTargets := "["
	// loop through the services that were passed in
	for _, serviceName := range idledServices {
		// query the api to get the current scale of the deployment config
		apiResponse, err := openshift.GetDeploymentScale(openshiftURL, openshiftToken, openshiftNamespace, serviceName)
		if err != nil {
			return apiResponse, err
		}
		// add it to deploymentreplica type and unmarshal the json into it
		depReplicas := DeploymentReplicas{}
		json.Unmarshal([]byte(apiResponse), &depReplicas)
		// append the deploymentreplica into the idledeploymentreplicas
		idleDepReplicas.Services = append(idleDepReplicas.Services, depReplicas)
		// add the current scale t the unidle-target that we are going to patch
		previousScale := strconv.Itoa(depReplicas.Spec.Replicas)
		unidleTargets += fmt.Sprintf(`{\"kind\":\"DeploymentConfig\",\"name\":\"%v\",\"replicas\":%v},`, serviceName, previousScale)
		// verbose
		log.Println(openshift.logPrefix, "getting service replicas", serviceName)
	}
	// combine all the unidle-targets together and remove the trailing comma
	unidleTargets = strings.TrimSuffix(unidleTargets, ",")
	unidleTargets = unidleTargets + "]"
	// add it to the endpoint patch json with a timestamp
	curTime := time.Now().UTC()
	endPointPatch := fmt.Sprintf(`{
    "metadata": {
      "annotations": {
        "idling.alpha.openshift.io/idled-at": "%v",
        "idling.alpha.openshift.io/unidle-targets": "%v"
      }
    }
  }`, curTime.Format(time.RFC3339), unidleTargets)
	// patch each endpoint with all of the unidle-targets so that they all unidle when one is requested
	for _, patchService := range idleDepReplicas.Services {
		if patchService.Spec.Replicas != 0 {
			// run the patch request
			apiResponse2, err2 := openshift.PatchServiceEndpoint(openshiftURL, openshiftToken, openshiftNamespace, patchService.Metadata.Name, endPointPatch)
			if err2 != nil {
				return apiResponse2 + ": PATCH", err2
			}
			patchService.Metadata.Annotations.IdlingAlphaOpenshiftIoIdledAt = curTime
			patchService.Metadata.Annotations.IdlingAlphaOpenshiftIoPreviousScale = strconv.Itoa(patchService.Spec.Replicas)
			//verbose
			log.Println(openshift.logPrefix, "patching endpoint", patchService.Metadata.Name)
		} else {
			log.Println(openshift.logPrefix, "already idled", patchService.Metadata.Name)
		}
	}
	// put the new replica of 0 for each service
	for _, patchService := range idleDepReplicas.Services {
		if patchService.Spec.Replicas != 0 {
			patchService.Spec.Replicas = 0
			updateDepReplicas, err := json.Marshal(patchService)
			if err != nil {
				return "", err
			}
			apiResponse3, err3 := openshift.PutDeploymentScale(openshiftURL, openshiftToken, openshiftNamespace, patchService.Metadata.Name, string(updateDepReplicas))

			if err3 != nil {
				return apiResponse3, err3
			}
			// verbose
			log.Println(openshift.logPrefix, "idling service", patchService.Metadata.Name)
		}
	}
	return "", nil
}
