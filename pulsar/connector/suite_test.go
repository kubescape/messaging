package connector

import (
	"context"
	_ "embed"
	"fmt"
	"os/exec"

	"net/http"
	"testing"
	"time"

	"encoding/json"

	"github.com/kubescape/messaging/pulsar/config"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/stretchr/testify/suite"
)

const (
	pulsarClientPort    = "6651"
	pulsarAdminPort     = "8081"
	pulsarDockerCommand = `docker run --name=pulsar-test -d -p %s:6650  -p %s:8080 docker.io/apachepulsar/pulsar:2.11.0 bin/pulsar standalone`
	pulsarStopCommand   = "docker stop pulsar-test && docker rm pulsar-test"
)

func TestBasicConnection(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

type MainTestSuite struct {
	suite.Suite
	defaultTestConfig config.PulsarConfig
	pulsarClient      Client
	shutdownFunc      func()
}

func (suite *MainTestSuite) SetupSuite() {
	suite.defaultTestConfig = config.PulsarConfig{
		URL:                    fmt.Sprintf("pulsar://localhost:%s", pulsarClientPort),
		AdminUrl:               fmt.Sprintf("http://localhost:%s", pulsarAdminPort),
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
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()

}

func (suite *MainTestSuite) SetupTest() {
	// suite.T().Log("setup test")
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
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf(pulsarDockerCommand, pulsarClientPort, pulsarAdminPort)).Output()
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
		consumer.Ack(msg)
	}
	return actualPayloads
}
