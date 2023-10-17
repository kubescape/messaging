package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/kubescape/messaging/pulsar/config"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/avast/retry-go"
)

const (
	defaultPulsarCluster = "pulsar"
	adminPath            = "/admin/v2"
	tenantsPath          = adminPath + "/tenants"
	namespacesPath       = adminPath + "/namespaces"

	dlqNamespaceSuffix = "-dlqs"
)

type PulsarClientOptions struct {
	retryAttempts    *int
	retryMaxDelay    *time.Duration
	config           *config.PulsarConfig
	operationTimeout *time.Duration
}

type Client interface {
	pulsar.Client
	GetConfig() config.PulsarConfig
	NewProducer(createProducerOption ...CreateProducerOption) (Producer, error)
	NewConsumer(createConsumerOpts ...CreateConsumerOption) (Consumer, error)
}

type pulsarClient struct {
	pulsar.Client
	config config.PulsarConfig
}

func (p *pulsarClient) GetConfig() config.PulsarConfig {
	return p.config
}

func (p *pulsarClient) NewProducer(createProducerOption ...CreateProducerOption) (Producer, error) {
	return newProducer(p, createProducerOption...)
}

func (p *pulsarClient) NewConsumer(createConsumerOpts ...CreateConsumerOption) (Consumer, error) {
	return newSharedConsumer(p, createConsumerOpts...)
}

func (p *pulsarClient) ping() error {
	_, err := p.TopicPartitions("test")
	return err
}

// NewClient creates a new pulsar client
func NewClient(options ...func(*PulsarClientOptions)) (Client, error) {
	clientOptions := &PulsarClientOptions{}
	for _, o := range options {
		o(clientOptions)
	}

	cfg := clientOptions.config
	if cfg == nil {
		return nil, fmt.Errorf("pulsar config is nil. use WithConfig to set it")
	}
	retryAttempts := 5
	if clientOptions.retryAttempts != nil {
		retryAttempts = *clientOptions.retryAttempts
	}

	retryMaxDelay := time.Second
	if clientOptions.retryMaxDelay != nil {
		retryMaxDelay = *clientOptions.retryMaxDelay
	}

	operationTimeout := time.Second * 10
	if clientOptions.operationTimeout != nil {
		operationTimeout = *clientOptions.operationTimeout
	}

	var initErr error

	client, initErr := pulsar.NewClient(pulsar.ClientOptions{
		URL:              cfg.URL,
		OperationTimeout: operationTimeout,
	})

	if initErr != nil {
		return nil, fmt.Errorf("failed to create pulsar client: %w", initErr)
	}
	pulsarClient := &pulsarClient{
		Client: client,
		config: *cfg,
	}

	log.Printf("pulsar client created - testing connection")
	if initErr = retry.Do(pulsarClient.ping,
		retry.LastErrorOnly(true), retry.Attempts(uint(retryAttempts)), retry.MaxDelay(retryMaxDelay)); initErr != nil {
		log.Printf("pulsar connection test failed")
		return nil, fmt.Errorf("failed to ping pulsar: %w", initErr)
	}
	if cfg.AdminUrl != "" {
		log.Printf("pulsar admin url is set")

		body := map[string]interface{}{}
		if len(cfg.Clusters) != 0 {
			body["allowedClusters"] = cfg.Clusters
		} else {
			body["allowedClusters"] = []string{defaultPulsarCluster}
		}
		tenantsPath := cfg.AdminUrl + tenantsPath + "/" + cfg.Tenant
		log.Printf("creating tenant %s\n", tenantsPath)
		if initErr = pulsarAdminRequest(http.MethodPut, tenantsPath, body); initErr != nil {
			return nil, fmt.Errorf("failed to create tenant: %w", initErr)
		}

		namespacePath := cfg.AdminUrl + namespacesPath + "/" + cfg.Tenant + "/" + cfg.Namespace
		log.Printf("creating namespace %s\n", namespacePath)
		if initErr = pulsarAdminRequest(http.MethodPut, namespacePath, nil); initErr != nil {
			return nil, fmt.Errorf("failed to create namespace: %w", initErr)
		}
		dlqNamespacePath := namespacePath + dlqNamespaceSuffix
		log.Printf("creating dlq namespace %s\n", dlqNamespacePath)
		if initErr = pulsarAdminRequest(http.MethodPut, dlqNamespacePath, nil); initErr != nil {
			return nil, fmt.Errorf("failed to create namespace: %w", initErr)
		}
	}
	return pulsarClient, nil
}

func Ptr[T any](v T) *T {
	return &v
}

func WithOperationTimeout(operationTimeout time.Duration) func(*PulsarClientOptions) {
	return func(o *PulsarClientOptions) {
		o.operationTimeout = Ptr(operationTimeout)
	}
}

func WithConfig(config *config.PulsarConfig) func(*PulsarClientOptions) {
	return func(o *PulsarClientOptions) {
		o.config = config
	}
}

func WithRetryAttempts(retryAttempts int) func(*PulsarClientOptions) {
	return func(o *PulsarClientOptions) {
		o.retryAttempts = Ptr(retryAttempts)
	}
}

func WithRetryMaxDelay(maxDelay time.Duration) func(*PulsarClientOptions) {
	return func(o *PulsarClientOptions) {
		o.retryMaxDelay = Ptr(maxDelay)
	}
}

func pulsarAdminRequest(method, url string, body interface{}) error {
	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Print("pulsar admin request: ", url, " succeeded")
		return nil
	} else if resp.StatusCode == http.StatusConflict {
		log.Print("pulsar admin request: ", url, " already exists")
		return nil
	}
	return fmt.Errorf("pulsar admin request: %s failed with status code: %d", url, resp.StatusCode)
}
