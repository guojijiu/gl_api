package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/cloudapi"
)

func (a *API) FetchUserProjects(ctx context.Context, platform string, userToken string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectGetUserAllProjectURL(base), userToken)
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

func (a *API) FetchContracts(ctx context.Context, platform string, userToken string, contractNumber string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ContractListOfProjectURL(base, 1, 200, contractNumber), userToken)
}

func (a *API) FetchProjectsByContractID(ctx context.Context, platform string, userToken string, contractID int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.ProjectListByContractURL(base, contractID, 1, 1000), userToken)
}
