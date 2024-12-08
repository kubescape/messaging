package common

// --------- Ingesters structs and consts -------------

// supported topics and properties:
// [topic]/[propName]/[propValue]

// attack-chain-scan-state-v1/action/update
// attack-chain-viewed-v1/action/update

// pulsar messaging properties and values
const (
	MsgPropAction            = "action"
	MsgPropActionValueUpdate = "update"
	MsgPropActionValueDelete = "delete"

	MsgPropAccount = "account"
	MsgPropCluster = "cluster"

	MsgPropMessageType = "messageType"
	MsgType            = "CDRALERT"
)

// MsgType for message types
const (
	CdrAlert = "CDRALERT"
)
