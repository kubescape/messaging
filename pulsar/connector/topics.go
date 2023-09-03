package connector

import (
	"fmt"
)

type TopicName string

type TopicPersistency string

const (
	//topic persistency prefix
	TopicTypePersistent    TopicPersistency = "persistent"
	TopicTypeNonPersistent TopicPersistency = "non-persistent"
	// topic names
	AttackChainStateScanStateTopic   = "attack-chain-scan-state-v1"
	AttackChainStateViewedTopic      = "attack-chain-viewed-v1"
	KubescapeScanReportFinishedTopic = "kubescape-scan-report-finished-v1"
)

func BuildTopic(persistency TopicPersistency, tenant, namespace string, topicName TopicName) string {
	return fmt.Sprintf("%s://%s/%s/%s", persistency, tenant, namespace, topicName)
}

func BuildPersistentTopic(tenant, namespace string, topicName TopicName) string {
	return BuildTopic(TopicTypePersistent, tenant, namespace, topicName)
}

func BuildNonPersistentTopic(tenant, namespace string, topicName TopicName) string {
	return BuildTopic(TopicTypeNonPersistent, tenant, namespace, topicName)
}
