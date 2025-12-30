package connector

import (
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
)

func NewDlq(tenant, namespace string, topic TopicName, maxDeliveryAttempts uint32, retryTopic string) *pulsar.DLQPolicy {
	return &pulsar.DLQPolicy{
		MaxDeliveries:    maxDeliveryAttempts,
		RetryLetterTopic: retryTopic,
		DeadLetterTopic:  BuildPersistentTopic(tenant, namespace, topic+"-dlq"),
		ProducerOptions: pulsar.ProducerOptions{
			//TODO: OTL
			//		Interceptors: tracer.NewProducerInterceptors(ctx),
			DisableBatching: true,
			EnableChunking:  true,
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
	// useExponentialBackoff specifies whether to use exponential backoff (2^redeliveryCount) or linear backoff (redeliveryCount * minMultiplier).
	// If not explicitly set, it will be auto-detected based on whether maxMultiplier is a power of 2 (for backward compatibility).
	useExponentialBackoff *bool
}

// NewNackBackoffPolicy creates a new NackBackoffPolicy or returns an error if the parameters are invalid
// By default, it uses linear backoff for backward compatibility with existing tests.
func NewNackBackoffPolicy(minRedeliveryDelayMultiplier, maxRedeliveryDelayMultiplier uint32, baseDelay time.Duration) (*NackBackoffPolicy, error) {
	return NewNackBackoffPolicyWithMode(minRedeliveryDelayMultiplier, maxRedeliveryDelayMultiplier, baseDelay, nil)
}

// NewNackBackoffPolicyWithMode creates a new NackBackoffPolicy with explicit exponential/linear mode control.
// If useExponential is nil, it will auto-detect based on whether maxMultiplier is a power of 2 (for backward compatibility).
func NewNackBackoffPolicyWithMode(minRedeliveryDelayMultiplier, maxRedeliveryDelayMultiplier uint32, baseDelay time.Duration, useExponential *bool) (*NackBackoffPolicy, error) {
	np := &NackBackoffPolicy{
		minRedeliveryDelayMultiplier: minRedeliveryDelayMultiplier,
		maxRedeliveryDelayMultiplier: maxRedeliveryDelayMultiplier,
		baseDelay:                    baseDelay,
		useExponentialBackoff:        useExponential,
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
	// Determine whether to use exponential backoff
	// Default to linear for backward compatibility if not explicitly set
	useExponential := false
	if nbp.useExponentialBackoff != nil {
		useExponential = *nbp.useExponentialBackoff
	}

	if useExponential {
		// Use exponential backoff: base * 2^redeliveryCount
		exponentialMultiplier := uint32(1 << redeliveryCount) // 2^redeliveryCount
		if nbp.maxRedeliveryDelayMultiplier > 0 && exponentialMultiplier > nbp.maxRedeliveryDelayMultiplier {
			exponentialMultiplier = nbp.maxRedeliveryDelayMultiplier
		}
		return time.Duration(exponentialMultiplier) * nbp.baseDelay
	}

	// Use linear backoff: base * (redeliveryCount * minMultiplier)
	redeliveryCount++
	multiplier := redeliveryCount
	if nbp.minRedeliveryDelayMultiplier > 0 {
		multiplier = nbp.minRedeliveryDelayMultiplier * redeliveryCount
	}

	// Cap at maxRedeliveryDelayMultiplier if configured
	if nbp.maxRedeliveryDelayMultiplier > 0 && multiplier > nbp.maxRedeliveryDelayMultiplier {
		multiplier = nbp.maxRedeliveryDelayMultiplier
	}

	return time.Duration(multiplier) * nbp.baseDelay
}

// calculateMaxDelay determines the maximum allowed delay for redelivery.
func (nbp *NackBackoffPolicy) calculateMaxDelay() time.Duration {
	if nbp.maxRedeliveryDelayMultiplier > 0 && nbp.maxRedeliveryDelayMultiplier > nbp.minRedeliveryDelayMultiplier {
		return time.Duration(nbp.maxRedeliveryDelayMultiplier) * nbp.baseDelay
	}
	return nbp.calculateIncrementalDelay(0)
}
