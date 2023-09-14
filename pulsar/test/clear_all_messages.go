package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func (suite *PulsarTestSuite) clearAllMessages() error {
	tenants, err := suite.getTenants()
	if err != nil {
		return err
	}

	for _, tenant := range tenants {
		namespaces, err := suite.getNamespaces(tenant)
		if err != nil {
			return err
		}
		for _, namespace := range namespaces {
			topics, err := suite.getTopics(namespace)
			if err != nil {
				return err
			}
			for _, topic := range topics {
				subscriptions, err := suite.getSubscriptions(tenant, namespace, topic)
				if err != nil {
					return err
				}
				for _, subscription := range subscriptions {
					err := suite.skipAllMessages(tenant, namespace, topic, subscription)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return err
}

func (suite *PulsarTestSuite) getTenants() ([]string, error) {
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/tenants")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var tenants []string
	if err := json.Unmarshal(body, &tenants); err != nil {
		return nil, err
	}
	return tenants, nil
}

// getNamespaces returns a list of namespaces for a given tenant in this format: ["public/default", "public/test"]
func (suite *PulsarTestSuite) getNamespaces(tenant string) ([]string, error) {
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/namespaces/" + tenant)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var namespaces []string
	if err := json.Unmarshal(body, &namespaces); err != nil {
		return nil, err
	}
	return namespaces, nil
}

// getTopics returns a list of topics for a given namespace(e.g. "public/default") in this format: ["persistent://public/default/test", "persistent://public/default/test2"]
func (suite *PulsarTestSuite) getTopics(namespace string) ([]string, error) {
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/persistent/" + namespace)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get topics: %s", string(body))
	}
	var topics []string
	if err := json.Unmarshal(body, &topics); err != nil {
		return nil, err
	}
	return topics, nil
}

// getSubscriptions returns a map of subscription names to subscription metadata
func (suite *PulsarTestSuite) getSubscriptions(tenant, namespace, topic string) ([]string, error) {
	topicPath := strings.Replace(topic, "://", "/", 1)
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/" + topicPath + "/subscriptions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get subscriptions: %d; %s", resp.StatusCode, string(body))
	}
	var subscriptions []string
	if err := json.Unmarshal(body, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (suite *PulsarTestSuite) skipAllMessages(tenant, namespace, topic, subscription string) error {
	topicPath := strings.Replace(topic, "://", "/", 1)
	fmt.Println("Skipping messages for", topicPath, subscription)
	req, err := http.NewRequest("POST", suite.DefaultTestConfig.AdminUrl+"/admin/v2/"+topicPath+"/subscription/"+subscription+"/skip_all", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 || resp.StatusCode < 200 {
		body, err := io.ReadAll(resp.Body)
		fmt.Println("Error skipping messages:", string(body))
		return err
	}
	return nil
}
