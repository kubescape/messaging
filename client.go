package pulsarconnector

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kubescape/pulsar-connector/config"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/avast/retry-go"
)

var (
	client       pulsar.Client
	clientConfig *config.PulsarConfig
	initOnce     sync.Once
)

const (
	defaultPulsarCluster = "pulsar"
	adminPath            = "/admin/v2"
	tenantsPath          = adminPath + "/tenants"
	namespacesPath       = adminPath + "/namespaces"
)

// GetClientOnce returns a singleton pulsar client
func GetClientOnce(cfg *config.PulsarConfig) (pulsar.Client, error) {
	if cfg == nil {
		return nil, fmt.Errorf("pulsar config is nil")
	}

	var initErr error

	initOnce.Do(func() {
		client, initErr = pulsar.NewClient(pulsar.ClientOptions{
			URL:              cfg.URL,
			OperationTimeout: time.Second * 10,
		})
		if initErr != nil {
			return
		}
		log.Printf("pulsar client created - testing connection")
		if initErr = retry.Do(pingPulsar,
			retry.LastErrorOnly(true), retry.Attempts(5), retry.MaxDelay(time.Second)); initErr != nil {
			log.Printf("pulsar connection test failed")
			return
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
				return
			}

			namespacePath := cfg.AdminUrl + namespacesPath + "/" + cfg.Tenant + "/" + cfg.Namespace
			log.Printf("creating namespace %s\n", namespacePath)
			if initErr = pulsarAdminRequest(http.MethodPut, namespacePath, nil); initErr != nil {
				return
			}
		}

		clientConfig = cfg
	})
	return client, initErr
}

func pingPulsar() error {
	_, err := client.TopicPartitions("test")
	return err
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

func GetClientConfig() *config.PulsarConfig {
	return clientConfig
}
