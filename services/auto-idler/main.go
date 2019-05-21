package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"time"

	ci "github.com/amazeeio/lagoon/services/autoidler/cliidler"
	si "github.com/amazeeio/lagoon/services/autoidler/serviceidler"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// set up the globals
var (
	tokenSigningKey  = os.Getenv("JWTSECRET")
	jwtAudience      = os.Getenv("JWTAUDIENCE")
	projectRegex     = os.Getenv("PROJECT_REGEX")
	logsDBAdminPass  = os.Getenv("LOGSDB_ADMIN_PASSWORD")
	logsDBHost       = os.Getenv("LOGSDB_ADMIN_HOST")
	graphQLEndpoint  = os.Getenv("GRAPHQL_ENDPOINT")
	serviceIdlerTime = os.Getenv("SERVICE_IDLER_TIME") // define as $h$m, $h, $m
	httpPort         = os.Getenv("HTTP_PORT")          //port to start on
	netClientTimeout = 10                              //in seconds, a timeout for requests
)

var graphQLQuery = `query developmentEnvironments {
  developmentEnvironments:allProjects {
    name
    autoIdle
    openshift {
      consoleUrl
      token
      name
    }
    environments(type: DEVELOPMENT) {
      openshiftProjectName
      name
      autoIdle
    }
  }
}`

// the query to perform against router logs for last hits
var serviceIdlerESQuery = `{
"size": 0,
	"query": {
		"bool": {
			"filter": {
				"range": {
					"@timestamp": {
						"gte": "now-` + serviceIdlerTime + `"
					}
				}
			}
		}
	}
}`

func main() {
	log.SetOutput(os.Stdout) // log to stdout
	// make sure we have the envvars or bail
	if len(os.Getenv("JWTSECRET")) == 0 {
		log.Fatalln("JWTSECRET not set")
	}
	if len(os.Getenv("JWTAUDIENCE")) == 0 {
		log.Fatalln("JWTAUDIENCE not set")
	}
	if len(os.Getenv("PROJECT_REGEX")) == 0 {
		log.Fatalln("PROJECT_REGEX not set")
	}
	if len(os.Getenv("LOGSDB_ADMIN_PASSWORD")) == 0 {
		log.Fatalln("LOGSDB_ADMIN_PASSWORD not set")
	}
	if len(os.Getenv("HTTP_PORT")) == 0 {
		log.Fatalln("HTTP_PORT not set")
	}
	if len(os.Getenv("LOGSDB_ADMIN_HOST")) == 0 {
		log.Fatalln("LOGSDB_ADMIN_HOST not set")
	}
	if len(os.Getenv("GRAPHQL_ENDPOINT")) == 0 {
		log.Fatalln("GRAPHQL_ENDPOINT not set")
	}
	if len(os.Getenv("SERVICE_IDLER_TIME")) == 0 {
		log.Fatalln("SERVICE_IDLER_TIME not set")
	}

	// set up a netclient with a timeout on requests
	var netClient = &http.Client{
		Timeout: time.Second * time.Duration(netClientTimeout),
	}
	// ignore ssl cert warnings
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	// set up some http listeners
	// send favicon to stop chrome sending two requests
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		w.Header().Set("Cache-Control", "public, max-age=7776000")
		fmt.Fprintf(w, "data:image/gif;base64,R0lGODlhAQABAIABAAAAAP///yH5BAEAAAEALAAAAAABAAEAAAICTAEAOw==")
	})
	// one for the cli idler
	http.HandleFunc("/idler/cli", func(w http.ResponseWriter, r *http.Request) {
		go StartCliIdler(netClient, uuid.New().String())
		fmt.Fprintf(w, "cli idler started")
	})
	// one for the service idler
	http.HandleFunc("/idler/service", func(w http.ResponseWriter, r *http.Request) {
		go StartServiceIdler(netClient, uuid.New().String())
		fmt.Fprintf(w, "service idler started")
	})
	// one for the generic
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "nothing to see here")
	})
	// start the http listener
	log.Fatal(http.ListenAndServe(":"+httpPort, nil))
}

// GetJWTToken generate a jwt token
func GetJWTToken(tokenAudience string, tokenSigningKey string) (string, error) {
	// generate a token with our claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iat":  int32(time.Now().Unix()),
		"role": "admin",
		"iss":  "auto-idler",
		"aud":  "" + tokenAudience + "",
		"sub":  "auto-idler",
	})
	// sign and get the complete encoded token as a string using the secret
	tokenString, tokenErr := token.SignedString([]byte(tokenSigningKey))
	return tokenString, tokenErr
}

