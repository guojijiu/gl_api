package flow

import (
	"context"

	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Http/Requests"
	AiGateway "cloud-platform-api/app/Services/ai_gateway"

	"github.com/gin-gonic/gin"
)

type aiRobotDeps struct {
	ctx   context.Context
	gin   *gin.Context
	req   *Requests.AiRobotChatRequest
	token string
	cfg   *Config.AiGatewayConfig
	svc   *AiGateway.Service

	projectClient  *AiGateway.ProjectClient
	contractClient *AiGateway.ContractClient
	taskClient     *AiGateway.TaskClient
}

type aiDomainHandler func(*aiRobotDeps)

// 以下 3 个方法是懒初始化：
// 只有真正命中对应业务分支时，才会创建 client，减少无效对象创建。
func (d *aiRobotDeps) project() *AiGateway.ProjectClient {
	if d.projectClient == nil {
		d.projectClient = AiGateway.NewProjectClient(d.svc)
	}
	return d.projectClient
}

func (d *aiRobotDeps) contract() *AiGateway.ContractClient {
	if d.contractClient == nil {
		d.contractClient = AiGateway.NewContractClient(d.svc)
	}
	return d.contractClient
}

func (d *aiRobotDeps) task() *AiGateway.TaskClient {
	if d.taskClient == nil {
		d.taskClient = AiGateway.NewTaskClient(d.svc)
	}
	return d.taskClient
}
