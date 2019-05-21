package serviceidler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/amazeeio/lagoon/services/autoidler/minioc"

	"github.com/go-errors/errors"
)

type serviceIdler interface {
	RunServiceIdler() error
	GetRouterLogs(string, string) (string, error)
}

// ServiceIdler type
type ServiceIdler struct {
	openshiftURL        string
	openshiftNamespace  string
	openshiftToken      string
	lagoonProject       string
	lagoonEnvironment   string
	logsDBAdminPass     string
	logsDBHost          string
	netClient           *http.Client
	logPrefix           string
	serviceIdlerTime    string
	serviceIdlerESQuery string
	reqUUID             string
}

// NewServiceIdler sets the idler with info from the main idler service
func NewServiceIdler(openshiftURL string, openshiftNamespace string, openshiftToken string, lagoonProject string, lagoonEnvironment string, logsDBHost string, logsDBAdminPass string, serviceIdlerTime string, serviceIdlerESQuery string, netClient *http.Client, reqUUID string) (serviceIdler, error) {
	logPrefix := fmt.Sprintf("%v: %v - %v: %v: ", reqUUID, openshiftURL, lagoonProject, lagoonEnvironment)
	return &ServiceIdler{
		openshiftURL:        openshiftURL,
		openshiftNamespace:  openshiftNamespace,
		openshiftToken:      openshiftToken,
		lagoonProject:       lagoonProject,
		lagoonEnvironment:   lagoonEnvironment,
		logsDBHost:          logsDBHost,
		logsDBAdminPass:     logsDBAdminPass,
		netClient:           netClient,
		logPrefix:           logPrefix,
		serviceIdlerTime:    serviceIdlerTime,
		serviceIdlerESQuery: serviceIdlerESQuery,
	}, nil
}

// RunServiceIdler the actual service idler process
func (s *ServiceIdler) RunServiceIdler() error {
	oshift := minioc.NewOpenshift(s.netClient, s.logPrefix)
	// check the router logs for the project to see if there have been any hits in the last "serviceIdlerTime"
	routerLogs, podErr := s.GetRouterLogs(s.logsDBHost, "GET")
	if podErr != nil {
		return errors.WrapPrefix(podErr, s.logPrefix, 0)
	}
	// dump the result from elastic
	elasticsearchResponse := ElasticSearchResponse{}
	json.Unmarshal([]byte(routerLogs), &elasticsearchResponse)
	projectHits := elasticsearchResponse.Hits.Total
	// check if the number of hits in the last "serviceIdlerTime" is 0 or not
	if projectHits == 0 {
		log.Println(s.logPrefix, "had no hits in last", s.serviceIdlerTime, ", starting to idle")
	} else if projectHits > 0 {
		log.Println(s.logPrefix, "has had", projectHits, "hits, hits in last", s.serviceIdlerTime, ", no idling")
		return nil
	}

	// check if there are any running builds, we don't want to idle during a build
	runningBuildsBody, runningBuildsErr := oshift.GetRunningBuilds(s.openshiftURL, s.openshiftToken, s.openshiftNamespace)
	if runningBuildsErr != nil {
		return errors.WrapPrefix(runningBuildsErr, s.logPrefix, 0)
	}
	// dump the result of the running builds
	runningBuilds := BuildStatus{}
	json.Unmarshal([]byte(runningBuildsBody), &runningBuilds)
	hasBuilds := false
	// check for running builds
	for _, build := range runningBuilds.Items {
		if build.Status.Phase == "Running" {
			log.Printf("%v builds: has running builds, skip idling\n", s.logPrefix)
			hasBuilds = true
			return nil
		}
	}
	if hasBuilds != true {
		log.Printf("%v builds: no running builds, keep idling\n", s.logPrefix)
	}

	// get the endpoints of the services
	points, pErr := oshift.GetEndpointsInNamespace(s.openshiftURL, s.openshiftToken, s.openshiftNamespace)
	if pErr != nil {
		return errors.WrapPrefix(pErr, s.logPrefix, 0)
	}
	// dump the result
	pointsJSON := EntryPointsConfig{}
	json.Unmarshal([]byte(points), &pointsJSON)
	allIdledServices := make([]string, 0)
	for _, endPoint := range pointsJSON.Items {
		switch endPoint.Metadata.Name {
		// check that the service isnt a database of some sort
		case "mariadb", "postgres":
			break
		// add it to the list
		default:
			allIdledServices = append(allIdledServices, endPoint.Metadata.Name)
		}
	}
	// idle the services
	if len(allIdledServices) > 0 {
		rep, err2 := oshift.IdleServices(s.openshiftURL, s.openshiftToken, s.openshiftNamespace, allIdledServices)
		if err2 != nil {
			log.Println(rep)
			return errors.WrapPrefix(err2, s.logPrefix, 0)
		}
	}
	return nil
}

// GetRouterLogs gets the router logs */
func (s *ServiceIdler) GetRouterLogs(httpHost string, httpMethod string) (string, error) {
	// generic http request function
	apiRequest, apiRequestErr := http.NewRequest(httpMethod, httpHost+"/router-logs-"+s.lagoonProject+"-*/_search", bytes.NewBuffer([]byte(s.serviceIdlerESQuery)))
	if apiRequestErr != nil {
		return "", apiRequestErr
	}
	apiRequest.SetBasicAuth("admin", s.logsDBAdminPass)
	apiRequest.Header.Set("Accept", "application/json")
	apiRequest.Header.Set("Content-Type", "application/json")
	apiResponse, apiResponseErr := s.netClient.Do(apiRequest)
	if apiResponseErr != nil {
		return "", apiResponseErr
	}
	defer apiResponse.Body.Close()
	apiResponseBody, _ := ioutil.ReadAll(apiResponse.Body)
	apiReturnBody := string(apiResponseBody)
	if apiResponse.StatusCode != 200 {
		retErr := fmt.Sprintf("%v http %v: error performing check", s.logPrefix, apiResponse.StatusCode)
		return "", errors.New(retErr)
	}
	return apiReturnBody, nil
}
