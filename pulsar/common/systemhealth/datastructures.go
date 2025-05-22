package systemhealth

import (
	"time"

	"github.com/armosec/armoapi-go/armotypes"
)

type ClusterStatusOnFinishedMessage struct {
	ReportTime                time.Time                   `json:"reportTime"`
	SendTime                  time.Time                   `json:"sendTime"`
	CustomerGUID              string                      `json:"customerGUID"`
	ClusterStatusNotification []ClusterStatusNotification `json:"clusterStatusNotification"`
}

type ClusterStatusNotification struct {
	CustomerGUID string `json:"customer_guid,omitempty"`

	CloudMetadata  *armotypes.CloudMetadata `json:"cloud_metadata,omitempty"`
	Cluster        string                   `json:"cluster,omitempty"`
	Status         string                   `json:"status,omitempty"`
	Provider       string                   `json:"provider,omitempty"`
	LastKeepAlive  *time.Time               `json:"last_keep_alive,omitempty"`
	ConnectionTime *time.Time               `json:"connection_time,omitempty"`

	// for degrated
	AffectedPods  []string `json:"affected_pods,omitempty"`
	AffectedNodes []string `json:"affected_nodes,omitempty"`

	Link string `json:"link,omitempty"`
}

func (c *ClusterStatusNotification) SameClusterStatus(status string) bool {
	return c.Status == status
}
