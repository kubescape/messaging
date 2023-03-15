package connector

import (
	"context"

	"github.com/kubescape/messaging/pulsar/common/tracer"

	"github.com/apache/pulsar-client-go/pulsar"
)

func NewDlq(topic string, ctx context.Context) *pulsar.DLQPolicy {
	return &pulsar.DLQPolicy{
		MaxDeliveries:   uint32(GetClientConfig().MaxDeliveryAttempts),
		DeadLetterTopic: topic + "-dlq",
		ProducerOptions: pulsar.ProducerOptions{
			Interceptors: tracer.NewProducerInterceptors(ctx),
		},
	}
}
