package intent

import "strings"

// IntentPlan 大模型输出的意图路由结果。
type IntentPlan struct {
	Intent string `json:"intent"`
	Reason string `json:"reason"`
}

func BuildIntentEnumForPrompt(intents []string) string {
	if len(intents) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(intents))
	for _, v := range intents {
		quoted = append(quoted, `"`+v+`"`)
	}
	return strings.Join(quoted, "|")
}
