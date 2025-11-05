package userinput

type UserInputCommand string

const (
	// msgKeys
	CommandKey            = "command"
	UserInputMessageIDKey = "userInputMessageID"
	ErrorKey              = "error"
	ReplyTopicKey         = "replyTopic"
	TimestampKey          = "timestamp"
	SkipReplyMessageKey   = "skipReplyMessage"

	// customer config keys
	CustomerConfigFieldKey = "customerConfigField"

	// customer config field values
	CustomerConfigFieldTicketProvider = "ticketProvider"
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
	UserInputCommandChangeIncidentStatus  = UserInputCommand("change-incident-status")
	// RuntimeIncidentPolicy mutating actions
	UserInputCommandDeleteRuntimeIncidentPolicy = UserInputCommand("delete-runtime-incident-policy")
	UserInputCommandUpdateRuntimeIncidentPolicy = UserInputCommand("update-runtime-incident-policy")
	UserInputCommandCreateRuntimeIncidentPolicy = UserInputCommand("create-runtime-incident-policy")
	// RuntimeRule mutating actions
	UserInputCommandDeleteRuntimeRule = UserInputCommand("delete-runtime-rule")
	UserInputCommandUpdateRuntimeRule = UserInputCommand("update-runtime-rule")
	UserInputCommandCreateRuntimeRule = UserInputCommand("create-runtime-rule")
	UserInputCommandCreateIgnoreRule  = UserInputCommand("create-ignore-rule")
	UserInputCommandUpdateIgnoreRule  = UserInputCommand("update-ignore-rule")
	UserInputCommandDeleteIgnoreRule  = UserInputCommand("delete-ignore-rule")
	//ApplicationProfile actions
	UserInputCommandAddToApplicationProfile      = UserInputCommand("add-to-application-profile")
	UserInputCommandDeleteFromApplicationProfile = UserInputCommand("delete-from-application-profile")

	UserInputCommandDeleteWorkflow = UserInputCommand("delete-workflow")
	UserInputCommandUpdateWorkflow = UserInputCommand("update-workflow")
	UserInputCommandCreateWorkflow = UserInputCommand("create-workflow")

	UserInputCommandDeleteRegistry      = UserInputCommand("delete-registry")
	UserInputCommandUpdateRegistry      = UserInputCommand("update-registry")
	UserInputCommandCreateRegistry      = UserInputCommand("create-registry")
	UserInputCommandCheckRegistry       = UserInputCommand("check-registry")
	UserInputCommandClearRegistryStatus = UserInputCommand("clear-registry-status")

	UserInputCommandDeleteTeamsChannel = UserInputCommand("delete-teams-channel")
	UserInputCommandUpdateTeamsChannel = UserInputCommand("update-teams-channel")
	UserInputCommandCreateTeamsChannel = UserInputCommand("create-teams-channel")

	//cloud account actions
	UserInputCommandCreateCloudAccount   = UserInputCommand("create-cloud-account")
	UserInputCommandUpdateCloudAccount   = UserInputCommand("update-cloud-account")
	UserInputCommandDeleteCloudAccount   = UserInputCommand("delete-cloud-account")
	UserInputCommandDeleteAccountFeature = UserInputCommand("delete-account-feature")
	UserInputCommandScanCloudAccount     = UserInputCommand("scan-cloud-account")

	//aws organization actions
	UserInputCommandCreateOrganization        = UserInputCommand("create-organization")
	UserInputCommandUpdateOrganization        = UserInputCommand("update-organization")
	UserInputCommandPartialUpdateOrganization = UserInputCommand("partial-update-organization")
	UserInputCommandDeleteOrganization        = UserInputCommand("delete-organization")
	UserInputCommandDeleteOrganizationFeature = UserInputCommand("delete-organization-feature")
	UserInputCommandScanOrganization          = UserInputCommand("sync-organization")

	UserInputCommandSendOperatorApiCommand = UserInputCommand("send-operator-api-command")

	// TO DEPRECATE
	UserInputCommandNewJiraTicketByUser = UserInputCommand("new-jira-ticket-by-user")
	// TO DEPRECATE
	UserInputCommandUnlinkJiraTicketByUser = UserInputCommand("unlink-jira-ticket-by-user")

	UserInputCommandNewTicketByUser    = UserInputCommand("new-ticket-by-user")
	UserInputCommandUnlinkTicketByUser = UserInputCommand("unlink-ticket-by-user")

	UserInputCommandRuntimeIncidentResponse = UserInputCommand("runtime-incident-response")

	UserInputCommandApplyNetworkPolicy  = UserInputCommand("apply-network-policy")
	UserInputCommandApplySeccompProfile = UserInputCommand("apply-seccomp-profile")

	UserInputCommandDeleteSavedFilter = UserInputCommand("delete-saved-filter")
	UserInputCommandUpdateSavedFilter = UserInputCommand("update-saved-filter")
	UserInputCommandCreateSavedFilter = UserInputCommand("create-saved-filter")

	UserInputCommandDeleteWebhook = UserInputCommand("delete-webhook")
	UserInputCommandUpdateWebhook = UserInputCommand("update-webhook")
	UserInputCommandCreateWebhook = UserInputCommand("create-webhook")

	// Siem Integration actions
	UserInputCommandCreateSIEMIntegration = UserInputCommand("create-siem-integration")
	UserInputCommandUpdateSIEMIntegration = UserInputCommand("update-siem-integration")
	UserInputCommandDeleteSIEMIntegration = UserInputCommand("delete-siem-integration")

	// Customer config actions
	UserInputCommandUpdateCustomerConfig = UserInputCommand("update-customer-config")
)
