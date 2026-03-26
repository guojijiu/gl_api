package flow

import "cloud-platform-api/app/Http/Requests"

type aiBusinessHandler func(*aiRobotDeps)

// dispatchByRegistry 为统一分发总入口：
// 1) 能力校验（平台 + question_type）；
// 2) 平台分流（公网/内网/图片）；
// 3) 平台内业务分流（project/task）。
func (p *Processor) dispatchByRegistry(d *aiRobotDeps) {
	if !Requests.IsSupportedPlatform(d.req.Platform) {
		p.frontFailed(d.gin, "不支持的 platform", nil)
		return
	}
	if !Requests.IsSupportedQuestionType(d.req.Platform, d.req.QuestionType) {
		p.frontFailed(d.gin, "当前平台下不支持的提问类型", nil)
		return
	}

	platformHandler := p.platformRegistry()[d.req.Platform]
	if platformHandler == nil {
		p.frontFailed(d.gin, "不支持的 platform", nil)
		return
	}
	platformHandler(d)
}

func (p *Processor) platformRegistry() map[string]aiDomainHandler {
	return map[string]aiDomainHandler{
		Requests.AiPlatformCloudPublic:   p.handleCloudPublicPlatform,
		Requests.AiPlatformCloudIntranet: p.handleCloudIntranetPlatform,
		Requests.AiPlatformImageCompare:  p.handleImageComparePlatform,
	}
}

// 当前公网/内网平台先共用同一套业务处理器。
// 如果后续两套逻辑完全分离，可在这里替换为不同处理函数。
func (p *Processor) handleCloudPublicPlatform(d *aiRobotDeps) {
	businessHandler := p.cloudPublicBusinessRegistry()[d.req.QuestionType]
	if businessHandler == nil {
		p.frontFailed(d.gin, "公网云平台下不支持的提问类型", nil)
		return
	}
	businessHandler(d)
}

func (p *Processor) handleCloudIntranetPlatform(d *aiRobotDeps) {
	businessHandler := p.cloudIntranetBusinessRegistry()[d.req.QuestionType]
	if businessHandler == nil {
		p.frontFailed(d.gin, "内网云平台下不支持的提问类型", nil)
		return
	}
	businessHandler(d)
}

func (p *Processor) handleImageComparePlatform(d *aiRobotDeps) {
	businessHandler := p.imageCompareBusinessRegistry()[d.req.QuestionType]
	if businessHandler == nil {
		p.frontFailed(d.gin, "图片对比平台下不支持的提问类型", nil)
		return
	}
	businessHandler(d)
}

func (p *Processor) cloudPublicBusinessRegistry() map[int]aiBusinessHandler {
	return map[int]aiBusinessHandler{
		Requests.AiQuestionTypeProject: p.handleProjectDomain,
		Requests.AiQuestionTypeTask:    p.handleTaskDomain,
	}
}

func (p *Processor) cloudIntranetBusinessRegistry() map[int]aiBusinessHandler {
	return map[int]aiBusinessHandler{
		Requests.AiQuestionTypeProject: p.handleProjectDomain,
		Requests.AiQuestionTypeTask:    p.handleTaskDomain,
	}
}

func (p *Processor) imageCompareBusinessRegistry() map[int]aiBusinessHandler {
	return map[int]aiBusinessHandler{
		Requests.AiQuestionTypeProject: p.handleImageDomain,
	}
}
