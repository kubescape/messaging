package cspm

import (
	"encoding/json"
	"fmt"
)

// IDO: move to infra
type CloudScanCommand string

var (
	CloudScanNowCommand      CloudScanCommand = "scan-now"
	CloudStopScanCommand     CloudScanCommand = "stop-scan"
	CloudUpdateAllScheduling CloudScanCommand = "update-all-scheduling"
	CloudUpdateSchedule      CloudScanCommand = "update-schedule"
)

func (c CloudScanCommand) String() string {
	return string(c)
}

func ValidateRequest[T IScanRequest](msg []byte) (*T, error) {
	var req T
	err := json.Unmarshal(msg, &req)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal request: %v", err)
	}
	return &req, nil

}

type IScanRequest interface {
}

type ScanNowRequest struct {
	IScanRequest
	CloudAccountGUID string `json:"cloudAccountGUID"`
}
