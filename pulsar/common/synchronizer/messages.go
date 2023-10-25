package synchronizer

const (
	// MsgPropEvent is the property name for the event type
	MsgPropEvent                         = "event"
	MsgPropEventValueGetObjectMessage    = "GetObject"
	MsgPropEventValuePatchObjectMessage  = "PatchObject"
	MsgPropEventValueVerifyObjectMessage = "VerifyObject"
)

// FIXME we need to document how to process these messages

// structs for Pulsar messages for synchronizer topic

type DeleteObjectMessage struct {
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
}

type GetObjectMessage struct {
	BaseObject   []byte `json:"baseObject"`
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
}

type NewChecksumMessage struct {
	Checksum     string `json:"checksum"`
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
}

type NewObjectMessage struct {
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
	Object       []byte `json:"patch"`
}

type PatchObjectMessage struct {
	Checksum     string `json:"checksum"`
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
	Patch        []byte `json:"patch"`
}

type PutObjectMessage struct {
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
	Object       []byte `json:"patch"`
}

type VerifyObjectMessage struct {
	Checksum     string `json:"checksum"`
	ClusterName  string `json:"clusterName"`
	CustomerGUID string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	Depth        int    `json:"depth"`
	Kind         string `json:"kind"`
	MsgId        string `json:"msgId"`
	Name         string `json:"name"`
}
