package userinput

type UserInputCommand string

const (
	// msgKeys
	CommandKey            = "command"
	UserInputMessageIDKey = "userInputMessageID"
	ErrorKey              = "error"
	ReplyTopicKey         = "replyTopic"
	TimestampKey          = "timestamp"
)

const (
	UserInputCommandDeleteCluster         = UserInputCommand("delete-cluster")
	UserInputCommandDeleteAllCustomerData = UserInputCommand("delete-all-customer-data")
	UserInputCommandCreateCustomerData    = UserInputCommand("create-customer-data")
	// vulnerability scans actions
	UserInputCommandMarkForDeletionRegistryScan                 = UserInputCommand("mark-for-deletion-registry-scan")
	UserInputCommandStoreContainerScanningSummaryStub           = UserInputCommand("store-container-scanning-summary-stub")
	UserInputCommandMarkForDeletionKubernetesResources          = UserInputCommand("mark-for-deletion-kubernetes-resources")
	UserInputCommandMarkForDeletionContainerScanSummary         = UserInputCommand("mark-for-deletion-container-scan-summary")
	UserInputCommandMarkForDeletionRegistryScanVulnerabilities  = UserInputCommand("mark-for-deletion-registry-scan-vulnerabilities")
	UserInputCommandMarkForDeletionContainerScanVulnerabilities = UserInputCommand("mark-for-deletion-container-scan-vulnerabilities")
	// PostureReport actions
	UserInputCommandMarkForDeletionClusterPostureReport    = UserInputCommand("mark-for-deletion-cluster-posture-report")
	UserInputCommandMarkForDeletionRepositoryPostureReport = UserInputCommand("mark-for-deletion-repository-posture-report")
	// SecurityRisk actions
	UserInputCommandStoreExceptionSecurityRisk                   = UserInputCommand("store-exception-security-risk")
	UserInputCommandUpdateExceptionSecurityRisk                  = UserInputCommand("update-exception-security-risk")
	UserInputCommandMarkForDeletionExceptionSecurityRisk         = UserInputCommand("mark-for-deletion-exception-security-risk")
	UserInputCommandMarkForDeletionKubernetesResourceRelatedData = UserInputCommand("mark-for-deletion-kubernetes-resource-related-data")
	// RuntimeIncident actions
	UserInputCommandUpdateRunTimeIncident = UserInputCommand("update-run-time-incident")
	UserInputCommandDeleteRunTimeIncident = UserInputCommand("delete-run-time-incident") // for backoffice
	// RuntimeIncidentPolicy mutating actions
	UserInputCommandDeleteRuntimeIncidentPolicy = UserInputCommand("delete-runtime-incident-policy")
	UserInputCommandUpdateRuntimeIncidentPolicy = UserInputCommand("update-runtime-incident-policy")
	UserInputCommandCreateRuntimeIncidentPolicy = UserInputCommand("create-runtime-incident-policy")
	UserInputCommandCreateIgnoreRule            = UserInputCommand("create-ignore-rule")
	UserInputCommandUpdateIgnoreRule            = UserInputCommand("update-ignore-rule")
	UserInputCommandDeleteIgnoreRule            = UserInputCommand("delete-ignore-rule")
	//ApplicationProfile actions
	UserInputCommandAddToApplicationProfile      = UserInputCommand("add-to-application-profile")
	UserInputCommandDeleteFromApplicationProfile = UserInputCommand("delete-from-application-profile")

	UserInputCommandDeleteWorkflow = UserInputCommand("delete-workflow")
	UserInputCommandUpdateWorkflow = UserInputCommand("update-workflow")
	UserInputCommandCreateWorkflow = UserInputCommand("create-workflow")

	UserInputCommandDeleteRegistry = UserInputCommand("delete-registry")
	UserInputCommandUpdateRegistry = UserInputCommand("update-registry")
	UserInputCommandCreateRegistry = UserInputCommand("create-registry")
	UserInputCommandCheckRegistry  = UserInputCommand("check-registry")

	UserInputCommandDeleteTeamsChannel = UserInputCommand("delete-teams-channel")
	UserInputCommandUpdateTeamsChannel = UserInputCommand("update-teams-channel")
	UserInputCommandCreateTeamsChannel = UserInputCommand("create-teams-channel")
)
