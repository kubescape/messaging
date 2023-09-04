package attackchains

const (
	// topic names
	AttackChainStateScanStateTopic   = "attack-chain-scan-state-v1"
	AttackChainStateViewedTopic      = "attack-chain-viewed-v1"
	AttackChainStateDeleteTopic      = "attack-chain-delete-v1"
	KubescapeScanReportFinishedTopic = "kubescape-scan-report-finished-v1"
)

// --------- Ingesters structs and consts -------------

// supported topics and properties:
// [topic]/[propName]/[propValue]

// attack-chain-scan-state-v1/action/update
// attack-chain-viewed-v1/action/update

const (
	MsgPropAction            = "action"
	MsgPropActionValueUpdate = "update"
	NsgPropActionValueDelete = "delete"
)
