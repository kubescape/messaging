package pulsarconnector

import (
	"context"
	_ "embed"
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

func (p *TestPayloadImpl) GetId() string {
	return p.Id
}

const TestTopicName = "test-topic"
const TestSubscriptionName = "test-consumer"

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

	testPayload := []byte("{\"id\":\"1\",\"data\":\"Hello World\"}")

	//send test payloads
	produceMessages(suite, ctx, producer, loadJson[[]TestPayload](testPayload))
	//consume payloads for one second
	actualPayloads := consumeMessages[*TestPayloadImpl](suite, pubsubCtx, consumer)
	//expect payloads of 2 tenants (one tenant one unsubscribed user)
	suite.Equal(len(actualPayloads), 2, "expected 2 tenants")
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
		Interceptors: tracer.NewProducerInterceptors(ctx),
	})
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
