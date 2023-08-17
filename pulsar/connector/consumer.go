package connector

import (
	"github.com/apache/pulsar-client-go/pulsar"
)

type createConsumerOptions struct {
	Topic            string
	Topics           []string
	SubscriptionName string
	MessageChannel   chan pulsar.ConsumerMessage
}

type CreateConsumerOption func(*createConsumerOptions)

func WithTopic(topic string) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.Topic = topic
	}
}

func WithTopics(topics []string) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.Topics = topics
	}
}

func WithSubscriptionName(subscriptionName string) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.SubscriptionName = subscriptionName
	}
}

func WithMessageChannel(messageChannel chan pulsar.ConsumerMessage) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.MessageChannel = messageChannel
	}
}

func CreateConsumer(pulsarClient pulsar.Client, createConsumerOpts ...CreateConsumerOption) (pulsar.Consumer, error) {
	opts := &createConsumerOptions{}
	for _, o := range createConsumerOpts {
		o(opts)
	}

	consumer, err := pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:            opts.Topic,
		Topics:           opts.Topics,
		SubscriptionName: opts.SubscriptionName,
		Type:             pulsar.Shared,
		MessageChannel:   opts.MessageChannel,
	})

	return consumer, err
}
