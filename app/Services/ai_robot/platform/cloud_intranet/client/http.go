package client

import (
	"net/http"
	"time"

	"cloud-platform-api/app/Config"
)

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
