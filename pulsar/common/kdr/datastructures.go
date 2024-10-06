package kdr

import (
	"time"

	"github.com/armosec/armoapi-go/identifiers"
	"go.uber.org/zap"
)

const (
	RuntimeIncidentIngesterOnFinishedMessageTypeProp = "RuntimeIncidentIngesterOnFinishedMessage"
)

type RuntimeIncidentIngesterOnFinishedMessage struct {
	ReportTime          time.Time                    `json:"reportTime"`
	SendTime            time.Time                    `json:"sendTime"`
	CustomerGUID        string                       `json:"customerGUID"`
	IncidentPolicyGUIDs []string                     `json:"incidentPolicyGUIDs"`
	IncidentGUID        string                       `json:"incidentGUID"`
	IncidentName        string                       `json:"incidentName"` // incidentType.Name - ThreatName
	Severity            string                       `json:"severity"`
	Resource            identifiers.PortalDesignator `json:"resource"` // Pod, Node, Workload, Namespace, Cluster, etc.
	ResponseUpdated     bool                         `json:"responseUpdated,omitempty"`
}

func (si *RuntimeIncidentIngesterOnFinishedMessage) GetLoggerFields() []zap.Field {
	fields := []zap.Field{
		zap.String("customerGUID", si.CustomerGUID),
	}

	return fields
}
