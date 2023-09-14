package test

import (
	_ "embed"
	"errors"
	"fmt"
	"net"
	"os/exec"

	"net/http"
	"time"

	"encoding/json"

	"github.com/kubescape/messaging/pulsar/config"
	"github.com/kubescape/messaging/pulsar/connector"

	"github.com/stretchr/testify/suite"
)

const (
	pulsarKAURL = "%s/admin/v2/brokers/ready"
)

//go:embed scripts/pulsar.sh
var startPulsarScript string

//go:embed scripts/pulsar_stop.sh
var pulsarStopCommand string

type PulsarTestSuite struct {
	suite.Suite
	DefaultTestConfig config.PulsarConfig
	Client            connector.Client
	AppPortStart      int
	AdminPortStart    int
	shutdownFunc      func()
}

func (suite *PulsarTestSuite) SetupSuite() {
	suite.T().Log("setup suite")
	suite.DefaultTestConfig = config.PulsarConfig{
		Tenant:                 "ca-messaging",
		Namespace:              "test-namespace",
		Clusters:               []string{"standalone"},
		MaxDeliveryAttempts:    2,
		RedeliveryDelaySeconds: 0,
	}

	randomContainerName := fmt.Sprintf("pulsar-test-%d", time.Now().UnixNano())
	if suite.AppPortStart == 0 {
		suite.AppPortStart = 6650
	}
	if suite.AdminPortStart == 0 {
		suite.AdminPortStart = 8080
	}
	//start pulsar
	suite.startPulsar(randomContainerName)

	x, _ := json.Marshal(suite.DefaultTestConfig)
	fmt.Println(string(x))
	var err error
	//ensure pulsar connection
	suite.Client, err = connector.NewClient(connector.WithConfig(&suite.DefaultTestConfig))
	if err != nil {
		suite.FailNow("failed to create pulsar client", err.Error())
	}
	suite.shutdownFunc = func() {
		defer func() {
			suite.Client.Close()
			formmatedScript := fmt.Sprintf(pulsarStopCommand, randomContainerName)
			outbytes, err := exec.Command("/bin/sh", "-c", formmatedScript).CombinedOutput()
			if err != nil {
				suite.FailNow("failed to stop pulsar", err.Error(), string(outbytes))
			}
		}()
	}
}
func (suite *PulsarTestSuite) checkPulsarIsAlive() bool {
	kaURL := fmt.Sprintf(pulsarKAURL, suite.DefaultTestConfig.AdminUrl)
	req, err := http.NewRequest(http.MethodGet, kaURL, nil)
	if err != nil {
		suite.FailNow("failed to create request", err.Error())
	}
	client := http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err == nil && resp.StatusCode == http.StatusOK {
		suite.T().Log("pulsar started")
		resp.Body.Close()
		return true
	}
	return false
}
func (suite *PulsarTestSuite) TearDownSuite() {
	suite.T().Log("tear down suite")
	suite.shutdownFunc()
	suite.Assert().NoError(killPortProcess(suite.AppPortStart))
	suite.Assert().NoError(killPortProcess(suite.AdminPortStart))
}

func (suite *PulsarTestSuite) SetupTest() {
	suite.T().Log("setup test")
}

func (suite *PulsarTestSuite) TearDownTest() {
	suite.T().Log("tear down test")
	// clear all pulsar topics messages
	if err := suite.clearAllMessages(); err != nil {
		suite.FailNow("failed to clear all messages", err.Error())
	}
}

func findFreePort(rangeStart, rangeEnd int) (int, error) {
	for port := rangeStart; port <= rangeEnd; port++ {
		address := fmt.Sprintf("localhost:%d", port)
		conn, err := net.DialTimeout("tcp", address, 1*time.Second)
		if conn != nil {
			conn.Close()
		}
		if err != nil { // port is available since we got no response
			return port, nil
		}
		conn.Close()
	}
	return 0, errors.New("no free port found")
}

func (suite *PulsarTestSuite) startPulsar(contName string) {
	suite.T().Log("stopping existing pulsar container")
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
	suite.T().Log("starting pulsar")

	pulsarAppPort, err := findFreePort(suite.AppPortStart, suite.AppPortStart+100)
	if err != nil {
		suite.FailNow("failed to find free port", err.Error())
	}
	suite.DefaultTestConfig.URL = fmt.Sprintf("pulsar://localhost:%d", pulsarAppPort)
	suite.AppPortStart = pulsarAppPort
	pulsarAdminPort, err := findFreePort(suite.AdminPortStart, suite.AdminPortStart+100)
	if err != nil {
		suite.FailNow("failed to find free port for pulsar admin", err.Error())
	}
	suite.DefaultTestConfig.AdminUrl = fmt.Sprintf("http://localhost:%d", pulsarAdminPort)
	suite.AdminPortStart = pulsarAdminPort

	formattedScript := fmt.Sprintf(startPulsarScript, pulsarAppPort, pulsarAdminPort, contName)
	out, err := exec.Command("/bin/sh", "-c", formattedScript).Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			suite.FailNow("failed to start pulsar", err.Error(), string(exitErr.Stderr), string(out))
		}
		suite.FailNow("failed to start pulsar", err.Error(), string(out))
	}
	suite.T().Log("waiting for pulsar to start")
	for i := 0; i < 30; i++ {
		isAlive := suite.checkPulsarIsAlive()
		if isAlive {
			return
		}
		time.Sleep(2 * time.Second)
	}
	suite.FailNow("failed to start pulsar")
}
