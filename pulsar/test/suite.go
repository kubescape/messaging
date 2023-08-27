package test

import (
	_ "embed"
	"fmt"
	"os/exec"

	"net/http"
	"time"

	"encoding/json"

	"github.com/kubescape/messaging/pulsar/config"
	"github.com/kubescape/messaging/pulsar/connector"

	"github.com/stretchr/testify/suite"
)

const (
	pulsarClientPort    = "6651"
	pulsarAdminPort     = "8081"
	pulsarDockerCommand = `docker run --name=pulsar-test -d -p %s:6650  -p %s:8080 docker.io/apachepulsar/pulsar:2.11.0 bin/pulsar standalone`
	pulsarStopCommand   = "docker stop pulsar-test && docker rm pulsar-test"
)

type PulsarTestSuite struct {
	suite.Suite
	DefaultTestConfig config.PulsarConfig
	Client            connector.Client
	shutdownFunc      func()
}

func (suite *PulsarTestSuite) SetupSuite() {
	suite.DefaultTestConfig = config.PulsarConfig{
		URL:                    fmt.Sprintf("pulsar://localhost:%s", pulsarClientPort),
		AdminUrl:               fmt.Sprintf("http://localhost:%s", pulsarAdminPort),
		Tenant:                 "ca-messaging",
		Namespace:              "test-namespace",
		Clusters:               []string{"standalone"},
		MaxDeliveryAttempts:    2,
		RedeliveryDelaySeconds: 0,
	}

	x, _ := json.Marshal(suite.DefaultTestConfig)
	fmt.Println(string(x))

	//start pulsar
	suite.startPulsar()

	var err error
	//ensure pulsar connection
	suite.Client, err = connector.NewClient(connector.WithConfig(&suite.DefaultTestConfig))
	if err != nil {
		suite.FailNow("failed to create pulsar client", err.Error())
	}
	suite.shutdownFunc = func() {
		defer func() {
			suite.Client.Close()
		}()
	}
}

func (suite *PulsarTestSuite) TearDownSuite() {
	suite.T().Log("tear down suite")
	suite.shutdownFunc()
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
}

func (suite *PulsarTestSuite) SetupTest() {
	suite.T().Log("setup test")
}

func (suite *PulsarTestSuite) startPulsar() {
	suite.T().Log("stopping existing pulsar container")
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
	time.Sleep(2 * time.Second)

	suite.T().Log("starting pulsar")
	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf(pulsarDockerCommand, pulsarClientPort, pulsarAdminPort)).Output()
	if err != nil {
		suite.FailNow("failed to start pulsar", err.Error(), string(out))
	}
	req, err := http.NewRequest(http.MethodGet, suite.DefaultTestConfig.AdminUrl+"/admin/v2/brokers/ready", nil)
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
