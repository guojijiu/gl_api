package cloudclient

import (
	"net/http"
	"time"

	"cloud-platform-api/app/Config"
)

// NewHTTPClient 与 LLM 共用超时，供 NewAPI / llm.NewClient 注入。
func NewHTTPClient(cfg *Config.AiGatewayConfig) *http.Client {
	if cfg == nil {
		cfg = &Config.AiGatewayConfig{}
		cfg.SetDefaults()
	}
	timeout := time.Duration(cfg.LLMTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &http.Client{Timeout: timeout}
}
