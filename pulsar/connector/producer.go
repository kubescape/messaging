package connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.uber.org/zap"
)

type Producer interface {
	pulsar.Producer
}

type producer struct {
	pulsar.Producer
	//TODO override Send for OTL
}

type createProducerOptions struct {
	Topic     TopicName
	FullTopic string
	Tenant    string
	Namespace string
}

func (opt *createProducerOptions) defaults(client Client) {
	if opt.Tenant == "" {
		opt.Tenant = client.GetConfig().Tenant
	}
	if opt.Namespace == "" {
		opt.Namespace = client.GetConfig().Namespace
	}
}

func (opt *createProducerOptions) validate() error {
	if opt.Topic == "" && opt.FullTopic == "" {
		return fmt.Errorf("topic must be specified")
	}
	if opt.Topic != "" && opt.FullTopic != "" {
		return fmt.Errorf("only one of topic or fullTopic must be specified")
	}
	return nil
}

func WithProducerNamespace(tenant, namespace string) CreateProducerOption {
	return func(o *createProducerOptions) {
		o.Tenant = tenant
		o.Namespace = namespace
	}
}

func WithProducerTopic(topic TopicName) CreateProducerOption {
	return func(o *createProducerOptions) {
		o.Topic = topic
	}
}

func WithProducerFullTopic(topic TopicName) CreateProducerOption {
	return func(o *createProducerOptions) {
		o.FullTopic = string(topic)
	}
}

type CreateProducerOption func(*createProducerOptions)

func newProducer(pulsarClient Client, createProducerOption ...CreateProducerOption) (Producer, error) {
	opts := &createProducerOptions{}
	opts.defaults(pulsarClient)
	for _, o := range createProducerOption {
		o(opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	if opts.FullTopic == "" {
		opts.FullTopic = BuildPersistentTopic(opts.Tenant, opts.Namespace, opts.Topic)
	}
	p, err := pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: opts.FullTopic,
	})
	if err != nil {
		return nil, fmt.Errorf("CreateProducer: failed to create producer: %w", err)
	}
	return producer{Producer: p}, nil

}

type produceMessageOptions struct {
	msgToSend    interface{}
	pulsarClient Client
	ctx          context.Context
	properties   map[string]string
	key          string
}

type ProduceMessageOption func(*produceMessageOptions)

func WithContext(ctx context.Context) ProduceMessageOption {
	return func(o *produceMessageOptions) {
		o.ctx = ctx
	}
}

func WithMessageToSend(msgToSend interface{}) ProduceMessageOption {
	return func(o *produceMessageOptions) {
		o.msgToSend = msgToSend
	}
}

func WithMessageKey(key string) ProduceMessageOption {
	return func(o *produceMessageOptions) {
		o.key = key
	}
}

func WithPulsarClient(pulsarClient Client) ProduceMessageOption {
	return func(o *produceMessageOptions) {
		o.pulsarClient = pulsarClient
	}
}

func WithProperties(properties map[string]string) ProduceMessageOption {
	return func(o *produceMessageOptions) {
		o.properties = properties
	}
}

func ProduceMessage(producer pulsar.Producer, producerOpts ...ProduceMessageOption) error {
	opts := &produceMessageOptions{}
	for _, o := range producerOpts {
		o(opts)
	}
	msgBytes, err := json.Marshal(opts.msgToSend)
	if err != nil {
		return fmt.Errorf("ProduceMessage: failed to marshal message: %w", err)
	}

	msg := pulsar.ProducerMessage{
		Payload:    msgBytes,
		Properties: opts.properties,
		Key:        opts.key,
	}

	if err != nil {
		return fmt.Errorf("ProduceMessage: failed to create producer: %w", err)
	}

	// Publish the mock message to the topic
	if msgID, err := producer.Send(opts.ctx, &msg); err != nil {
		return fmt.Errorf("ProduceMessage: failed to publish message: %w", err)
	} else {
		zap.L().Debug("Published message to pulsar", zap.String("msgId", string(msgID.Serialize())))
	}

	return nil
}
