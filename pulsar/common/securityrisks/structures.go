package securityrisks

import (
	"time"

	"github.com/armosec/armoapi-go/identifiers"
	"go.uber.org/zap"
)

type SecurityRisksIngestionFinishedIngesterMessage struct {
	ReportTime             time.Time                  `json:"reportTime"`
	SendTime               time.Time                  `json:"sendTime"`
	CustomerGUID           string                     `json:"customerGUID"`
	DetectedSecurityIssues []AggregatedSecurityIssues `json:"detectedSecurityIssues"`
}

type AggregatedSecurityIssues struct {
	SecurityRiskID       string                         `json:"securityRiskID"`
	SecurityRiskName     string                         `json:"securityRiskName"`
	SecurityRiskCategory string                         `json:"securityRiskCategory"`
	SecurityRiskSeverity string                         `json:"securityRiskSeverity"`
	Resources            []identifiers.PortalDesignator `json:"resources"`
}

func (si *SecurityRisksIngestionFinishedIngesterMessage) GetLoggerFields() []zap.Field {
	fields := []zap.Field{
		zap.String("customerGUID", si.CustomerGUID),
	}

	return fields
}
