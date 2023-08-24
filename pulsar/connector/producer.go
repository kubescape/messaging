package connector

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/apache/pulsar-client-go/pulsar"
	"go.uber.org/zap"
)

func CreateProducer(pulsarClient pulsar.Client, topic TopicName) (pulsar.Producer, error) {
	// Get a producer instance
	producer, err := pulsarClient.CreateProducer(pulsar.ProducerOptions{
		Topic: GetTopic(topic),
	})
	if err != nil {
		return nil, fmt.Errorf("CreateProducer: failed to create producer: %w", err)
	}
	return producer, nil
}

type produceMessageOptions struct {
	msgToSend    interface{}
	pulsarClient pulsar.Client
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

func WithPulsarClient(pulsarClient pulsar.Client) ProduceMessageOption {
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
