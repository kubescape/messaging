package pulsarconnector

import (
	"context"
	_ "embed"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/kubescape/pulsar-connector/common/tracer"
	"github.com/kubescape/pulsar-connector/common/utils"
	"github.com/kubescape/pulsar-connector/config"
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

const TestTopicName = "test-topic"
const TestSubscriptionName = "test-consumer"
const testProducerName = "test-producer"

func (suite *MainTestSuite) TestConsumerAndProducer() {
	testConf := suite.defaultTestConfig
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//create producer to input test payloads
	pubsubCtx := utils.NewContextWithValues(ctx, "testConsumer")
	producer, err := CreateTestProducer(pubsubCtx, &testConf)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()

	//create consumer to get actual payloads
	consumer, err := CreateTestConsumer(pubsubCtx, &testConf)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	testPayload := []byte("[{\"id\":\"1\",\"data\":\"Hello World\"},{\"id\":\"18\",\"data\":\"Hello from the other World\"}]")

	//send test payloads
	produceMessages(suite, ctx, producer, loadJson[[]TestPayloadImpl](testPayload))
	//consume payloads for one second
	actualPayloads := consumeMessages[TestPayloadImpl](suite, pubsubCtx, consumer)
	//expect payloads of 2 tenants (one tenant one unsubscribed user)
	suite.Equal(len(actualPayloads), 2, "expected 2 messages")
	//compare with expected payloads
	// compareAndUpdate(suite.T(), actualPayloads, expectedRecipientsPayloads, expectedRecipientsPayloadsFile, updatedExpected, ignoreTime)
}

// CreateTestProducer creates a producer
func CreateTestProducer(ctx context.Context, config *config.PulsarConfig) (pulsar.Producer, error) {
	client, err := GetClientOnce(config)
	if err != nil {
		return nil, err
	}

	producer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:        GetTopic(TestTopicName),
		Name:         testProducerName,
		Interceptors: tracer.NewProducerInterceptors(ctx),
	})

	if err != nil && IsProducerNameExistsError(testProducerName, err) {
		//other instance became the producer
		return nil, nil
	}

	if producer != nil && err == nil {
		utils.SetContextProducer(ctx, producer)
	}
	return producer, err
}

func CreateTestConsumer(ctx context.Context, config *config.PulsarConfig) (pulsar.Consumer, error) {
	client, err := GetClientOnce(config)
	if err != nil {
		return nil, err
	}

	topic := GetTopic(TestTopicName)
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:                          topic,
		SubscriptionName:               TestSubscriptionName,
		Type:                           pulsar.Shared,
		DLQ:                            NewDlq(topic, ctx),
		Interceptors:                   tracer.NewConsumerInterceptors(ctx),
		NackRedeliveryDelay:            time.Duration(config.RedeliveryDelaySeconds) * time.Second,
		EnableDefaultNackBackoffPolicy: true,
	})
	if consumer != nil && err == nil {
		utils.SetContextConsumer(ctx, consumer)
	}
	return consumer, err
}

func (suite *MainTestSuite) TestDLQ() {
	//set test config
	testConf := suite.defaultTestConfig
	//start tenant check
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	//create producer to input test payloads
	pubsubCtx := utils.NewContextWithValues(ctx, "testConsumer")
	producer, err := CreateTestProducer(pubsubCtx, &testConf)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	consumer, err := CreateTestConsumer(pubsubCtx, &testConf)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()
	dlqConsumer := suite.newDlqConsumer(TestTopicName)
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
		//consume payloads for one second
		actualPayloads = consumeMessages[TestPayloadImpl](suite, pubsubCtx, consumer)
	}()
	var dlqPayloads map[string]TestInvalidPayloadImpl
	go func() {
		defer wg.Done()
		//consume payloads for one second
		dlqPayloads = consumeMessages[TestInvalidPayloadImpl](suite, pubsubCtx, dlqConsumer)
	}()
	wg.Wait()

	//expect payloads of 2 tenants (one tenant one unsubscribed user)
	suite.Equal(len(actualPayloads), 1, "expected 1 msg in successful consumer")
	suite.Equal(len(dlqPayloads), 1, "expected 1 msg in dlq consumer")

	//compare with expected payloads
	// compareAndUpdate(suite.T(), dlqPayloads, dlqRecipientsPayloads, dlqRecipientsPayloadsFile, updatedExpected, ignoreTime)

	//do bad payload test
	suite.badPayloadTest(ctx, producer, dlqConsumer)
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
