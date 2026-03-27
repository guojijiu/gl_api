package aigateway

import (
	"context"
	"errors"
	"strings"
)

func (s *Service) FetchProjectArticle(ctx context.Context, platform string, userToken string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	params := map[string]interface{}{
		"page": 1,
		"size": 1000,
	}
	return s.CallCloudPlatformPOSTJSON(ctx, s.cfg.ProjectArticleListURL(platform), userToken, params)
}
