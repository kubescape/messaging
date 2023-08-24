package connector

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

type createConsumerOptions struct {
	Topic                TopicName
	Topics               []TopicName
	SubscriptionName     string
	MaxDeliveryAttempts  uint32
	RedeliveryDelay      time.Duration
	MessageChannel       chan pulsar.ConsumerMessage
	DefaultBackoffPolicy bool
	BackoffPolicy        pulsar.NackBackoffPolicy
}

func (pot *createConsumerOptions) defaults() {
	if pot.MaxDeliveryAttempts == 0 {
		pot.MaxDeliveryAttempts = uint32(GetClientConfig().MaxDeliveryAttempts)
	}
	if pot.RedeliveryDelay == 0 {
		pot.RedeliveryDelay = time.Duration(GetClientConfig().RedeliveryDelaySeconds)
	}
}

func (opt *createConsumerOptions) validate() error {
	if opt.Topic == "" && len(opt.Topics) == 0 {
		return fmt.Errorf("topic or topics must be specified")
	}
	if opt.Topic != "" && len(opt.Topics) != 0 {
		return fmt.Errorf("cannot specify both topic and topics")
	}
	if opt.SubscriptionName == "" {
		return fmt.Errorf("subscription name must be specified")
	}
	if opt.DefaultBackoffPolicy && opt.BackoffPolicy != nil {
		return fmt.Errorf("cannot specify both default backoff policy and backoff policy")
	}
	return nil
}

type CreateConsumerOption func(*createConsumerOptions)

func WithRedeliveryDelay(redeliveryDelay time.Duration) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.RedeliveryDelay = redeliveryDelay
	}
}

func WithBackoffPolicy(backoffPolicy pulsar.NackBackoffPolicy) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.BackoffPolicy = backoffPolicy
	}
}

func WithDefaultBackoffPolicy() CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.DefaultBackoffPolicy = true
	}
}

// maxDeliveryAttempts before sending to DLQ - 0 means no DLQ
// by default, maxDeliveryAttempts is 5
func WithDLQ(maxDeliveryAttempts uint32) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.MaxDeliveryAttempts = maxDeliveryAttempts
	}
}

func WithTopic(topic TopicName) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.Topic = topic
	}
}

func WithTopics(topics []TopicName) CreateConsumerOption {
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

func CreateSharedConsumer(pulsarClient pulsar.Client, createConsumerOpts ...CreateConsumerOption) (pulsar.Consumer, error) {
	opts := &createConsumerOptions{}
	opts.defaults()
	for _, o := range createConsumerOpts {
		o(opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	var dlq *pulsar.DLQPolicy
	if opts.MaxDeliveryAttempts != 0 {
		dlq = NewDlq(opts.Topic, opts.MaxDeliveryAttempts)
	}
	var topic string
	var topics []string
	if opts.Topic != "" {
		topic = GetTopic(opts.Topic)
	} else {
		topics = make([]string, len(opts.Topics))
		for i, t := range opts.Topics {
			topics[i] = GetTopic(t)
		}
	}
	return pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:                          topic,
		Topics:                         topics,
		SubscriptionName:               opts.SubscriptionName,
		Type:                           pulsar.Shared,
		MessageChannel:                 opts.MessageChannel,
		DLQ:                            dlq,
		EnableDefaultNackBackoffPolicy: opts.DefaultBackoffPolicy,

		//	Interceptors:        tracer.NewConsumerInterceptors(ctx),
		NackRedeliveryDelay: opts.RedeliveryDelay,
		NackBackoffPolicy:   opts.BackoffPolicy,
	})

}
