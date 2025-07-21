package connector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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
	topicsPath           = adminPath + "/persistent"

	dlqNamespaceSuffix   = "-dlqs"
	retryNamespaceSuffix = "-retry"
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
	SetTopicMaxUnackedMessagesPerConsumer(topicName string, maxUnackedMessages int) error
	GetTopicMaxUnackedMessagesPerConsumer(topicName string) (int, error)
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

// NewConsumer creates a new consumer. When subscription type is not provided, it defaults to Shared
func (p *pulsarClient) NewConsumer(createConsumerOpts ...CreateConsumerOption) (Consumer, error) {
	return newConsumer(p, createConsumerOpts...)
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
			return nil, fmt.Errorf("failed to create dlq namespace: %w", initErr)
		}
		retryNamespacePath := namespacePath + retryNamespaceSuffix
		log.Printf("creating retry namespace %s\n", retryNamespacePath)
		if initErr = pulsarAdminRequest(http.MethodPut, retryNamespacePath, nil); initErr != nil {
			return nil, fmt.Errorf("failed to create retry namespace: %w", initErr)
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

// pulsarAdminRawRequest sends an HTTP request to the Pulsar admin API with a raw body, returns response body and status code
func pulsarAdminRawRequest(method, url string, body io.Reader, contentType string) ([]byte, int, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, 0, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return respBody, resp.StatusCode, nil
}

func (p *pulsarClient) SetTopicMaxUnackedMessagesPerConsumer(topicName string, maxUnackedMessages int) error {
	if p.config.AdminUrl == "" {
		return fmt.Errorf("admin URL is not configured")
	}
	setURL := p.topicAdminURL(topicName)
	body := strings.NewReader(fmt.Sprintf("%d", maxUnackedMessages))
	respBody, status, err := pulsarAdminRawRequest(http.MethodPost, setURL, body, "application/json")
	if err != nil {
		return err
	}
	if status != 204 && status != 200 {
		return fmt.Errorf("failed to set maxUnackedMessagesOnConsumer: status %d, body: %s", status, string(respBody))
	}
	return nil
}

func (p *pulsarClient) GetTopicMaxUnackedMessagesPerConsumer(topicName string) (int, error) {
	if p.config.AdminUrl == "" {
		return -1, fmt.Errorf("admin URL is not configured")
	}
	getURL := p.topicAdminURL(topicName)
	respBody, status, err := pulsarAdminRawRequest(http.MethodGet, getURL, nil, "")
	if err != nil {
		return -1, err
	}
	val := strings.TrimSpace(string(respBody))
	if status != 200 {
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", status, val)
	}
	if val == "" {
		return -1, nil
	}
	num, err := strconv.Atoi(val)
	if err != nil {
		return -1, fmt.Errorf("failed to parse value: %s", val)
	}
	return num, nil
}

// topicAdminURL builds the admin URL for maxUnackedMessagesOnConsumer for a topic
func (p *pulsarClient) topicAdminURL(topicName string) string {
	fullTopicName := topicName
	if !isFullTopicName(topicName) {
		fullTopicName = fmt.Sprintf("%s/%s/%s", p.config.Tenant, p.config.Namespace, topicName)
	}
	return fmt.Sprintf("%s/admin/v2/persistent/%s/maxUnackedMessagesOnConsumer", p.config.AdminUrl, fullTopicName)
}

// isFullTopicName checks if the topic name includes tenant/namespace
func isFullTopicName(topicName string) bool {
	// A full topic name should contain at least two slashes (tenant/namespace/topic)
	count := 0
	for _, char := range topicName {
		if char == '/' {
			count++
		}
	}
	return count >= 2
}
