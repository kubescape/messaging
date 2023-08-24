package test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/connector"
)

func ProduceToTopic(suite *PulsarTestSuite, topic connector.TopicName, payloads [][]byte) {
	producer, err := connector.CreateProducer(suite.PulsarClient, topic)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	defer producer.Close()
	for _, payload := range payloads {
		if _, err := producer.Send(context.Background(), &pulsar.ProducerMessage{Payload: payload}); err != nil {
			suite.FailNow(err.Error(), "send payload")
		}
	}
}

func ProduceObjectsToTopic[P any](suite *PulsarTestSuite, topic connector.TopicName, payloads []P) {
	producer, err := connector.CreateProducer(suite.PulsarClient, topic)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	defer producer.Close()
	ProduceMessages[P](suite, context.Background(), producer, payloads)
}

func ProduceMessages[P any](suite *PulsarTestSuite, ctx context.Context, producer pulsar.Producer, payloads []P) {
	for _, payload := range payloads {
		payloadBytes, err := json.Marshal(payload)
		if err != nil {
			suite.FailNow(err.Error(), "marshal payload")
		}
		if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: payloadBytes}); err != nil {
			suite.FailNow(err.Error(), "send payload")
		}
	}
}

// SubscribeToTopic - subscribe to a topic and returns a function that consumes messages from the topic
func SubscribeToTopic[P any](suite *PulsarTestSuite, topic connector.TopicName, payloads []P, subscription string) (consumeFunc func() []P) {
	consumer, err := connector.CreateSharedConsumer(suite.PulsarClient, connector.WithTopic(topic), connector.WithSubscriptionName(subscription))
	if err != nil {
		suite.FailNow(err.Error(), "subscribe")
	}
	return func() []P {
		defer consumer.Close()
		return ConsumeMessages[P](suite, context.Background(), consumer, subscription, 1)
	}
}

func ConsumeMessages[P any](suite *PulsarTestSuite, ctx context.Context, consumer pulsar.Consumer, consumerId string, timeoutSeconds int) (actualPayloads []P) {
	//consume payloads for X seconds
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(timeoutSeconds))
	defer consumerCancel()
	actualPayloads = []P{}
	for {
		msg, err := consumer.Receive(testConsumerCtx)

		if err != nil {
			if testConsumerCtx.Err() == nil {
				fmt.Printf("%s: consumer error: %s", consumerId, err.Error())
				suite.FailNow(err.Error(), "consumer error")
			}
			fmt.Printf("%s: breaking - %s", consumerId, err.Error())
			break
		}
		var payload P
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			suite.NoError(err, "unmarshal failed")
			consumer.Nack(msg)
			continue
		}
		actualPayloads = append(actualPayloads, payload)
		consumer.Ack(msg)
	}
	return actualPayloads
}
