//go:build integration

package test_test

import (
	"net/url"
	"strconv"
	"testing"

	messagingtest "github.com/kubescape/messaging/pulsar/test"
	"github.com/stretchr/testify/suite"
)

type publicPulsarSuite struct {
	messagingtest.PulsarTestSuite
}

func TestPublicPulsarTestSuite(t *testing.T) {
	suite.Run(t, new(publicPulsarSuite))
}

func (suite *publicPulsarSuite) TestClientAndMappedEndpointsAreReady() {
	suite.Require().NotNil(suite.Client)
	suite.Require().Positive(suite.AppPortStart)
	suite.Require().Positive(suite.AdminPortStart)
	suite.Equal(strconv.Itoa(suite.AppPortStart), endpointPort(suite.DefaultTestConfig.URL))
	suite.Equal(strconv.Itoa(suite.AdminPortStart), endpointPort(suite.DefaultTestConfig.AdminUrl))

	partitions, err := suite.Client.TopicPartitions("test-topic")
	suite.Require().NoError(err)
	suite.NotEmpty(partitions)
}

func endpointPort(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	return parsed.Port()
}
