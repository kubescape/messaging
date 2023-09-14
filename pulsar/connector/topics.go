package connector

import (
	"fmt"
	"strings"
)

type TopicName string

type TopicPersistency string

const (
	//topic persistency prefix
	TopicTypePersistent    TopicPersistency = "persistent"
	TopicTypeNonPersistent TopicPersistency = "non-persistent"
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

func GetTopicName(topic string) string {
	parts := strings.Split(topic, "/")
	return parts[len(parts)-1]
}
