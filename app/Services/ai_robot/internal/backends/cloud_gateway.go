package backends

import (
	intranetclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet/client"
	cloudclient "cloud-platform-api/app/Services/ai_robot/platform/cloud_public/client"
	"context"
)

type ProjectClient interface {
	List(context.Context, string, string) ([]byte, int, error)
	ListByContract(context.Context, string, string, int) ([]byte, int, error)
	ZipURL(context.Context, string, string, int) ([]byte, int, error)
	OriginalDataURL(context.Context, string, string, int, string) ([]byte, int, error)
}

type ContractClient interface {
	List(context.Context, string, string, string) ([]byte, int, error)
}

type TaskClient interface {
	ListToolByUUID(context.Context, string, string, string) ([]byte, int, error)
	ListWorkflowByUUID(context.Context, string, string, string) ([]byte, int, error)
	StatusByUUIDs(context.Context, string, string, string) ([]byte, int, error)
	Result(context.Context, string, string, int) ([]byte, int, error)
	DownloadResult(context.Context, string, string, int) ([]byte, int, error)
	ModuleDetail(context.Context, string, string, int) ([]byte, int, error)
	ModuleResultURL(context.Context, string, string, int) ([]byte, int, error)
}

type ProjectArticleClient interface {
	List(context.Context, string, string) ([]byte, int, error)
}

// CloudGateway 绑定 cloud_public/client.API 与 platform；URL 由 cfg.CloudAPIBaseByPlatform(platform) 决定。
type CloudGateway struct {
	platformID string

	project        ProjectClient
	contract       ContractClient
	task           TaskClient
	projectArticle ProjectArticleClient
}

func newCloudGateway(platformID string, project ProjectClient, contract ContractClient, task TaskClient, projectArticle ProjectArticleClient) *CloudGateway {
	return &CloudGateway{
		platformID:     platformID,
		project:        project,
		contract:       contract,
		task:           task,
		projectArticle: projectArticle,
	}
}

func NewPublicCloudGateway(platformID string, api *cloudclient.API) *CloudGateway {
	return newCloudGateway(
		platformID,
		cloudclient.NewProjectClient(api),
		cloudclient.NewContractClient(api),
		cloudclient.NewTaskClient(api),
		cloudclient.NewProjectArticleClient(api),
	)
}

func NewIntranetCloudGateway(platformID string, api *intranetclient.API) *CloudGateway {
	return newCloudGateway(
		platformID,
		intranetclient.NewProjectClient(api),
		intranetclient.NewContractClient(api),
		intranetclient.NewTaskClient(api),
		intranetclient.NewProjectArticleClient(api),
	)
}

func (g *CloudGateway) Project() ProjectClient { return g.project }

func (g *CloudGateway) Contract() ContractClient { return g.contract }

func (g *CloudGateway) Task() TaskClient { return g.task }

func (g *CloudGateway) ProjectArticle() ProjectArticleClient { return g.projectArticle }

func (g *CloudGateway) PlatformID() string { return g.platformID }