// StartCliIdler kick off a cli idler over all the environments from graphql
func StartCliIdler(netClient *http.Client, reqUUID string) {
	graphQLBody, err := GetTokenRunQuery(netClient)
	if err != nil {
		log.Println(err)
		return
	}
	// dump the results of the graphql query
	lagoonEnvironments := LagoonEnvironments{}
	json.Unmarshal([]byte(graphQLBody), &lagoonEnvironments)

	for _, project := range lagoonEnvironments.Data.DevelopmentEnvironments {
		if project.AutoIdle == 1 {
			r, _ := regexp.Compile(projectRegex)
			if r.Match([]byte(project.Name)) {
				log.Printf("%v: %v - %v: Found\n", reqUUID, project.Openshift.ConsoleURL, project.Name)
				if len(project.Environments) > 0 {
					for _, environment := range project.Environments {
						// create the cli idler
						//log.Println(environment.AutoIdle)
						cliIdler, _ := ci.NewCliIdler(project.Openshift.ConsoleURL, environment.OpenshiftProjectName, project.Openshift.Token, project.Name, environment.Name, netClient, reqUUID)
						cliIdlerErr := cliIdler.RunCliIdler()
						if cliIdlerErr != nil {
							log.Println(cliIdlerErr)
						}
					}
				} else {
					log.Printf("%v: %v - %v: has no environments\n", reqUUID, project.Openshift.ConsoleURL, project.Name)
				}
			}
		}
	}
}

// StartServiceIdler kick off a service idler over all the environments from graphql
func StartServiceIdler(netClient *http.Client, reqUUID string) {
	graphQLBody, err := GetTokenRunQuery(netClient)
	if err != nil {
		log.Println(err)
		return
	}
	// dump the results of the graphql query
	lagoonEnvironments := LagoonEnvironments{}
	json.Unmarshal([]byte(graphQLBody), &lagoonEnvironments)
	for _, project := range lagoonEnvironments.Data.DevelopmentEnvironments {
		if project.AutoIdle == 1 {
			r, _ := regexp.Compile(projectRegex)
			if r.Match([]byte(project.Name)) {
				log.Printf("%v: %v - %v: Found\n", reqUUID, project.Openshift.ConsoleURL, project.Name)
				if len(project.Environments) > 0 {
					for _, environment := range project.Environments {
						// check if the environment has autoidling enabled
						if environment.AutoIdle == 1 {
							// create the services idler
							serviceIdler, _ := si.NewServiceIdler(project.Openshift.ConsoleURL, environment.OpenshiftProjectName, project.Openshift.Token, project.Name, environment.Name, logsDBHost, logsDBAdminPass, serviceIdlerTime, serviceIdlerESQuery, netClient, reqUUID)
							serviceIdlerErr := serviceIdler.RunServiceIdler()
							if serviceIdlerErr != nil {
								log.Println(serviceIdlerErr)
							}
						} else {
							log.Printf("%v: %v - %v: %v has auto-idling disabled\n", reqUUID, project.Openshift.ConsoleURL, project.Name, environment.Name)
						}
					}
				} else {
					log.Printf("%v: %v - %v: has no environments\n", reqUUID, project.Openshift.ConsoleURL, project.Name)
				}
			}
		}
	}
}

// GetTokenRunQuery get a jwt, then run the query. return the graphql result
func GetTokenRunQuery(netClient *http.Client) ([]byte, error) {
	// sign and get the complete encoded token as a string using the secret
	tokenString, tokenErr := GetJWTToken(jwtAudience, tokenSigningKey)
	if tokenErr != nil {
		log.Println(tokenErr)
		return []byte(""), tokenErr
	}
	// query the graphql endpoint
	graphQLBody, graphQLErr := RunGraphQLQuery(graphQLEndpoint, graphQLQuery, tokenString, netClient)
	if graphQLErr != nil {
		log.Println(graphQLErr)
		return []byte(""), graphQLErr
	}
	return graphQLBody, nil
}

// RunGraphQLQuery run a graphql query against an api endpoint, return the result
func RunGraphQLQuery(apiURL string, graphQLQuery string, tokenString string, netClient *http.Client) ([]byte, error) {
	// fix up the graphql so its one line
	re := regexp.MustCompile("\n")
	newQuery := re.ReplaceAllString(graphQLQuery, "\\n")
	// make the request
	apiRequest, apiRequestErr := http.NewRequest("POST", apiURL, bytes.NewBuffer([]byte("{\"query\": \""+newQuery+"\"}")))
	if apiRequestErr != nil {
		log.Println(apiRequestErr)
		return []byte(""), apiRequestErr
	}
	// set the headers
	apiRequest.Header.Add("Authorization", "bearer "+tokenString)
	apiRequest.Header.Set("Content-Type", "application/json")
	apiResponse, apiResponseErr := netClient.Do(apiRequest)
	if apiResponseErr != nil {
		log.Println(apiResponseErr)
		return []byte(""), apiResponseErr
	}
	defer apiResponse.Body.Close()
	apiResponseBody, _ := ioutil.ReadAll(apiResponse.Body)
	// return the resulting json
	if apiResponse.StatusCode != 200 {
		sprint := fmt.Sprintf("%v: %v", apiResponse.StatusCode, string(apiResponseBody))
		return apiResponseBody, errors.New(sprint)
	}
	return apiResponseBody, nil
}
