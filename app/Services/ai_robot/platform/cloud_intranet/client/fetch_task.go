package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/internal/filters"
	"cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/cloudapi"
)

func (a *API) FetchTaskListByUUID(ctx context.Context, platform string, userToken string, opts filters.TaskFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskListURL(base, cloudapi.TaskListQuery{
		Page:           1,
		Size:           1000,
		ProjectNumber:  opts.ProjectNumber,
		ProjectName:    opts.ProjectName,
		UUID:           opts.UUID,
		Name:           opts.Name,
		ToolName:       opts.ToolName,
		StatusValue:    opts.StatusValue,
		WorkflowNameCN: opts.WorkflowNameCN,
		CreatedAtStart: opts.CreatedAtStart,
		CreatedAtEnd:   opts.CreatedAtEnd,
	}), userToken)
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

func (a *API) FetchWorkflowTaskListByUUID(ctx context.Context, platform string, userToken string, opts filters.TaskFilters) ([]byte, int, error) {
	if err := a.RequireCloudAPIBase(platform); err != nil {
		return nil, 0, err
	}
	base := a.cfg.CloudAPIBaseByPlatform(platform)
	return a.CallCloudPlatformGET(ctx, cloudapi.TaskListOfWorkflowURL(base, cloudapi.TaskListQuery{
		Page:           1,
		Size:           1000,
		ProjectNumber:  opts.ProjectNumber,
		ProjectName:    opts.ProjectName,
		UUID:           opts.UUID,
		Name:           opts.Name,
		ToolName:       opts.ToolName,
		StatusValue:    opts.StatusValue,
		WorkflowNameCN: opts.WorkflowNameCN,
		CreatedAtStart: opts.CreatedAtStart,
		CreatedAtEnd:   opts.CreatedAtEnd,
	}), userToken)
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
