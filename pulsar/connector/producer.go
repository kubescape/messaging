package connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.uber.org/zap"
)

type createProducerOptions struct {
	Topic     TopicName
	Tenant    string
	Namespace string
}

func (opt *createProducerOptions) defaults(client PulsarClient) {
	if opt.Tenant == "" {
		opt.Tenant = client.GetConfig().Tenant
	}
	if opt.Namespace == "" {
		opt.Namespace = client.GetConfig().Namespace
	}
}

func (opt *createProducerOptions) validate() error {
	if opt.Topic == "" {
		return fmt.Errorf("topic must be specified")
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

type CreateProducerOption func(*createProducerOptions)

func CreateProducer(pulsarClient PulsarClient, createProducerOption ...CreateProducerOption) (pulsar.Producer, error) {
	opts := &createProducerOptions{}
	opts.defaults(pulsarClient)
	for _, o := range createProducerOption {
		o(opts)
	}
	if err := opts.validate(); err != nil {
		return nil, err
	}
	producer, err := pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: BuildPersistentTopic(opts.Tenant, opts.Namespace, opts.Topic),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateProducer: failed to create producer: %w", err)
	}
	return producer, nil
}

type produceMessageOptions struct {
	msgToSend    interface{}
	pulsarClient PulsarClient
	ctx          context.Context
	properties   map[string]string
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

func WithPulsarClient(pulsarClient PulsarClient) ProduceMessageOption {
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
