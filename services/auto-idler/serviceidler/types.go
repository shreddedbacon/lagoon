package serviceidler

import (
	"time"
)

// BuildStatus type
type BuildStatus struct {
	Items []struct {
		Status struct {
			Phase string `json:"phase"`
		} `json:"status"`
	} `json:"items"`
}

// ElasticSearchResponse type
type ElasticSearchResponse struct {
	Took     int  `json:"took"`
	TimedOut bool `json:"timed_out"`
	Shards   struct {
		Total      int `json:"total"`
		Successful int `json:"successful"`
		Skipped    int `json:"skipped"`
		Failed     int `json:"failed"`
	} `json:"_shards"`
	Hits struct {
		Total    int           `json:"total"`
		MaxScore float64       `json:"max_score"`
		Hits     []interface{} `json:"hits"`
	} `json:"hits"`
}

// EntryPointsConfig type
type EntryPointsConfig struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink        string `json:"selfLink"`
		ResourceVersion string `json:"resourceVersion"`
	} `json:"metadata"`
	Items []EntryPointConfig `json:"items"`
}

// EntryPointConfig type
type EntryPointConfig struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		Name              string    `json:"name"`
		Namespace         string    `json:"namespace"`
		SelfLink          string    `json:"selfLink"`
		UID               string    `json:"uid"`
		ResourceVersion   string    `json:"resourceVersion"`
		CreationTimestamp time.Time `json:"creationTimestamp"`
		Labels            struct {
			Branch  string `json:"branch"`
			Project string `json:"project"`
			Service string `json:"service"`
		} `json:"labels"`
		Annotations struct {
			IdlingAlphaOpenshiftIoIdledAt       time.Time `json:"idling.alpha.openshift.io/idled-at"`
			IdlingAlphaOpenshiftIoUnidleTargets string    `json:"idling.alpha.openshift.io/unidle-targets"`
		} `json:"annotations"`
	} `json:"metadata"`
	Subsets []struct {
		Addresses []struct {
			IP        string `json:"ip"`
			NodeName  string `json:"nodeName"`
			TargetRef struct {
				Kind            string `json:"kind"`
				Namespace       string `json:"namespace"`
				Name            string `json:"name"`
				UID             string `json:"uid"`
				ResourceVersion string `json:"resourceVersion"`
			} `json:"targetRef"`
		} `json:"addresses"`
		Ports []struct {
			Name     string `json:"name"`
			Port     int    `json:"port"`
			Protocol string `json:"protocol"`
		} `json:"ports"`
	} `json:"subsets"`
}
