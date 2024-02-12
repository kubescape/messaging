package connector

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"
	"strconv"
	"sync"

	"net/http"
	"testing"
	"time"

	"encoding/json"

	"github.com/kubescape/messaging/pulsar/common/utils"
	"github.com/kubescape/messaging/pulsar/config"
	"golang.org/x/sync/errgroup"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/stretchr/testify/suite"
)

const (
	pulsarClientPort = 6651
	pulsarAdminPort  = 8081
)

var (
	//go:embed scripts/pulsar.sh
	pulsarDockerCommand string
	//go:embed scripts/pulsar_stop.sh
	pulsarStopCommand string
)

func TestBasicConnection(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

type MainTestSuite struct {
	suite.Suite
	failOnUnconsummedMessages bool
	defaultTestConfig         config.PulsarConfig
	pulsarClient              Client
	shutdownFunc              func()
}

func (suite *MainTestSuite) SetupSuite() {
	suite.defaultTestConfig = config.PulsarConfig{
		URL:                    fmt.Sprintf("pulsar://localhost:%d", pulsarClientPort),
		AdminUrl:               fmt.Sprintf("http://localhost:%d", pulsarAdminPort),
		Tenant:                 "ca-messaging",
		Namespace:              "test-namespace",
		Clusters:               []string{"standalone"},
		MaxDeliveryAttempts:    2,
		RedeliveryDelaySeconds: 0,
	}

	x, _ := json.Marshal(suite.defaultTestConfig)
	fmt.Println(string(x))

	//start pulsar
	suite.startPulsar()

	var err error
	//ensure pulsar connection
	suite.pulsarClient, err = NewClient(WithConfig(&suite.defaultTestConfig))
	if err != nil {
		suite.FailNow("failed to create pulsar client", err.Error())
	}
	suite.shutdownFunc = func() {
		defer func() {
			suite.pulsarClient.Close()
		}()
	}
}

func (suite *MainTestSuite) TearDownSuite() {
	suite.T().Log("tear down suite")
	suite.shutdownFunc()
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf(pulsarStopCommand, "basic-suite")).CombinedOutput()
	if err != nil {
		suite.FailNow("failed to stop pulsar", err.Error(), string(out))
	}
}

func (suite *MainTestSuite) SetupTest() {
	// suite.T().Log("setup test")
}

func (suite *MainTestSuite) TearDownTest() {
	//cleanup unconsumed messages
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*200)
	defer cancel()
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, 0),
	}
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	dlqConsumer, err := CreateTestDlqConsumer(suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer dlqConsumer.Close()
	errg := errgroup.Group{}
	errg.Go(func() error {
		for {
			msg, err := consumer.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				} else {
					return err
				}
			}
			if err := consumer.Ack(msg); err != nil {
				suite.FailNow(err.Error())
			}
			if suite.failOnUnconsummedMessages {
				return fmt.Errorf("unconsumed message topic:%s \npayload:%s \nproperites:%v", msg.Topic(), msg.Payload(), msg.Properties())
			}
		}
	})
	errg.Go(func() error {
		for {
			msg, err := dlqConsumer.Receive(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return nil
				} else {
					return err
				}
			}
			if err := dlqConsumer.Ack(msg); err != nil {
				suite.FailNow(err.Error())
			}
			if suite.failOnUnconsummedMessages {
				return fmt.Errorf("unconsumed message topic:%s \npayload:%s \nproperites:%v", msg.Topic(), msg.Payload(), msg.Properties())
			}
		}
	})
	if err := errg.Wait(); err != nil {
		suite.FailNow(err.Error())
	}
	suite.failOnUnconsummedMessages = false
}

func (suite *MainTestSuite) TestCreateConsumer() {
	//create consumer
	chan1 := make(chan pulsar.ConsumerMessage)

	consumer, err := suite.pulsarClient.NewConsumer(WithMessageChannel(chan1), WithTopic("test-topic"), WithSubscriptionName("test-subscription"))
	if err != nil {
		suite.FailNow("failed to create consumer", err.Error())
	}
	defer consumer.Close()
}

