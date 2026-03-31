package backends

// Backend 表示「当前请求」所属平台的基础设施句柄。
// 扩展新平台时：在此包增加实现，并在 registry 中注册，勿在 deps 上堆叠全局 client。
type Backend interface {
	// PlatformID 与请求中的 platform 一致，作为下游 URL/路由选择的唯一来源（与 d.Req.Platform 应对齐）。
	PlatformID() string
}
