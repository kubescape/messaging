package pulsarconnector

import (
	"context"
	_ "embed"
	"os"

	"net/http"
	"os/exec"
	"testing"
	"time"

	"encoding/json"

	"github.com/kubescape/pulsar-connector/config"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/go-cmp/cmp"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	pulsarDockerCommand = `docker run --name=pulsar-test -d -p 6651:6650  -p 8081:8080 -e PULSAR_MEM=" -Xms256m -Xmx256m -XX:MaxDirectMemorySize=256m" apachepulsar/pulsar-all bin/pulsar standalone`
	pulsarStopCommand   = "docker stop pulsar-test && docker rm pulsar-test"
)

var updatedExpected = false

func TestNotificationService(t *testing.T) {
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

func compareAndUpdate[T any](t *testing.T, actual T, expectedBytes []byte, expectedFileName string, update bool, compareOptions ...cmp.Option) {
	expected := loadJson[T](expectedBytes)
	diff := cmp.Diff(expected, actual, compareOptions...)
	assert.Empty(t, diff, "expected to have no diff")
	if update && diff != "" {
		saveExpected(t, expectedFileName, actual)
	}
	assert.False(t, update, "update expected is true, set to false and rerun test")
}

func saveExpected(t *testing.T, fileName string, i interface{}) {
	data, _ := json.MarshalIndent(i, "", "    ")
	err := os.WriteFile(fileName, data, 0644)
	if err != nil {
		panic(err)
	}
	t.Log("Updating expected file: "+fileName, " with actual response: ", string(data))
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
