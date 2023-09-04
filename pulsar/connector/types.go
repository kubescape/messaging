package connector

// structs for Pulsar messages

type AttackChainEngineIngesterMessage struct {
	ImageScanID  string `json:"imageScanID"`
	ReportGUID   string `json:"reportGUID"`
	CustomerGUID string `json:"customerGUID"`
	ClusterName  string `json:"clusterName"`
}

type AttackChainFirstSeen struct {
	AttackChainID    string `json:"attackChainID,omitempty" bson:"attackChainID,omitempty"` // name/cluster/resourceID
	CustomerGUID     string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	ViewedMainScreen string `json:"viewedMainScreen,omitempty" bson:"viewedMainScreen,omitempty"`
}

type AttackChainScanStatus struct {
	ClusterName      string `json:"clusterName,omitempty" bson:"clusterName,omitempty"`
	CustomerGUID     string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
	ProcessingStatus string `json:"processingStatus,omitempty" bson:"processingStatus,omitempty"` // "processing"/ "done"
}
type AttackChainDelete struct {
	ClusterName    string `json:"clusterName,omitempty" bson:"clusterName,omitempty"` 
	CustomerGUID     string `json:"customerGUID,omitempty" bson:"customerGUID,omitempty"`
}
func (acps *AttackChainScanStatus) GetCustomerGUID() string {
	return acps.CustomerGUID
}

func (acfs *AttackChainFirstSeen) GetCustomerGUID() string {
	return acfs.CustomerGUID
}
