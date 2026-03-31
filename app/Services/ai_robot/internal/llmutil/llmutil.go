package llmutil

import "strings"

// ExtractJSONFromLLM 从模型返回中剥离 markdown 代码块等，得到纯 JSON 文本。
// 该函数属于 ai_robot 的“纯工具”，避免 ai_robot 依赖 ai_gateway/llm 包的具体实现。
func ExtractJSONFromLLM(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```") {
		lines := strings.SplitN(s, "\n", 2)
		if len(lines) > 1 {
			s = lines[1]
		}
		if idx := strings.LastIndex(s, "```"); idx > 0 {
			s = s[:idx]
		}
	}
	return strings.TrimSpace(s)
}

// Truncate 截断长文本用于日志/错误信息。
func Truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
