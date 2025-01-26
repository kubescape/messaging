package cspm

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
