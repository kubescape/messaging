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
	CustomerGUID string `json:"customerGUID,omitempty"`

	CloudMetadata *armotypes.CloudMetadata `json:"cloudMetadata,omitempty"`

	// this is populated only if we have cloud metadata and armo account name associated with it.
	ArmoAccountName string `json:"armoAccountName,omitempty"`

	Cluster        string     `json:"cluster,omitempty"`
	Status         string     `json:"status,omitempty"`
	Provider       string     `json:"provider,omitempty"`
	LastKeepAlive  *time.Time `json:"lastKeepAlive,omitempty"`
	ConnectionTime *time.Time `json:"connectionTime,omitempty"`

	// for degrated
	AffectedPods  []string `json:"affectedPods,omitempty"`
	AffectedNodes []string `json:"affectedNodes,omitempty"`

	Link string `json:"link,omitempty"`
}

func (c *ClusterStatusNotification) SameClusterStatus(status string) bool {
	return c.Status == status
}
