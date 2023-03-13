package pulsarconnector

import (
	"fmt"
)

type TopicName string

// default is persistent
func GetTopic(topicName TopicName) string {
	return BuildPersistentTopic(topicName)
}

type TopicPersistency string

const (
	//topic persistency prefix
	TopicTypePersistent    TopicPersistency = "persistent"
	TopicTypeNonPersistent TopicPersistency = "non-persistent"
)

func BuildTopic(persistency TopicPersistency, topicName TopicName) string {
	return fmt.Sprintf("%s://%s/%s/%s", persistency, GetClientConfig().Tenant, GetClientConfig().Namespace, topicName)
}

func BuildPersistentTopic(topicName TopicName) string {
	return BuildTopic(TopicTypePersistent, topicName)
}

func BuildNonPersistentTopic(topicName TopicName) string {
	return BuildTopic(TopicTypePersistent, topicName)
}
