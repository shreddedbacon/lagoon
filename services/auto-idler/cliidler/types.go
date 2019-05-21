package cliidler

// DeploymentConfig type
type DeploymentConfig struct {
	Status struct {
		Replicas int `json:"replicas"`
	} `json:"status"`
}

// BuildStatus type
type BuildStatus struct {
	Items []struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}
