package common

const (
	// attack chains states topic names
	AttackChainStateScanStateTopic = "attack-chain-scan-state-v1"
	AttackChainStateViewedTopic    = "attack-chain-viewed-v1"
	AttackChainStateDeleteTopic    = "attack-chain-delete-v1"

	// on finish topics
	KubescapeScanReportFinishedTopic   = "kubescape-scan-report-finished-v1"
	ContainerScanReportFinishedTopic   = "container-scan-report-finished-v1"
	AttackChainScanReportFinishTopic   = "attack-chain-scan-report-finished-v1"
	SynchronizerFinishTopic            = "synchronizer-finished-v1"
	UserInputFinishTopic               = "user-input-finished-v1"
	PostureOnFinishPulsarTopic         = "kubescape-scan-report-finished-v1"
	SecurityRisksOnFinishPulsarTopic   = "security-risks-scan-report-finished-v1"
	UIViewsIngesterOnFinishPulsarTopic = "ui-views-finished-v1"
	RuntimeIncidentOnFinishPulsarTopic = "runtime-incident-finished-v1"
	K8sObjectOnFinishPulsarTopic       = "k8s-objects-finished-v1"

	UNSOnFinishSubscriptionName = "uns-report-finished"

	CheckTenantConsumerNameTopic string = "check-tenant-consumer"

	NodeProfileTopic = "node-profile-v1"

	CloudScannerOnFinishTopic   = "cloud-scanner-cloud-findings-finished-v1"
	CloudSchedulerOnFinishTopic = "cloud-scheduler-scan-tasks-finished-v1"

	RegistryStatusTopic = "registry-status-v1"
)
