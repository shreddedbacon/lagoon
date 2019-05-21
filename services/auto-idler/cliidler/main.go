package cliidler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/amazeeio/lagoon/services/autoidler/minioc"

	"github.com/go-errors/errors"
)

type cliIdler interface {
	RunCliIdler() error
}

// CliIdler type
type CliIdler struct {
	openshiftURL       string
	openshiftNamespace string
	openshiftToken     string
	lagoonProject      string
	lagoonEnvironment  string
	netClient          *http.Client
	logPrefix          string
	reqUUID            string
}

//the regex used to look up cli pods from the namespace
var cliRegex = "cli-[0-9]-[a-z0-9]{5}"

// NewCliIdler sets the idler with info from the main idler service
func NewCliIdler(openshiftURL string, openshiftNamespace string, openshiftToken string, lagoonProject string, lagoonEnvironment string, netClient *http.Client, reqUUID string) (cliIdler, error) {
	logPrefix := fmt.Sprintf("%v: %v - %v: %v: ", reqUUID, openshiftURL, lagoonProject, lagoonEnvironment)
	return &CliIdler{
		openshiftURL:       openshiftURL,
		openshiftNamespace: openshiftNamespace,
		openshiftToken:     openshiftToken,
		lagoonProject:      lagoonProject,
		lagoonEnvironment:  lagoonEnvironment,
		netClient:          netClient,
		logPrefix:          logPrefix,
		reqUUID:            reqUUID,
	}, nil
}

// RunCliIdler runs the idler functions
func (cli *CliIdler) RunCliIdler() error {
	oshift := minioc.NewOpenshift(cli.netClient, cli.logPrefix)
	log.Printf("%v: %v - %v: handling environment %v\n", cli.reqUUID, cli.openshiftURL, cli.lagoonProject, cli.lagoonEnvironment)
	_, projectErr := oshift.GetOpenshiftProject(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace)
	if projectErr != nil {
		//log.Println(projectErr)
		return errors.WrapPrefix(projectErr, cli.logPrefix, 0)
	}
	log.Printf("%v checking deployment configs\n", cli.logPrefix)
	deploymentConfigs, podErr := oshift.GetDeploymentConfig(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace, "cli")
	if podErr != nil {
		return errors.WrapPrefix(podErr, cli.logPrefix, 0)
	}
	depConfig := DeploymentConfig{}
	json.Unmarshal([]byte(deploymentConfigs), &depConfig)

	if depConfig.Status.Replicas != 0 {
		log.Printf("%v cli has running pods, checking if they are non-busy\n", cli.logPrefix)
		log.Printf("%v checking running builds\n", cli.logPrefix)
		runningBuildsBody, runningBuildsErr := oshift.GetRunningBuilds(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace)
		if runningBuildsErr != nil {
			return errors.WrapPrefix(runningBuildsErr, cli.logPrefix, 0)
		}
		runningBuilds := BuildStatus{}
		json.Unmarshal([]byte(runningBuildsBody), &runningBuilds)
		hasBuilds := false
		for _, build := range runningBuilds.Items {
			if build.Status.Phase == "Running" {
				log.Printf("%v builds: has running builds\n", cli.logPrefix)
				hasBuilds = true
				return nil
			}
		}
		if hasBuilds != true {
			log.Printf("%v builds: no running builds\n", cli.logPrefix)
		}

		// CHECK PROCESS COUNT
		log.Printf("%v checking running processes\n", cli.logPrefix)
		podName, podErr := oshift.GetPodsInNamespaceByRegex(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace, cliRegex)
		if podErr != nil {
			return errors.WrapPrefix(podErr, cli.logPrefix, 0)
		}
		processCount, getPodCountErr := oshift.GetPodProcessCount(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace, podName)
		if getPodCountErr != nil {
			return errors.WrapPrefix(getPodCountErr, cli.logPrefix, 0)
		}
		if processCount[len(processCount)-1:] == "0" {
			log.Printf("%v processes: cli has %v running processes, idling down\n", cli.logPrefix, processCount)
			_, idleErr := oshift.IdleDeployment(cli.openshiftURL, cli.openshiftToken, cli.openshiftNamespace, "cli")
			if idleErr != nil {
				return errors.WrapPrefix(idleErr, cli.logPrefix, 0)
			}
		} else {
			log.Printf("%v processes: cli has %v running processes\n", cli.logPrefix, processCount)
		}
	}
	return nil
}
