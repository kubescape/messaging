package common

const (
	KubescapeScanReportFinishedTopic = "kubescape-scan-report-finished-v1"
	ContainerScanReportFinishedTopic = "container-scan-report-finished-v1"
	AttackChainScanReportFinishTopic = "attack-chain-scan-report-finished-v1"
	SynchronizerFinishTopic          = "synchronizer-finished-v1"
)

// --------- Ingesters structs and consts -------------

// supported topics and properties:
// [topic]/[propName]/[propValue]

// attack-chain-scan-state-v1/action/update
// attack-chain-viewed-v1/action/update

// pulsar messaging properties and values
const (
	MsgPropAction            = "action"
	MsgPropActionValueUpdate = "update"
	NsgPropActionValueDelete = "delete"
)
