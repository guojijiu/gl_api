package cloudclient

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/internal/filters"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_public/cloudapi"
)

type projectArticleListPayload struct {
	Page         int                                `json:"page"`
	Size         int                                `json:"size"`
	SearchFilter []filters.ProjectArticleFilterItem `json:"search_filter,omitempty"`
}

func (a *API) FetchProjectArticle(ctx context.Context, platform string, userToken string, opts filters.ProjectArticleFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	params := projectArticleListPayload{Page: 1, Size: 1000, SearchFilter: opts.SearchFilter}
	return a.CallCloudPlatformPOSTJSON(ctx, cloudapi.ProjectArticleListURL(base), userToken, params)
}
