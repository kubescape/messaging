package test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kubescape/messaging/pulsar/config"
	"github.com/kubescape/messaging/pulsar/connector"
	"github.com/kubescape/messaging/pulsar/internal/pulsartest"

	"github.com/stretchr/testify/suite"
)

type PulsarTestSuite struct {
	suite.Suite
	DefaultTestConfig config.PulsarConfig
	Client            connector.Client
	// AppPortStart is retained for source compatibility. SetupSuite replaces it
	// with the dynamically mapped Pulsar broker port.
	AppPortStart int
	// AdminPortStart is retained for source compatibility. SetupSuite replaces it
	// with the dynamically mapped Pulsar admin port.
	AdminPortStart int
	broker         *pulsartest.Broker
	cleanupOnce    sync.Once
	cleanupErr     error
}

func (suite *PulsarTestSuite) SetupSuite() {
	suite.T().Log("setup suite")
	broker, err := pulsartest.Start(context.Background())
	if err != nil {
		suite.FailNow("failed to start Pulsar", err.Error())
	}
	suite.broker = broker
	suite.AppPortStart = broker.Port
	suite.AdminPortStart = broker.AdminPort
	suite.T().Cleanup(func() {
		suite.Require().NoError(suite.cleanup())
	})

	suite.DefaultTestConfig = config.PulsarConfig{
		URL:                    broker.URL,
		AdminUrl:               broker.AdminURL,
		Tenant:                 "ca-messaging",
		Namespace:              "test-namespace",
		Clusters:               []string{"standalone"},
		MaxDeliveryAttempts:    2,
		RedeliveryDelaySeconds: 0,
	}

	x, _ := json.Marshal(suite.DefaultTestConfig)
	fmt.Println(string(x))
	// Ensure the wrapper can connect and initialize its namespaces.
	suite.Client, err = connector.NewClient(connector.WithConfig(&suite.DefaultTestConfig))
	if err != nil {
		suite.FailNow("failed to create pulsar client", err.Error())
	}
}

func (suite *PulsarTestSuite) TearDownSuite() {
	suite.T().Log("tear down suite")
	suite.Require().NoError(suite.cleanup())
}

func (suite *PulsarTestSuite) cleanup() error {
	suite.cleanupOnce.Do(func() {
		if suite.Client != nil {
			suite.Client.Close()
		}
		if suite.broker != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			suite.cleanupErr = suite.broker.Terminate(ctx)
		}
	})
	return suite.cleanupErr
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
