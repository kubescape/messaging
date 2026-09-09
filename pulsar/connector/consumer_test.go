package connector

import (
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConsumerBuildsOptions(t *testing.T) {
	messageChannel := make(chan pulsar.ConsumerMessage)
	tests := []struct {
		name            string
		options         []CreateConsumerOption
		wantTopic       string
		wantTopics      []string
		wantQueueSize   int
		wantMessageChan chan pulsar.ConsumerMessage
	}{
		{
			name: "single topic",
			options: []CreateConsumerOption{
				WithTopic("test-topic"),
				WithSubscriptionName("test-subscription"),
				WithMessageChannel(messageChannel),
			},
			wantTopic:       "persistent://ca-messaging/test-namespace/test-topic",
			wantMessageChan: messageChannel,
		},
		{
			name: "full topics",
			options: []CreateConsumerOption{
				WithFullTopics([]TopicName{"persistent://other-tenant/other-namespace/test-topic"}),
				WithSubscriptionName("test-subscription"),
			},
			wantTopics: []string{"persistent://other-tenant/other-namespace/test-topic"},
		},
		{
			name: "receiver queue size",
			options: []CreateConsumerOption{
				WithTopic("test-topic"),
				WithSubscriptionName("test-subscription"),
				WithQueueSize(10),
			},
			wantTopic:     "persistent://ca-messaging/test-namespace/test-topic",
			wantQueueSize: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotOptions pulsar.ConsumerOptions
			client := &mockClient{
				config: config.PulsarConfig{
					Tenant:              "ca-messaging",
					Namespace:           "test-namespace",
					MaxDeliveryAttempts: 2,
				},
				subscribeFn: func(options pulsar.ConsumerOptions) (pulsar.Consumer, error) {
					gotOptions = options
					return &mockPulsarConsumer{}, nil
				},
			}

			created, err := newConsumer(client, tt.options...)
			require.NoError(t, err)
			require.NotNil(t, created)
			assert.Equal(t, tt.wantTopic, gotOptions.Topic)
			if tt.wantTopics == nil {
				assert.Empty(t, gotOptions.Topics)
			} else {
				assert.Equal(t, tt.wantTopics, gotOptions.Topics)
			}
			assert.Equal(t, "test-subscription", gotOptions.SubscriptionName)
			assert.Equal(t, pulsar.Shared, gotOptions.Type)
			assert.Equal(t, tt.wantQueueSize, gotOptions.ReceiverQueueSize)
			assert.Equal(t, tt.wantMessageChan, gotOptions.MessageChannel)
			require.NotNil(t, gotOptions.DLQ)
			assert.Equal(t, uint32(2), gotOptions.DLQ.MaxDeliveries)
		})
	}
}

func TestConsumerReconsumeLaterPanicsWhenRetryDisabled(t *testing.T) {
	wrapped := consumer{
		Consumer: &mockPulsarConsumer{},
		options:  createConsumerOptions{},
	}

	assert.PanicsWithValue(
		t,
		"reconsumeLater called on consumer without retry enabled option set to true",
		func() { wrapped.ReconsumeLater(&mockMessage{properties: map[string]string{}}, time.Millisecond) },
	)
}

func TestConsumerReconsumeLaterPanicsWhenSafeRetryRequired(t *testing.T) {
	wrapped := consumer{
		Consumer: &mockPulsarConsumer{},
		options: createConsumerOptions{
			retryEnabled:      true,
			forceDLQSafeRetry: true,
		},
	}

	assert.PanicsWithValue(
		t,
		"reconsumeLater: when forceDLQSafeRetry option is true ReconsumeLaterDLQSafe must be used",
		func() { wrapped.ReconsumeLater(&mockMessage{properties: map[string]string{}}, time.Millisecond) },
	)
}