func (suite *MainTestSuite) TestCreateProducer() {
	//create producer
	producer, err := suite.pulsarClient.NewProducer(WithProducerTopic("test-topic"))
	if err != nil {
		suite.FailNow("failed to create producer", err.Error())
	}
	defer producer.Close()
}

func (suite *MainTestSuite) TestCreateProducerFullTopic() {
	//create producer
	producer, err := suite.pulsarClient.NewProducer(WithProducerFullTopic("persistent://ca-messaging/test-namespace/test-topic"))
	if err != nil {
		suite.FailNow("failed to create producer", err.Error())
	}
	defer producer.Close()
}

func (suite *MainTestSuite) TestCreateProducerFullTopicNonPersistent() {
	//create producer
	BuildNonPersistentTopic("test-t", "test-ns", "test-topic")
	producer, err := suite.pulsarClient.NewProducer(WithProducerFullTopic("non-persistent://ca-messaging/test-namespace/test-topic"))
	if err != nil {
		suite.FailNow("failed to create producer", err.Error())
	}
	defer producer.Close()
}

func (suite *MainTestSuite) TestCreateProducerFullTopicInvalid() {
	//create producer
	producer, err := suite.pulsarClient.NewProducer(WithProducerFullTopic("est-t/test-ns/test-topic"))
	if err == nil {
		suite.FailNow("created invalid topic producer")
	}
	if producer != nil {
		defer producer.Close()
	}
}

func (suite *MainTestSuite) TestProduceMessage() {
	//create producer
	producer, err := suite.pulsarClient.NewProducer(WithProducerTopic("test-topic"))
	if err != nil {
		suite.FailNow("failed to create producer", err.Error())
	}
	defer producer.Close()

	//produce message
	msg := "test message"
	if err := ProduceMessage(producer, WithMessageToSend(msg), WithContext(context.Background())); err != nil {
		suite.FailNow("failed to produce message", err.Error())
	}
}

func (suite *MainTestSuite) startPulsar() {
	suite.T().Log("stopping existing pulsar container")
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
	time.Sleep(2 * time.Second)

	suite.T().Log("starting pulsar")
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf(pulsarDockerCommand, pulsarClientPort, pulsarAdminPort, "basic-suite")).CombinedOutput()
	if err != nil {
		suite.FailNow("failed to start pulsar", err.Error(), string(out))
	}
	req, err := http.NewRequest(http.MethodGet, suite.defaultTestConfig.AdminUrl+"/admin/v2/brokers/ready", nil)
	if err != nil {
		suite.FailNow("failed to create request", err.Error())
	}
	suite.T().Log("waiting for pulsar to start")
	client := http.Client{}
	for i := 0; i < 40; i++ {
		resp, err := client.Do(req)
		if err == nil && resp.StatusCode == http.StatusOK {
			suite.T().Log("pulsar started")
			resp.Body.Close()
			return
		}
		time.Sleep(2 * time.Second)
	}
	suite.FailNow("failed to start pulsar")
}

func loadJson[T any](jsonBytes []byte) T {
	var obj T
	if err := json.Unmarshal(jsonBytes, &obj); err != nil {
		panic(err)
	}
	return obj
}

func produceMessages[P TestPayload](suite *MainTestSuite, ctx context.Context, producer pulsar.Producer, payloads []P) {
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

func consumeMessages[P TestPayload](suite *MainTestSuite, ctx context.Context, consumer pulsar.Consumer, consumerId string, timeoutSeconds int) (actualPayloads map[string]P) {
	//consume payloads for X seconds
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(timeoutSeconds))
	defer consumerCancel()
	actualPayloads = map[string]P{}
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
			fmt.Printf("%s: unmarshal failed - Nack()", consumerId)
			consumer.Nack(msg)
			continue
		}
		if _, ok := actualPayloads[payload.GetId()]; ok {
			fmt.Printf("%s: duplicate ID %s", consumerId, payload.GetId())
			suite.FailNow("duplicate id")

		}
		actualPayloads[payload.GetId()] = payload
		fmt.Printf("%s: Ack() - ID %s", consumerId, payload.GetId())
		if err := consumer.Ack(msg); err != nil {
			suite.FailNow(err.Error())
		}
	}
	return actualPayloads
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
}

