package connector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProducerBuildsTopic(t *testing.T) {
	tests := []struct {
		name      string
		options   []CreateProducerOption
		wantTopic string
	}{
		{
			name:      "default namespace",
			options:   []CreateProducerOption{WithProducerTopic("test-topic")},
			wantTopic: "persistent://ca-messaging/test-namespace/test-topic",
		},
		{
			name: "overridden namespace",
			options: []CreateProducerOption{
				WithProducerTopic("test-topic"),
				WithProducerNamespace("other-tenant", "other-namespace"),
			},
			wantTopic: "persistent://other-tenant/other-namespace/test-topic",
		},
		{
			name:      "full persistent topic",
			options:   []CreateProducerOption{WithProducerFullTopic("persistent://tenant/namespace/topic")},
			wantTopic: "persistent://tenant/namespace/topic",
		},
		{
			name:      "full non-persistent topic",
			options:   []CreateProducerOption{WithProducerFullTopic("non-persistent://tenant/namespace/topic")},
			wantTopic: "non-persistent://tenant/namespace/topic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOptions pulsar.ProducerOptions
			pulsarProducer := &mockProducer{}
			client := &mockClient{
				config: config.PulsarConfig{Tenant: "ca-messaging", Namespace: "test-namespace"},
				createProducerFn: func(options pulsar.ProducerOptions) (pulsar.Producer, error) {
					gotOptions = options
					return pulsarProducer, nil
				},
			}

			created, err := newProducer(client, tt.options...)
			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, tt.wantTopic, gotOptions.Topic)
		})
	}
}

func TestNewProducerWrapsClientError(t *testing.T) {
	client := &mockClient{
		config: config.PulsarConfig{Tenant: "tenant", Namespace: "namespace"},
		createProducerFn: func(pulsar.ProducerOptions) (pulsar.Producer, error) {
			return nil, errors.New("invalid topic")
		},
	}

	producer, err := newProducer(client, WithProducerFullTopic("invalid-topic"))

	require.Nil(t, producer)
	require.ErrorContains(t, err, "CreateProducer: failed to create producer: invalid topic")
}

func TestProduceMessageForwardsOptions(t *testing.T) {
	type contextKey struct{}
	ctx := context.WithValue(context.Background(), contextKey{}, "value")
	properties := map[string]string{"source": "unit-test"}
	delay := 250 * time.Millisecond
	var gotContext context.Context
	var gotMessage *pulsar.ProducerMessage
	producer := &mockProducer{
		sendFn: func(ctx context.Context, message *pulsar.ProducerMessage) (pulsar.MessageID, error) {
			gotContext = ctx
			gotMessage = message
			return pulsar.NewMessageID(1, 2, 3, 4), nil
		},
	}

	err := ProduceMessage(
		producer,
		WithContext(ctx),
		WithMessageToSend("test message"),
		WithMessageKey("message-key"),
		WithProperties(properties),
		WithDelay(delay),
	)

	require.NoError(t, err)
	assert.Same(t, ctx, gotContext)
	require.NotNil(t, gotMessage)
	assert.JSONEq(t, `"test message"`, string(gotMessage.Payload))
	assert.Equal(t, properties, gotMessage.Properties)
	assert.Equal(t, "message-key", gotMessage.Key)
	assert.Equal(t, delay, gotMessage.DeliverAfter)
}
