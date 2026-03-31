package backends

// Minimal 无下游云平台 HTTP 的平台（仅占位、仅用 LLM 编排等）。
type Minimal struct {
	platformID string
}

func NewMinimal(platformID string) *Minimal {
	return &Minimal{platformID: platformID}
}

func (m *Minimal) PlatformID() string { return m.platformID }
