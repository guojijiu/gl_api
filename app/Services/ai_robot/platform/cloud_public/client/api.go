package cloudclient

import (
	"errors"
	"net/http"
	"strings"

	"cloud-platform-api/app/Config"
)

// API 封装云平台 HTTP 与按 platform 的 URL 访问。
type API struct {
	cfg *Config.AiGatewayConfig
	hc  *http.Client
}

func NewAPI(cfg *Config.AiGatewayConfig, hc *http.Client) *API {
	return &API{cfg: cfg, hc: hc}
}

// RequireCloudAPIBase 校验当前 platform 已配置云平台 base URL。
func (a *API) RequireCloudAPIBase(platform string) error {
	base := strings.TrimSpace(a.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return nil
}
