package aigateway

import (
	"cloud-platform-api/app/Config"
	"net/http"
	"time"
)

type Service struct {
	cfg *Config.AiGatewayConfig
	hc  *http.Client
}

func NewService(cfg *Config.AiGatewayConfig) *Service {
	if cfg == nil {
		cfg = &Config.AiGatewayConfig{}
		cfg.SetDefaults()
	}
	timeout := time.Duration(cfg.LLMTimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 120 * time.Second
	}
	return &Service{
		cfg: cfg,
		hc:  &http.Client{Timeout: timeout},
	}
}
