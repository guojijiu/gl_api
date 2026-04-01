package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/internal/filters"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/cloudapi"
)

func (a *API) FetchUserProjects(ctx context.Context, platform string, userToken string, opts filters.ProjectFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectGetUserAllProjectURL(base, cloudapi.ProjectGetUserAllProjectQuery{
		IsFilterTime: opts.IsFilterTime,
		IsUsedFree:   opts.IsUsedFree,
	}), userToken)
}

func (a *API) FetchProjectZipURL(ctx context.Context, platform string, userToken string, projectID int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectGetZipURL(base, projectID), userToken)
}

func (a *API) FetchOriginalDataURL(ctx context.Context, platform string, userToken string, dataID int, accessCode string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectGetOriginalDataURL(base, dataID, accessCode), userToken)
}

func (a *API) FetchContracts(ctx context.Context, platform string, userToken string, opts filters.ContractFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ContractListOfProjectURL(base, cloudapi.ContractListOfProjectQuery{
		Page:           1,
		Size:           200,
		ContractNumber: opts.ContractNumber,
		Name:           opts.Name,
	}), userToken)
}

func (a *API) FetchProjectsByContractID(ctx context.Context, platform string, userToken string, contractID int, opts filters.ProjectFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectListByContractURL(base, cloudapi.ProjectListByContractQuery{
		ContractID:     contractID,
		Page:           1,
		Size:           1000,
		Number:         opts.Number,
		Name:           opts.Name,
		WorkflowNameCN: opts.WorkflowNameCN,
	}), userToken)
}
