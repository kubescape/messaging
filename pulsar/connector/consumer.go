package connector

import (
	"fmt"
	"strconv"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/config"
)

type Consumer interface {
	pulsar.Consumer
	//ReconsumeLaterDLQSafe returns false if the message is not Reconsumable (i.e. out of redeliveries or out of time. next attempt will go to dlq)
	//If the message was sent for reconsuming the return value is true
	ReconsumeLaterDLQSafe(msg pulsar.Message, delay time.Duration) bool
	//IsReconsumable returns true if the message can be reconsumed or false if next attempt will go to dlq
	IsReconsumable(msg pulsar.Message) bool
}

type consumer struct {
	pulsar.Consumer
	options createConsumerOptions
}

const (
	propertyRetryStartTime         = "RETRY_START_TIME"
	propertySavedReconsumeAttempts = "SAVED_RECONSUME_ATTEMPS"
)

func (c consumer) ReconsumeLater(msg pulsar.Message, delay time.Duration) {
	if !c.options.retryEnabled {
		panic("reconsumeLater called on consumer without retry enabled option set to true")
	}
	if c.options.forceDLQSafeRetry {
		panic("reconsumeLater: when forceDLQSafeRetry option is true ReconsumeLaterDLQSafe must be used")
	}
	c.applyReconsumeDurationProperties(msg)
	c.Consumer.ReconsumeLater(msg, delay)
}

func (c consumer) IsReconsumable(msg pulsar.Message) bool {
	if !c.options.retryEnabled {
		return false
	}
	c.applyReconsumeDurationProperties(msg)
	reconsumeTimes := 1
	if msg.Properties() != nil {
		if s, ok := msg.Properties()[pulsar.SysPropertyReconsumeTimes]; ok {
			reconsumeTimes, _ = strconv.Atoi(s)
			reconsumeTimes++
		}
	}
	return reconsumeTimes <= int(c.options.MaxDeliveryAttempts)
}

func (c consumer) ReconsumeLaterDLQSafe(msg pulsar.Message, delay time.Duration) bool {
	if !c.IsReconsumable(msg) {
		return false
	}
	c.Consumer.ReconsumeLater(msg, delay)
	return true
}

func (c *consumer) applyReconsumeDurationProperties(msg pulsar.Message) {
	if !c.options.retryEnabled {
		return
	}
	if c.options.retryDuration > 0 && msg.Properties() != nil {
		//get/set startTime
		startTime := time.Now()
		if _, ok := msg.Properties()[propertyRetryStartTime]; !ok {
			msg.Properties()[propertyRetryStartTime] = startTime.Format(time.RFC3339)
		} else {
			startTime, _ = time.Parse(time.RFC3339, msg.Properties()[propertyRetryStartTime])
		}
		//get reconsume Times
		reconsumeTimes := 1
		if s, ok := msg.Properties()[pulsar.SysPropertyReconsumeTimes]; ok {
			reconsumeTimes, _ = strconv.Atoi(s)
			reconsumeTimes++

		}
		//get actual retries
		actualRetries := 0
		if s, ok := msg.Properties()[propertySavedReconsumeAttempts]; ok {
			actualRetries, _ = strconv.Atoi(s)
		}
		//if next delivery is about to exceed max deliveries (and nack) and the duration has not passed yet
		if reconsumeTimes > int(c.options.MaxDeliveryAttempts) && time.Since(startTime) < c.options.retryDuration {
			//reset the reconsume times to 1
			msg.Properties()[pulsar.SysPropertyReconsumeTimes] = "1"
			//reduce the actual retries by 1
			actualRetries--
			//add the retry count the saved attempts property
			msg.Properties()[propertySavedReconsumeAttempts] = strconv.Itoa(actualRetries + reconsumeTimes - 1) //reduce one becuase it was not reconsumed yet
			//reset the reconsume times to 1
			msg.Properties()[pulsar.SysPropertyReconsumeTimes] = "1"

		} else if reconsumeTimes <= int(c.options.MaxDeliveryAttempts) && time.Since(startTime) >= c.options.retryDuration {
			//duration passed set the retry to MaxDeliveryAttempts
			msg.Properties()[propertySavedReconsumeAttempts] = strconv.Itoa(actualRetries + reconsumeTimes)
			msg.Properties()[pulsar.SysPropertyReconsumeTimes] = strconv.Itoa(int(c.options.MaxDeliveryAttempts))
		}
	}
}

type createConsumerOptions struct {
	Topic                TopicName
	Topics               []TopicName
	FullTopics           []TopicName
	SubscriptionName     string
	subscriptionType     *pulsar.SubscriptionType
	MaxDeliveryAttempts  uint32
	dlqNamespace         string
	RedeliveryDelay      time.Duration
	MessageChannel       chan pulsar.ConsumerMessage
	DefaultBackoffPolicy bool
	BackoffPolicy        pulsar.NackBackoffPolicy
	Tenant               string
	Namespace            string
	//retry options
	retryEnabled bool
	//duration of retry (overides the max delivery attemps)
	retryDuration time.Duration
	//safe retry with no sending to DLQ when set must use ReconsumeLaterDLQSafe and not ReconsumeLater
	forceDLQSafeRetry bool
	//automatically set to <namepace>-retry/<topic>-retry
	retryTopic string
}

