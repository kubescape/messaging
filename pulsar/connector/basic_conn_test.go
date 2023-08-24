package connector

import (
	"context"
	_ "embed"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/messaging/pulsar/common/utils"
)

type TestPayload interface {
	GetId() string
}

type TestPayloadImpl struct {
	Id   string `json:"id"`
	Data string `json:"data"`
}

type TestInvalidPayloadImpl struct {
	Id   int    `json:"id"`
	Data string `json:"data"`
}

type TestPayloadImplInterface struct {
	Id   interface{} `json:"id"`
	Data string      `json:"data"`
}

func (p TestInvalidPayloadImpl) GetId() string {
	return fmt.Sprintf("%d", p.Id)
}

func (p TestPayloadImpl) GetId() string {
	return p.Id
}

func (p TestPayloadImplInterface) GetId() string {
	return fmt.Sprintf("%v", p.Id)
}

const TestTopicName TopicName = "test-topic"
const TestSubscriptionName = "test-consumer"
const testProducerName = "test-producer"

func (suite *MainTestSuite) TestConsumerAndProducer() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//create producer to input test payloads
	pubsubCtx := utils.NewContextWithValues(ctx, "testConsumer")
	producer, err := CreateTestProducer(pubsubCtx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()

	//create consumer to get actual payloads
	consumer, err := CreateTestConsumer(pubsubCtx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	testPayload := []byte("[{\"id\":\"1\",\"data\":\"Hello World\"},{\"id\":\"18\",\"data\":\"Hello from the other World\"}]")

	//send test payloads
	produceMessages(suite, ctx, producer, loadJson[[]TestPayloadImpl](testPayload))
	//consume payloads for one second
	actualPayloads := consumeMessages[TestPayloadImpl](suite, pubsubCtx, consumer, "", 1)

	suite.Equal(2, len(actualPayloads), "expected 2 messages")
}

// CreateTestProducer creates a producer
func CreateTestProducer(ctx context.Context, client PulsarClient) (pulsar.Producer, error) {

	producer, err := CreateProducer(client, WithProducerTopic(TestTopicName))

	if err != nil && utils.IsProducerNameExistsError(testProducerName, err) {
		//other instance became the producer
		return nil, nil
	}

	if producer != nil && err == nil {
		utils.SetContextProducer(ctx, producer)
	}
	return producer, err
}

func CreateTestConsumer(ctx context.Context, client PulsarClient) (pulsar.Consumer, error) {
	return CreateSharedConsumer(client,
		WithTopic(TestTopicName),
		WithSubscriptionName(TestSubscriptionName),
		WithRedeliveryDelay(time.Duration(client.GetConfig().RedeliveryDelaySeconds)*time.Second),
		WithDLQ(uint32(client.GetConfig().MaxDeliveryAttempts)),
		WithDefaultBackoffPolicy(),
	)

}

func CreateTestDlqConsumer(client PulsarClient) (pulsar.Consumer, error) {
	return CreateSharedConsumer(client,
		WithTopic(TestTopicName),
		WithSubscriptionName(TestSubscriptionName+"-dlq"),
		WithRedeliveryDelay(0),
		WithDLQ(0),
	)
}

func (suite *MainTestSuite) TestDLQ() {
	//start tenant check
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//create producer to input test payloads
	pubsubCtx := utils.NewContextWithValues(ctx, "testConsumer")
	producer, err := CreateTestProducer(pubsubCtx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	consumer, err := CreateTestConsumer(pubsubCtx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()
	dlqConsumer, err := CreateTestDlqConsumer(suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer dlqConsumer.Close()

	testPayload := []byte("[{\"id\":\"1\",\"data\":\"Hello World\"},{\"id\":2,\"data\":\"Hello from the other World\"}]")

	//send test payloads
	produceMessages(suite, ctx, producer, loadJson[[]TestPayloadImplInterface](testPayload))
	//sleep to allow redelivery
	time.Sleep(time.Second * 5)
	//create next stage consumer and dlq consumer
	wg := sync.WaitGroup{}
	wg.Add(2)
	var actualPayloads map[string]TestPayloadImpl
	go func() {
		defer wg.Done()
		// consume payloads for one second
		actualPayloads = consumeMessages[TestPayloadImpl](suite, pubsubCtx, consumer, "consumer", 20)
		//sleep to allow redelivery
		//
	}()
	var dlqPayloads map[string]TestInvalidPayloadImpl
	go func() {
		defer wg.Done()
		time.Sleep(time.Second * 10)
		// consume payloads for one second
		dlqPayloads = consumeMessages[TestInvalidPayloadImpl](suite, pubsubCtx, dlqConsumer, "dlqConsumer", 20)
	}()
	wg.Wait()

	suite.Equal(1, len(actualPayloads), "expected 1 msg in successful consumer")
	suite.Contains(actualPayloads, "1", "expected msg with ID 1 in successful consumer")

	suite.Equal(1, len(dlqPayloads), "expected 1 msg in dlq consumer")
	suite.Contains(dlqPayloads, "2", "expected msg with ID 2 in dlq consumer")

	// TODO: bad payload test
	// suite.badPayloadTest(ctx, producer, dlqConsumer)
}

func (suite *MainTestSuite) badPayloadTest(ctx context.Context, producer pulsar.Producer, dlqConsumer pulsar.Consumer) {
	actualPayloads := []string{}
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte("some bad payload")}); err != nil {
		suite.FailNow(err.Error())
	}

	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*6)
	defer consumerCancel()

	for {
		msg, err := dlqConsumer.Receive(testConsumerCtx)
		if err != nil {
			if testConsumerCtx.Err() == nil {
				suite.FailNow(err.Error(), "consumer error")
			}
			break
		}
		actualPayloads = append(actualPayloads, string(msg.Payload()))
		dlqConsumer.Ack(msg)
	}

	suite.Equal(1, len(actualPayloads), "expected 1 payload in dlq")
	if len(actualPayloads) == 1 {
		suite.Equal("some bad payload", actualPayloads[0], "expected payload to be some bad payload")
	}
}
