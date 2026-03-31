package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/cloudapi"
)

func (a *API) FetchProjectArticle(ctx context.Context, platform string, userToken string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	params := map[string]interface{}{
		"page": 1,
		"size": 1000,
	}
	return a.CallCloudPlatformPOSTJSON(ctx, cloudapi.ProjectArticleListURL(base), userToken, params)
}
