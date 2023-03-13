package config

type PulsarConfig struct {
	URL                    string   `json:"url"`
	Tenant                 string   `json:"tenant"`
	Namespace              string   `json:"namespace"`
	AdminUrl               string   `json:"adminUrl"`
	Clusters               []string `json:"clusters"`
	RedeliveryDelaySeconds int      `json:"redeliveryDelaySeconds,omitempty"`
	MaxDeliveryAttempts    int      `json:"maxDeliveryAttempts,omitempty"`
}

type TelemetryConfig struct {
	JaegerAgentHost string `json:"jaegerAgentHost"`
	JaegerAgentPort string `json:"jaegerAgentPort"`
}

type LoggerConfig struct {
	Level       string `json:"level"`
	LogFileName string `json:"logFileName"`
}