func (suite *MainTestSuite) TestReconsumeLaterWithNacks() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, 0),
	}
	//create consumer to get actual payloads
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	dlqConsumer, err := CreateTestDlqConsumer(suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer dlqConsumer.Close()
	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(time.Second*2))
	defer consumerCancel()
	msg, err := consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	//reconsume
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	//reconsume again
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	suite.False(consumer.IsReconsumable(msg), "expect message not to be reconsumable")
	consumer.Nack(msg)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	//nack again - 2nd redelivery
	consumer.Nack(msg)
	//expect dlq
	msg, err = dlqConsumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	if err := dlqConsumer.Ack(msg); err != nil {
		suite.FailNow(err.Error())
	}
	testMsg(msg)
	//fail on test teardown if there are unconsumed messages
	suite.failOnUnconsummedMessages = true

}

func (suite *MainTestSuite) TestReconsumeLaterMaxAttemps() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, 0),
	}
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	dlqConsumer, err := CreateTestDlqConsumer(suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer dlqConsumer.Close()
	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(time.Second*2))
	defer consumerCancel()
	msg, err := consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	//reconsume
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	//reconsume again
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	suite.False(consumer.IsReconsumable(msg), "expect message not to be reconsumable")
	//reconsume again - this time it should go to dlq
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	//expect dlq
	msg, err = dlqConsumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	if err := dlqConsumer.Ack(msg); err != nil {
		suite.FailNow(err.Error())
	}
	testMsg(msg)
	//fail on test teardown if there are unconsumed messages
	suite.failOnUnconsummedMessages = true

}

func (suite *MainTestSuite) TestReconsumeLater() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, 0),
	}
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(time.Second*2))
	defer consumerCancel()
	msg, err := consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	//reconsume
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	//reconsume again
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	msg, err = consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	testMsg(msg)
	suite.False(consumer.IsReconsumable(msg), "expect message not to be reconsumable")
	if err := consumer.Ack(msg); err != nil {
		suite.FailNow(err.Error())
	}
	//fail on test teardown if there are unconsumed messages
	suite.failOnUnconsummedMessages = true

}

func (suite *MainTestSuite) TestReconsumeLaterWithDuration() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, time.Second*2),
	}
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()

	dlqConsumer, err := CreateTestDlqConsumer(suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer dlqConsumer.Close()

	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(time.Second*2))
	defer consumerCancel()
	msg, err := consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	//reconsume 20 times
	for i := 0; i < 20; i++ {
		consumer.ReconsumeLater(msg, time.Millisecond*5)
		msg, err = consumer.Receive(testConsumerCtx)
		if err != nil {
			suite.FailNow(err.Error(), "reconsume payload")
		}
		testMsg(msg)
	}
	time.Sleep(time.Millisecond * 2300)
	suite.False(consumer.IsReconsumable(msg), "expect message not to be reconsumable")
	//reconsume anyway
	consumer.ReconsumeLater(msg, time.Millisecond*5)
	//expect dlq
	msg, err = dlqConsumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "reconsume payload")
	}
	if err := dlqConsumer.Ack(msg); err != nil {
		suite.FailNow(err.Error())
	}
	testMsg(msg)

	//fail on test teardown if there are unconsumed messages
	suite.failOnUnconsummedMessages = true
}

