package flow

import (
	"cloud-platform-api/app/Http/Requests"
	"cloud-platform-api/app/Services/ai_robot/internal/deps"
	cloudintranet "cloud-platform-api/app/Services/ai_robot/platform/cloud_intranet"
	cloudpublic "cloud-platform-api/app/Services/ai_robot/platform/cloud_public"
	imagecompare "cloud-platform-api/app/Services/ai_robot/platform/image_compare"
)

// dispatchByRegistry 为统一分发总入口：
// 1) 能力校验（平台 + question_type），平台标识以 Backend 为准（与 NewForPlatform 注入一致）；
// 2) 平台分流：仅转发到 platform/<平台> 包；
// 3) 各平台包内再按 question_type / 领域处理。
func (p *Processor) dispatchByRegistry(d *deps.Deps) {
	pid := d.PlatformID()
	if !Requests.IsSupportedPlatform(pid) {
		p.frontFailed(d.Gin, "不支持的 platform", nil)
		return
	}
	if !Requests.IsSupportedQuestionType(pid, d.Req.QuestionType) {
		p.frontFailed(d.Gin, "当前平台下不支持的提问类型", nil)
		return
	}

	if !p.dispatchPlatform(pid, d) {
		p.frontFailed(d.Gin, "不支持的 platform", nil)
		return
	}
}

// platformRegistry 与 backends.NewForPlatform、Requests 能力矩阵须一致；新增平台步骤见 internal/backends/registry.go 文件头 checklist。
func (p *Processor) dispatchPlatform(platformID string, d *deps.Deps) bool {
	switch platformID {
	case Requests.AiPlatformCloudPublic:
		p.handleCloudPublicPlatform(d)
	case Requests.AiPlatformCloudIntranet:
		p.handleCloudIntranetPlatform(d)
	case Requests.AiPlatformImageCompare:
		p.handleImageComparePlatform(d)
	default:
		return false
	}
	return true
}

func (p *Processor) handleCloudPublicPlatform(d *deps.Deps) {
	cloudpublic.DispatchPublic(p, d)
}

func (p *Processor) handleCloudIntranetPlatform(d *deps.Deps) {
	cloudintranet.DispatchIntranet(p, d)
}

func (p *Processor) handleImageComparePlatform(d *deps.Deps) {
	imagecompare.Dispatch(p, d)
}
