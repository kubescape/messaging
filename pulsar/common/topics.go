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
	ObjectModificationOnFinishTopic    = "object-modification-finished-v1"
	ClusterStatusOnFinishPulsarTopic   = "cluster-status-finished-v1"

	UNSOnFinishSubscriptionName = "uns-report-finished"

	CheckTenantConsumerNameTopic string = "check-tenant-consumer"

	NodeProfileTopic = "node-profile-v1"

	NetworkStreamTopic = "network-stream-v1"

	CloudScannerOnFinishTopic   = "cloud-scanner-results-v2"
	CloudSchedulerOnFinishTopic = "cloud-scanner-tasks-v2"
	CloudSchedulerCommandTopic  = "cloud-scheduler-command-v1"

	RegistryStatusTopic = "registry-status-v1"

	NewJiraTicketByUNSTopic = "new-jira-ticket-by-uns-v1"
)