func (suite *MainTestSuite) TestSafeReconsumeLaterWithDuration() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	//create consumer to get actual payloads
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, false, time.Second*2),
	}
	consumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer consumer.Close()
	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(2))
	defer consumerCancel()
	msg, err := consumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	//reconsume 20 times
	for i := 0; i < 20; i++ {
		sent := consumer.ReconsumeLaterDLQSafe(msg, time.Millisecond)
		if !sent {
			suite.FailNow("expected to message to be reconsumed got false in meesage num ", i)
		}
		msg, err = consumer.Receive(testConsumerCtx)
		if err != nil {
			suite.FailNow(err.Error(), "reconsume payload")
		}
		testMsg(msg)
	}
	time.Sleep(time.Millisecond * 2300)
	suite.False(consumer.IsReconsumable(msg), "expect message not to be reconsumable")
	//reconsume anyway
	sent := consumer.ReconsumeLaterDLQSafe(msg, time.Millisecond*5)
	suite.False(sent, "expect reconsume to send message")
	savedRetries, _ := strconv.Atoi(msg.Properties()[propertySavedReconsumeAttempts])
	retries, _ := strconv.Atoi(msg.Properties()[pulsar.SysPropertyReconsumeTimes])
	suite.Equal(20, savedRetries+retries, "expected 20 retries")
	if err := consumer.Ack(msg); err != nil {
		suite.FailNow(err.Error())
	}
	//fail on test teardown if there are unconsumed messages
	suite.failOnUnconsummedMessages = true
}

func (suite *MainTestSuite) TestReconsumeLaterPanicOnRetryDisabled() {
	defer func() {
		if r := recover(); r == nil {
			suite.T().Errorf("The code did not panic")
		} else {
			suite.T().Logf("Recovered in TestReconsumeLaterPanics: %v", r)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()

	noRetrayEnabledConsumer, err := CreateTestConsumer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer noRetrayEnabledConsumer.Close()
	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(2))
	defer consumerCancel()
	msg, err := noRetrayEnabledConsumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	noRetrayEnabledConsumer.ReconsumeLater(msg, time.Millisecond)
	suite.FailNow("should panic on call reconsume when retry option is not enabled")
}

func (suite *MainTestSuite) TestReconsumeLaterPanicOnUnSafeReconsume() {
	defer func() {
		if r := recover(); r == nil {
			suite.T().Errorf("The code did not panic")
		} else {
			suite.T().Logf("Recovered in TestReconsumeLaterPanicOnUnSafeReconsume: %v", r)
		}
	}()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	producer, err := CreateTestProducer(ctx, suite.pulsarClient)
	if err != nil {
		suite.FailNow(err.Error(), "create producer")
	}
	if producer == nil {
		suite.FailNow("producer is nil")
	}
	defer producer.Close()
	reconsumeLaterOptions := []CreateConsumerOption{
		WithRetryEnable(true, true, time.Second*2),
	}
	safeOnlyEnabledConsumer, err := CreateTestConsumer(ctx, suite.pulsarClient, reconsumeLaterOptions...)
	if err != nil {
		suite.FailNow(err.Error())
	}
	defer safeOnlyEnabledConsumer.Close()
	//produce
	if _, err := producer.Send(ctx, &pulsar.ProducerMessage{Payload: []byte(suite.T().Name())}); err != nil {
		suite.FailNow(err.Error(), "send payload")
	}
	testMsg := func(msg pulsar.Message) {
		if msg == nil {
			suite.FailNow("msg is nil")
		}
		if string(msg.Payload()) != suite.T().Name() {
			suite.FailNow("unexpected payload")
		}
	}
	//consume
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second*time.Duration(2))
	defer consumerCancel()
	msg, err := safeOnlyEnabledConsumer.Receive(testConsumerCtx)
	if err != nil {
		suite.FailNow(err.Error(), "receive payload")
	}
	testMsg(msg)
	safeOnlyEnabledConsumer.ReconsumeLater(msg, time.Millisecond)
	suite.FailNow("should panic on call ReconsumeLater when safe only option is set")
}
