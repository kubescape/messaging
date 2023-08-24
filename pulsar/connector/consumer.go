package connector

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/config"
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
	Tenant               string
	Namespace            string
}

func (opt *createConsumerOptions) defaults(config config.PulsarConfig) {
	if opt.MaxDeliveryAttempts == 0 {
		opt.MaxDeliveryAttempts = uint32(config.MaxDeliveryAttempts)
	}
	if opt.RedeliveryDelay == 0 {
		opt.RedeliveryDelay = time.Duration(config.RedeliveryDelaySeconds)
	}
	if opt.Tenant == "" {
		opt.Tenant = config.Tenant
	}
	if opt.Namespace == "" {
		opt.Namespace = config.Namespace
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

func WithNamespace(tenant, namespace string) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.Tenant = tenant
		o.Namespace = namespace
	}
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

func CreateSharedConsumer(pulsarClient PulsarClient, createConsumerOpts ...CreateConsumerOption) (pulsar.Consumer, error) {
	opts := &createConsumerOptions{}
	opts.defaults(pulsarClient.GetConfig())
	for _, o := range createConsumerOpts {
		o(opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	var dlq *pulsar.DLQPolicy
	if opts.MaxDeliveryAttempts != 0 {
		dlq = NewDlq(opts.Tenant, opts.Namespace, opts.Topic, opts.MaxDeliveryAttempts)
	}
	var topic string
	var topics []string
	if opts.Topic != "" {
		topic = BuildPersistentTopic(opts.Tenant, opts.Namespace, opts.Topic)
	} else {
		topics = make([]string, len(opts.Topics))
		for i, t := range opts.Topics {
			topics[i] = BuildPersistentTopic(opts.Tenant, opts.Namespace, t)
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
