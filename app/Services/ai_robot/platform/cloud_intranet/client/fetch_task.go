package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/cloudapi"
)

func (a *API) FetchTaskListByUUID(ctx context.Context, platform string, userToken string, uuid string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskListURL(base, 1, 1000, uuid), userToken)
}

func (a *API) FetchTaskStatusByUUIDs(ctx context.Context, platform string, userToken string, uuidsCSV string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskStatusByUUIDsURL(base, uuidsCSV), userToken)
}

func (a *API) FetchTaskResult(ctx context.Context, platform string, userToken string, taskID int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskGetResultURL(base, taskID), userToken)
}

func (a *API) DownloadTaskResult(ctx context.Context, platform string, userToken string, taskID int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformPOSTJSON(ctx, cloudapi.TaskDownloadResultURL(base), userToken, map[string]interface{}{"id": taskID})
}

func (a *API) FetchWorkflowTaskListByUUID(ctx context.Context, platform string, userToken string, uuid string) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskListOfWorkflowURL(base, 1, 1000, uuid), userToken)
}

func (a *API) FetchModuleTaskDetail(ctx context.Context, platform string, userToken string, id int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskDetailOfModuleToolURL(base, id), userToken)
}

func (a *API) FetchModuleTaskResultURL(ctx context.Context, platform string, userToken string, id int) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskPdfURLOfModuleToolURL(base, id), userToken)
}
