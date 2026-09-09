package connector

import (
	"context"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/config"
)

type mockClient struct {
	Client
	config           config.PulsarConfig
	createProducerFn func(pulsar.ProducerOptions) (pulsar.Producer, error)
	subscribeFn      func(pulsar.ConsumerOptions) (pulsar.Consumer, error)
}

func (m *mockClient) GetConfig() config.PulsarConfig {
	return m.config
}

func (m *mockClient) CreateProducer(options pulsar.ProducerOptions) (pulsar.Producer, error) {
	return m.createProducerFn(options)
}

func (m *mockClient) Subscribe(options pulsar.ConsumerOptions) (pulsar.Consumer, error) {
	return m.subscribeFn(options)
}

type mockProducer struct {
	pulsar.Producer
	sendFn func(context.Context, *pulsar.ProducerMessage) (pulsar.MessageID, error)
}

func (m *mockProducer) Send(ctx context.Context, message *pulsar.ProducerMessage) (pulsar.MessageID, error) {
	return m.sendFn(ctx, message)
}

type mockPulsarConsumer struct {
	pulsar.Consumer
	reconsumeLaterFn func(pulsar.Message, time.Duration)
}

func (m *mockPulsarConsumer) ReconsumeLater(message pulsar.Message, delay time.Duration) {
	if m.reconsumeLaterFn != nil {
		m.reconsumeLaterFn(message, delay)
	}
}

type mockMessage struct {
	pulsar.Message
	properties map[string]string
	payload    []byte
}

func (m *mockMessage) Properties() map[string]string {
	return m.properties
}

func (m *mockMessage) Payload() []byte {
	return m.payload
}
