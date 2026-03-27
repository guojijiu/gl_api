package aigateway

import "context"

type ProjectClient struct{ svc *Service }

func NewProjectClient(svc *Service) *ProjectClient { return &ProjectClient{svc: svc} }
func (c *ProjectClient) List(ctx context.Context, platform, token string) ([]byte, int, error) {
	return c.svc.FetchUserProjects(ctx, platform, token)
}
func (c *ProjectClient) ListByContract(ctx context.Context, platform, token string, contractID int) ([]byte, int, error) {
	return c.svc.FetchProjectsByContractID(ctx, platform, token, contractID)
}
func (c *ProjectClient) ZipURL(ctx context.Context, platform, token string, projectID int) ([]byte, int, error) {
	return c.svc.FetchProjectZipURL(ctx, platform, token, projectID)
}
func (c *ProjectClient) OriginalDataURL(ctx context.Context, platform, token string, dataID int, accessCode string) ([]byte, int, error) {
	return c.svc.FetchOriginalDataURL(ctx, platform, token, dataID, accessCode)
}

type ContractClient struct{ svc *Service }

func NewContractClient(svc *Service) *ContractClient { return &ContractClient{svc: svc} }
func (c *ContractClient) List(ctx context.Context, platform, token, contractNumber string) ([]byte, int, error) {
	return c.svc.FetchContracts(ctx, platform, token, contractNumber)
}

type TaskClient struct{ svc *Service }

func NewTaskClient(svc *Service) *TaskClient { return &TaskClient{svc: svc} }
func (c *TaskClient) ListToolByUUID(ctx context.Context, platform, token, uuid string) ([]byte, int, error) {
	return c.svc.FetchTaskListByUUID(ctx, platform, token, uuid)
}
func (c *TaskClient) ListWorkflowByUUID(ctx context.Context, platform, token, uuid string) ([]byte, int, error) {
	return c.svc.FetchWorkflowTaskListByUUID(ctx, platform, token, uuid)
}
func (c *TaskClient) StatusByUUIDs(ctx context.Context, platform, token, uuidsCSV string) ([]byte, int, error) {
	return c.svc.FetchTaskStatusByUUIDs(ctx, platform, token, uuidsCSV)
}
func (c *TaskClient) Result(ctx context.Context, platform, token string, taskID int) ([]byte, int, error) {
	return c.svc.FetchTaskResult(ctx, platform, token, taskID)
}
func (c *TaskClient) DownloadResult(ctx context.Context, platform, token string, taskID int) ([]byte, int, error) {
	return c.svc.DownloadTaskResult(ctx, platform, token, taskID)
}
func (c *TaskClient) ModuleDetail(ctx context.Context, platform, token string, id int) ([]byte, int, error) {
	return c.svc.FetchModuleTaskDetail(ctx, platform, token, id)
}
func (c *TaskClient) ModuleResultURL(ctx context.Context, platform, token string, id int) ([]byte, int, error) {
	return c.svc.FetchModuleTaskResultURL(ctx, platform, token, id)
}

type ProjectArticleClient struct{ svc *Service }

func NewProjectArticleClient(svc *Service) *ProjectArticleClient {
	return &ProjectArticleClient{svc: svc}
}
func (c *ProjectArticleClient) List(ctx context.Context, platform, token string) ([]byte, int, error) {
	return c.svc.FetchProjectArticle(ctx, platform, token)
}
