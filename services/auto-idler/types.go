package main

// LagoonEnvironments type
type LagoonEnvironments struct {
	Data struct {
		DevelopmentEnvironments []struct {
			Name      string `json:"name"`
			AutoIdle  int    `json:"autoIdle"`
			Openshift struct {
				ConsoleURL string `json:"consoleUrl"`
				Token      string `json:"token"`
				Name       string `json:"name"`
			} `json:"openshift"`
			Environments []struct {
				OpenshiftProjectName string `json:"openshiftProjectName"`
				Name                 string `json:"name"`
				AutoIdle             int    `json:"autoIdle,omitempty"`
			} `json:"environments,omitempty"`
		} `json:"developmentEnvironments"`
	} `json:"data"`
}
