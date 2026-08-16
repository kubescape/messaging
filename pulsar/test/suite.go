package test

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"net/http"
	"time"

	"encoding/json"

	"github.com/kubescape/messaging/pulsar/config"
	"github.com/kubescape/messaging/pulsar/connector"

	"github.com/stretchr/testify/suite"
)

const (
	pulsarKAURL = "%s/admin/v2/brokers/ready"

	// Pulsar now runs with --network=host (see startPulsar/pulsar.sh) so
	// its ports are fixed at the image's built-in defaults rather than
	// remappable per-container.
	pulsarBrokerPort = 6650
	pulsarAdminPort  = 8080

	// System-wide lock path: with --network=host every container binds
	// the same fixed ports, so concurrent PulsarTestSuite instances
	// (e.g. two Go packages running in parallel in the same CI job) must
	// be serialized rather than isolated via distinct remapped ports.
	pulsarTestLockPath = "/tmp/kubescape-pulsar-test.lock"
)

//go:embed scripts/pulsar.sh
var startPulsarScript string

//go:embed scripts/pulsar_stop.sh
var pulsarStopCommand string

type PulsarTestSuite struct {
	suite.Suite
	DefaultTestConfig config.PulsarConfig
	Client            connector.Client
	// AppPortStart/AdminPortStart are retained for API compatibility but
	// no longer used: Pulsar runs with --network=host and always binds
	// pulsarBrokerPort/pulsarAdminPort. See pulsarTestLockPath.
	AppPortStart   int
	AdminPortStart int
	shutdownFunc   func()
	lockFile       *os.File
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

	lockFile, err := os.OpenFile(pulsarTestLockPath, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		suite.FailNow("failed to open pulsar test lock file", err.Error())
	}
	suite.T().Log("waiting for pulsar test lock")
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX); err != nil {
		suite.FailNow("failed to acquire pulsar test lock", err.Error())
	}
	suite.T().Log("acquired pulsar test lock")
	suite.lockFile = lockFile
	// Safe no-op default: if startPulsar below fails via FailNow before
	// the real shutdownFunc is assigned, TearDownSuite still calls
	// shutdownFunc() unconditionally, and it must not panic on nil or the
	// lock release right after it would never run -- leaving every other
	// concurrent PulsarTestSuite blocked indefinitely.
	suite.shutdownFunc = func() {}

	randomContainerName := fmt.Sprintf("pulsar-test-%d", time.Now().UnixNano())
	//start pulsar
	suite.startPulsar(randomContainerName)

	x, _ := json.Marshal(suite.DefaultTestConfig)
	fmt.Println(string(x))
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
	fmt.Println("pulsar admin", kaURL)
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
	suite.Assert().NoError(killPortProcess(pulsarBrokerPort))
	suite.Assert().NoError(killPortProcess(pulsarAdminPort))
	suite.releasePulsarTestLock()
}

func (suite *PulsarTestSuite) releasePulsarTestLock() {
	if suite.lockFile == nil {
		return
	}
	if err := syscall.Flock(int(suite.lockFile.Fd()), syscall.LOCK_UN); err != nil {
		suite.T().Log("failed to release pulsar test lock:", err.Error())
	}
	suite.lockFile.Close()
	suite.lockFile = nil
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

func (suite *PulsarTestSuite) startPulsar(contName string) {
	suite.T().Log("stopping existing pulsar container")
	exec.Command("/bin/sh", "-c", pulsarStopCommand).Run()
	suite.T().Log("starting pulsar")

	// --network=host (see pulsar.sh): bridge-network port publishing
	// (-p host:container) has been observed hanging indefinitely on some
	// CI runner images -- the container comes up and Pulsar itself logs
	// ready, but the published host port never becomes reachable. Host
	// networking bypasses that path entirely, at the cost of Pulsar
	// always binding its fixed default ports rather than a remapped free
	// one; the pulsarTestLockPath flock in SetupSuite is what keeps
	// concurrent suites from colliding on those fixed ports.
	suite.DefaultTestConfig.URL = fmt.Sprintf("pulsar://localhost:%d", pulsarBrokerPort)
	suite.DefaultTestConfig.AdminUrl = fmt.Sprintf("http://localhost:%d", pulsarAdminPort)

	formattedScript := fmt.Sprintf(startPulsarScript, contName)
	out, err := exec.Command("/bin/sh", "-c", formattedScript).CombinedOutput()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			suite.FailNow("failed to start pulsar", err.Error(), string(exitErr.Stderr), string(out))
		}
		suite.FailNow("failed to start pulsar", err.Error(), string(out))
	}
	suite.T().Log("waiting for pulsar to start")
	// 90 attempts x 2s = 180s safety margin; --network=host has been
	// confirmed reachable within ~15s in practice.
	for i := 0; i < 90; i++ {
		isAlive := suite.checkPulsarIsAlive()
		if isAlive {
			return
		}
		time.Sleep(2 * time.Second)
	}
	formmatedScript := fmt.Sprintf(pulsarStopCommand, contName)
	outbytes, err := exec.Command("/bin/sh", "-c", formmatedScript).CombinedOutput()
	if err != nil {
		fmt.Println(string(outbytes), err.Error())
	}
	killPortProcess(pulsarBrokerPort)
	killPortProcess(pulsarAdminPort)
	suite.FailNow("failed to start pulsar")
}
