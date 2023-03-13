package pulsarconnector

import (
	"context"
	_ "embed"
	"os/exec"

	"net/http"
	"testing"
	"time"

	"encoding/json"

	"github.com/kubescape/pulsar-connector/config"

	"github.com/apache/pulsar-client-go/pulsar"

	"github.com/stretchr/testify/suite"
)

const (
	// pulsar 2.11.0
	pulsarDockerCommand = `docker run --name=pulsar-test -d -p 6651:6650  -p 8081:8080 -e PULSAR_MEM=" -Xms256m -Xmx256m -XX:MaxDirectMemorySize=256m" docker.io/apachepulsar/pulsar@sha256:3b755fb67d49abeb7ab6a76b7123cc474375e3881526db26f43c8cfccdaa3cf6 bin/pulsar standalone`
	pulsarStopCommand   = "docker stop pulsar-test && docker rm pulsar-test"
)

func TestBasicConnection(t *testing.T) {
	suite.Run(t, new(MainTestSuite))
}

type MainTestSuite struct {
	suite.Suite
	defaultTestConfig config.PulsarConfig
	shutdownFunc      func()
}

func (suite *MainTestSuite) SetupSuite() {
	suite.defaultTestConfig = config.PulsarConfig{
		URL:                    "pulsar://localhost:6651",
		AdminUrl:               "http://localhost:8081",
		Tenant:                 "ca-messaging",
		Namespace:              "test-namespace",
		Clusters:               []string{"standalone"},
		MaxDeliveryAttempts:    2,
		RedeliveryDelaySeconds: 0,
	}
	//start pulsar
	suite.startPulsar()

	//ensure pulsar connection
	pulsarClient, err := GetClientOnce(&suite.defaultTestConfig)
	if err != nil {
		suite.FailNow("failed to create pulsar client", err.Error())
	}
	suite.shutdownFunc = func() {
		defer func() {
			pulsarClient.Close()
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

func (suite *MainTestSuite) startPulsar() {
	suite.T().Log("stopping existing pulsar container")
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
	suite.T().Log("starting pulsar")
	out, err := exec.Command("/bin/sh", "-c", pulsarDockerCommand).Output()
	if err != nil {
		suite.FailNow("failed to start pulsar", err.Error(), string(out))
	}
	req, err := http.NewRequest(http.MethodGet, suite.defaultTestConfig.AdminUrl+"/admin/v2/brokers/ready", nil)
	if err != nil {
		suite.FailNow("failed to create request", err.Error())
	}
	suite.T().Log("waiting for pulsar to start")
	client := http.Client{}
	for i := 0; i < 30; i++ {
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
func consumeMessages[P TestPayload](suite *MainTestSuite, ctx context.Context, consumer pulsar.Consumer) (actualPayloads map[string]P) {
	//consume payloads for one second
	testConsumerCtx, consumerCancel := context.WithTimeout(ctx, time.Second)
	defer consumerCancel()
	actualPayloads = map[string]P{}
	for {
		msg, err := consumer.Receive(testConsumerCtx)

		if err != nil {
			if testConsumerCtx.Err() == nil {
				suite.FailNow(err.Error(), "consumer error")
			}
			break
		}
		var payload P
		if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
			suite.FailNow(err.Error())
		}
		if _, ok := actualPayloads[payload.GetId()]; ok {
			suite.FailNow("duplicate id")
		}
		actualPayloads[payload.GetId()] = payload
		consumer.Ack(msg)
	}
	return actualPayloads
}

func (suite *MainTestSuite) newDlqConsumer(topicName TopicName) pulsar.Consumer {
	client, err := GetClientOnce(&suite.defaultTestConfig)
	if err != nil {
		suite.FailNow(err.Error())
	}
	topic := GetTopic(topicName + "-dlq")
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topic:            topic,
		SubscriptionName: "Test",
		Type:             pulsar.Shared,
	})
	if err != nil {
		suite.FailNow(err.Error(), "Dlq Consumer creation")
	}
	return consumer
}
