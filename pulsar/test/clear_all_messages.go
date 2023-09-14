package test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
			topics, err := suite.getTopics(tenant, namespace)
			if err != nil {
				return err
			}
			for _, topic := range topics {
				subscriptions, err := suite.getSubscriptions(tenant, namespace, topic)
				if err != nil {
					return err
				}
				for subscription := range subscriptions {
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

func (suite *PulsarTestSuite) getTopics(tenant, namespace string) ([]string, error) {
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/persistent/" + tenant + "/" + namespace)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var topics []string
	if err := json.Unmarshal(body, &topics); err != nil {
		return nil, err
	}
	return topics, nil
}

// getSubscriptions returns a map of subscription names to subscription metadata
func (suite *PulsarTestSuite) getSubscriptions(tenant, namespace, topic string) (map[string]interface{}, error) {
	resp, err := http.Get(suite.DefaultTestConfig.AdminUrl + "/admin/v2/persistent/" + tenant + "/" + namespace + "/" + topic + "/subscriptions")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var subscriptions map[string]interface{}
	if err := json.Unmarshal(body, &subscriptions); err != nil {
		return nil, err
	}

	return subscriptions, nil
}

func (suite *PulsarTestSuite) skipAllMessages(tenant, namespace, topic, subscription string) error {
	req, err := http.NewRequest("POST", suite.DefaultTestConfig.AdminUrl+"/admin/v2/persistent/"+tenant+"/"+namespace+"/"+topic+"/subscription/"+subscription+"/skip_all", nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, err := io.ReadAll(resp.Body)
		fmt.Println("Error skipping messages:", string(body))
		return err
	}
	return nil
}