func (opt *createConsumerOptions) defaults(config config.PulsarConfig) {
	if opt.MaxDeliveryAttempts == 0 {
		opt.MaxDeliveryAttempts = uint32(config.MaxDeliveryAttempts)
	}
	if opt.RedeliveryDelay == 0 {
		opt.RedeliveryDelay = time.Second * time.Duration(config.RedeliveryDelaySeconds)
	}
	if opt.Tenant == "" {
		opt.Tenant = config.Tenant
	}
	if opt.Namespace == "" {
		opt.Namespace = config.Namespace
	}
	if opt.dlqNamespace == "" {
		opt.dlqNamespace = opt.Namespace + dlqNamespaceSuffix
	}
}

func (opt *createConsumerOptions) validate() error {
	if opt.Topic == "" && len(opt.Topics) == 0 && len(opt.FullTopics) == 0 {
		return fmt.Errorf("topic or topics must be specified")
	}
	if opt.Topic != "" && len(opt.Topics) != 0 {
		return fmt.Errorf("cannot specify both topic and topics")
	}
	if opt.Topic != "" && len(opt.FullTopics) != 0 {
		return fmt.Errorf("cannot specify both topic and fullTopics")
	}
	if opt.SubscriptionName == "" {
		return fmt.Errorf("subscription name must be specified")
	}
	if opt.DefaultBackoffPolicy && opt.BackoffPolicy != nil {
		return fmt.Errorf("cannot specify both default backoff policy and backoff policy")
	}
	if opt.MaxDeliveryAttempts == 0 && opt.retryEnabled {
		return fmt.Errorf("cannot enable retry without setting max delivery attempts")
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

func WithRetryEnable(enable, forceDLQSafeRetry bool, retryDuration time.Duration) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.retryEnabled = enable
		o.retryDuration = retryDuration
		o.forceDLQSafeRetry = forceDLQSafeRetry
	}
}

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

func WithSubscriptionType(subscriptionType pulsar.SubscriptionType) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.subscriptionType = &subscriptionType
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

func WithFullTopics(topics []TopicName) CreateConsumerOption {
	return func(o *createConsumerOptions) {
		o.FullTopics = topics
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

func newConsumer(pulsarClient Client, createConsumerOpts ...CreateConsumerOption) (Consumer, error) {
	opts := &createConsumerOptions{}
	opts.defaults(pulsarClient.GetConfig())
	for _, o := range createConsumerOpts {
		o(opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}

	var topic string
	topics := make([]string, 0, len(opts.Topics)+len(opts.FullTopics))

	if opts.Topic != "" {
		topic = BuildPersistentTopic(opts.Tenant, opts.Namespace, opts.Topic)
	} else {
		for _, topic := range opts.Topics {
			persistentTopic := BuildPersistentTopic(opts.Tenant, opts.Namespace, topic)
			topics = append(topics, persistentTopic)
		}
		for _, fullTopic := range opts.FullTopics {
			topics = append(topics, string(fullTopic))
		}
	}
	var dlq *pulsar.DLQPolicy
	if opts.MaxDeliveryAttempts != 0 {
		topicName := opts.Topic
		if topicName == "" && len(opts.Topics) > 0 {
			topicName = opts.Topics[0]
		}

		if opts.retryEnabled {
			opts.retryTopic = BuildPersistentTopic(opts.Tenant, opts.Namespace+retryNamespaceSuffix, topicName+"-retry")
		}
		dlq = NewDlq(opts.Tenant, opts.dlqNamespace, topicName, opts.MaxDeliveryAttempts, opts.retryTopic)
	}

	var subscriptionType pulsar.SubscriptionType
	if opts.subscriptionType != nil {
		subscriptionType = *opts.subscriptionType
	} else {
		// default to shared subscription type if not specified
		subscriptionType = pulsar.Shared
	}

	pulsarConsumer, err := pulsarClient.Subscribe(pulsar.ConsumerOptions{
		Topic:                          topic,
		Topics:                         topics,
		SubscriptionName:               opts.SubscriptionName,
		Type:                           subscriptionType,
		MessageChannel:                 opts.MessageChannel,
		DLQ:                            dlq,
		EnableDefaultNackBackoffPolicy: opts.DefaultBackoffPolicy,
		RetryEnable:                    opts.retryEnabled,
		//	Interceptors:        tracer.NewConsumerInterceptors(ctx),
		NackRedeliveryDelay: opts.RedeliveryDelay,
		NackBackoffPolicy:   opts.BackoffPolicy,
	})
	if err != nil {
		return nil, err
	}
	return consumer{pulsarConsumer, *opts}, nil

}
