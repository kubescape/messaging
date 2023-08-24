package connector

import (
	"testing"
	"time"
)

func TestNackBackoffPolicy(t *testing.T) {
	tests := []struct {
		name                  string
		minMultiplier         uint32
		maxMultiplier         uint32
		baseDelay             time.Duration
		redeliveryCount       uint32
		expectedDelay         time.Duration
		expectedValidationErr bool
	}{
		{
			name:                  "Base delay 100ms, Redelivery count 1",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             100 * time.Millisecond,
			redeliveryCount:       1,
			expectedDelay:         200 * time.Millisecond,
			expectedValidationErr: false,
		},
		{
			name:                  "Base delay 1s, Redelivery count 2",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             1 * time.Second,
			redeliveryCount:       2,
			expectedDelay:         3 * time.Second,
			expectedValidationErr: false,
		},
		{
			name:                  "Base delay 1s, Redelivery count 5",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             1 * time.Second,
			redeliveryCount:       5,
			expectedDelay:         5 * time.Second,
			expectedValidationErr: false,
		},
		// Test cases targeting maxMultiplier
		{
			name:                  "Base delay 100ms, Redelivery count at maxMultiplier",
			minMultiplier:         1,
			maxMultiplier:         3,
			baseDelay:             100 * time.Millisecond,
			redeliveryCount:       3,
			expectedDelay:         300 * time.Millisecond,
			expectedValidationErr: false,
		},

		{
			name:                  "Base delay 1s, Redelivery count at maxMultiplier",
			minMultiplier:         1,
			maxMultiplier:         4,
			baseDelay:             1 * time.Second,
			redeliveryCount:       4,
			expectedDelay:         4 * time.Second,
			expectedValidationErr: false,
		},
		{
			name:                  "Base delay 1s, Redelivery count beyond maxMultiplier",
			minMultiplier:         1,
			maxMultiplier:         2,
			baseDelay:             1 * time.Second,
			redeliveryCount:       3,
			expectedDelay:         2 * time.Second, // Should cap at maxMultiplier
			expectedValidationErr: false,
		},
		// Additional test cases
		{
			name:                  "Base delay 100ms, Redelivery count 0",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             100 * time.Millisecond,
			redeliveryCount:       0,
			expectedDelay:         100 * time.Millisecond,
			expectedValidationErr: false,
		},
		{
			name:                  "Base delay 1s, Redelivery count at maxMultiplier",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             1 * time.Second,
			redeliveryCount:       5,
			expectedDelay:         5 * time.Second,
			expectedValidationErr: false,
		},
		{
			name:                  "Base delay 1s, Redelivery count beyond maxMultiplier",
			minMultiplier:         1,
			maxMultiplier:         5,
			baseDelay:             1 * time.Second,
			redeliveryCount:       6,
			expectedDelay:         5 * time.Second,
			expectedValidationErr: false,
		},
		{
			name:                  "Invalid configuration: minMultiplier > maxMultiplier",
			minMultiplier:         6,
			maxMultiplier:         5,
			baseDelay:             1 * time.Second,
			redeliveryCount:       1,
			expectedValidationErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nbp, err := NewNackBackoffPolicy(tt.minMultiplier, tt.maxMultiplier, tt.baseDelay)
			if (err != nil) != tt.expectedValidationErr {
				t.Fatalf("Expected validation error: %v, got: %v", tt.expectedValidationErr, err)
			}
			if err == nil {
				delay := nbp.Next(tt.redeliveryCount)
				if delay != tt.expectedDelay {
					t.Fatalf("Expected delay: %v, got: %v", tt.expectedDelay, delay)
				}
			}
		})
	}
}
