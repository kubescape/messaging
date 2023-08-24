package connector

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func NewDlq(topic TopicName, maxDeliveryAttempts uint32) *pulsar.DLQPolicy {
	return &pulsar.DLQPolicy{
		MaxDeliveries:   maxDeliveryAttempts,
		DeadLetterTopic: GetTopic(topic + "-dlq"),
		ProducerOptions: pulsar.ProducerOptions{
			//TODO: OTL
			//		Interceptors: tracer.NewProducerInterceptors(ctx),
		},
	}
}

// NackBackoffPolicy implements pulsar's NackBackoffPolicy interface
// NackBackoffPolicy is a redelivery backoff mechanism which we can achieve redelivery with different
// delays according to the number of times the message is retried.
type NackBackoffPolicy struct {
	// minRedeliveryDelayMultiplier specifies the minimum multiplier for the base delay before a message is redelivered.
	minRedeliveryDelayMultiplier uint32
	// maxRedeliveryDelayMultiplier specifies the maximum multiplier for the base delay for redelivery.
	maxRedeliveryDelayMultiplier uint32
	// baseDelay represents the base delay unit for redelivery.
	baseDelay time.Duration
}

// NewNackBackoffPolicy creates a new NackBackoffPolicy or returns an error if the parameters are invalid
func NewNackBackoffPolicy(minRedeliveryDelayMultiplier, maxRedeliveryDelayMultiplier uint32, baseDelay time.Duration) (*NackBackoffPolicy, error) {
	np := &NackBackoffPolicy{
		minRedeliveryDelayMultiplier: minRedeliveryDelayMultiplier,
		maxRedeliveryDelayMultiplier: maxRedeliveryDelayMultiplier,
		baseDelay:                    baseDelay,
	}
	if err := np.Validate(); err != nil {
		return nil, err
	}
	return np, nil
}

// Validate checks the consistency of the backoff policy parameters.
func (nbp *NackBackoffPolicy) Validate() error {
	if nbp.minRedeliveryDelayMultiplier > nbp.maxRedeliveryDelayMultiplier {
		return fmt.Errorf("minRedeliveryDelayMultiplier cannot be greater than maxRedeliveryDelayMultiplier")
	}
	if nbp.baseDelay < time.Millisecond*100 {
		return fmt.Errorf("baseDelay cannot be less than 100ms")
	}
	return nil
}

// Next determines the next redelivery delay based on the given redelivery count.
func (nbp *NackBackoffPolicy) Next(redeliveryCount uint32) time.Duration {
	delay := nbp.calculateIncrementalDelay(redeliveryCount)
	maxDelay := nbp.calculateMaxDelay()
	if delay > maxDelay {
		return maxDelay
	}
	return delay
}

// calculateIncrementalDelay computes the delay for the next redelivery based on the redelivery count.
func (nbp *NackBackoffPolicy) calculateIncrementalDelay(redeliveryCount uint32) time.Duration {
	redeliveryCount++
	if nbp.minRedeliveryDelayMultiplier > 0 {
		return time.Duration(nbp.minRedeliveryDelayMultiplier*redeliveryCount) * nbp.baseDelay
	}
	return time.Duration(redeliveryCount) * nbp.baseDelay
}

// calculateMaxDelay determines the maximum allowed delay for redelivery.
func (nbp *NackBackoffPolicy) calculateMaxDelay() time.Duration {
	if nbp.maxRedeliveryDelayMultiplier > 0 && nbp.maxRedeliveryDelayMultiplier > nbp.minRedeliveryDelayMultiplier {
		return time.Duration(nbp.maxRedeliveryDelayMultiplier) * nbp.baseDelay
	}
	return nbp.calculateIncrementalDelay(0)
}
