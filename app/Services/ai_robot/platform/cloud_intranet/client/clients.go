package client

import (
	"context"

	"cloud-platform-api/app/Services/ai_robot/internal/filters"
)

type ProjectClient struct{ api *API }

func NewProjectClient(api *API) *ProjectClient { return &ProjectClient{api: api} }
func (c *ProjectClient) List(ctx context.Context, platform, token string, opts filters.ProjectFilters) ([]byte, int, error) {
	return c.api.FetchUserProjects(ctx, platform, token, opts)
}
func (c *ProjectClient) ListByContract(ctx context.Context, platform, token string, contractID int, opts filters.ProjectFilters) ([]byte, int, error) {
	return c.api.FetchProjectsByContractID(ctx, platform, token, contractID, opts)
}
func (c *ProjectClient) ZipURL(ctx context.Context, platform, token string, projectID int) ([]byte, int, error) {
	return c.api.FetchProjectZipURL(ctx, platform, token, projectID)
}
func (c *ProjectClient) OriginalDataURL(ctx context.Context, platform, token string, dataID int, accessCode string) ([]byte, int, error) {
	return c.api.FetchOriginalDataURL(ctx, platform, token, dataID, accessCode)
}

type ContractClient struct{ api *API }

func NewContractClient(api *API) *ContractClient { return &ContractClient{api: api} }
func (c *ContractClient) List(ctx context.Context, platform, token string, opts filters.ContractFilters) ([]byte, int, error) {
	return c.api.FetchContracts(ctx, platform, token, opts)
}

type TaskClient struct{ api *API }

func NewTaskClient(api *API) *TaskClient { return &TaskClient{api: api} }
func (c *TaskClient) ListToolByUUID(ctx context.Context, platform, token string, opts filters.TaskFilters) ([]byte, int, error) {
	return c.api.FetchTaskListByUUID(ctx, platform, token, opts)
}
func (c *TaskClient) ListWorkflowByUUID(ctx context.Context, platform, token string, opts filters.TaskFilters) ([]byte, int, error) {
	return c.api.FetchWorkflowTaskListByUUID(ctx, platform, token, opts)
}
func (c *TaskClient) StatusByUUIDs(ctx context.Context, platform, token, uuidsCSV string) ([]byte, int, error) {
	return c.api.FetchTaskStatusByUUIDs(ctx, platform, token, uuidsCSV)
}
func (c *TaskClient) Result(ctx context.Context, platform, token string, taskID int) ([]byte, int, error) {
	return c.api.FetchTaskResult(ctx, platform, token, taskID)
}
func (c *TaskClient) DownloadResult(ctx context.Context, platform, token string, taskID int) ([]byte, int, error) {
	return c.api.DownloadTaskResult(ctx, platform, token, taskID)
}
func (c *TaskClient) ModuleDetail(ctx context.Context, platform, token string, id int) ([]byte, int, error) {
	return c.api.FetchModuleTaskDetail(ctx, platform, token, id)
}
func (c *TaskClient) ModuleResultURL(ctx context.Context, platform, token string, id int) ([]byte, int, error) {
	return c.api.FetchModuleTaskResultURL(ctx, platform, token, id)
}

type ProjectArticleClient struct{ api *API }

func NewProjectArticleClient(api *API) *ProjectArticleClient {
	return &ProjectArticleClient{api: api}
}
func (c *ProjectArticleClient) List(ctx context.Context, platform, token string, opts filters.ProjectArticleFilters) ([]byte, int, error) {
	return c.api.FetchProjectArticle(ctx, platform, token, opts)
}

// 编译期接口一致性兜底（后续替换内网实现时避免签名漂移）。
type _projectClientContract interface {
	List(context.Context, string, string, filters.ProjectFilters) ([]byte, int, error)
	ListByContract(context.Context, string, string, int, filters.ProjectFilters) ([]byte, int, error)
	ZipURL(context.Context, string, string, int) ([]byte, int, error)
	OriginalDataURL(context.Context, string, string, int, string) ([]byte, int, error)
}

type _contractClientContract interface {
	List(context.Context, string, string, filters.ContractFilters) ([]byte, int, error)
}

type _taskClientContract interface {
	ListToolByUUID(context.Context, string, string, filters.TaskFilters) ([]byte, int, error)
	ListWorkflowByUUID(context.Context, string, string, filters.TaskFilters) ([]byte, int, error)
	StatusByUUIDs(context.Context, string, string, string) ([]byte, int, error)
	Result(context.Context, string, string, int) ([]byte, int, error)
	DownloadResult(context.Context, string, string, int) ([]byte, int, error)
	ModuleDetail(context.Context, string, string, int) ([]byte, int, error)
	ModuleResultURL(context.Context, string, string, int) ([]byte, int, error)
}

type _projectArticleClientContract interface {
	List(context.Context, string, string, filters.ProjectArticleFilters) ([]byte, int, error)
}

var (
	_ _projectClientContract        = (*ProjectClient)(nil)
	_ _contractClientContract       = (*ContractClient)(nil)
	_ _taskClientContract           = (*TaskClient)(nil)
	_ _projectArticleClientContract = (*ProjectArticleClient)(nil)
)
