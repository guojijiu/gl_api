package summary

import (
	"encoding/json"
	"sort"
	"strings"
)

const (
	maxPromptArraySamples = 5
	maxPromptMapKeys      = 16
	maxPromptDepth        = 4
	maxPromptStringLen    = 500
)

var promptPriorityKeys = []string{
	"code", "showMsg", "debugMsg", "message",
	"content", "data", "total", "page", "size", "current_page", "last_page",
	"id", "uuid", "number", "contract_number", "name", "name_cn", "name_en",
	"status", "status_value", "workflow_name_cn", "project_id", "task_id",
	"url", "file_path", "access_code", "publish_date", "journal_name",
}

func compressAPIJSONForPrompt(apiJSON []byte) string {
	if len(apiJSON) == 0 {
		return ""
	}
	var v interface{}
	if err := json.Unmarshal(apiJSON, &v); err != nil {
		return truncateAPIJSONForPrompt(string(apiJSON))
	}
	compressed := compressPromptValue(v, 0, "")
	b, err := json.Marshal(compressed)
	if err != nil {
		return truncateAPIJSONForPrompt(string(apiJSON))
	}
	return truncateAPIJSONForPrompt(string(b))
}

func compressPromptValue(v interface{}, depth int, parentKey string) interface{} {
	if depth >= maxPromptDepth {
		switch t := v.(type) {
		case []interface{}:
			// 层级过深时不再展开元素，只保留数量，避免多层嵌套与多种 truncated 语义。
			return map[string]interface{}{
				"count": len(t),
				"note":  "内容层级过深，已省略明细，仅保留条数",
			}
		case map[string]interface{}:
			return map[string]interface{}{
				"key_count": len(t),
				"note":      "内容层级过深，已省略字段明细",
			}
		case string:
			return truncatePromptString(t)
		default:
			return t
		}
	}

	switch t := v.(type) {
	case map[string]interface{}:
		return compressPromptMap(t, depth)
	case []interface{}:
		return compressPromptArray(t, depth)
	case string:
		return truncatePromptString(t)
	default:
		return t
	}
}

func compressPromptMap(m map[string]interface{}, depth int) map[string]interface{} {
	result := make(map[string]interface{})
	used := make(map[string]struct{})

	for _, key := range promptPriorityKeys {
		value, ok := m[key]
		if !ok {
			continue
		}
		result[key] = compressPromptValue(value, depth+1, key)
		used[key] = struct{}{}
		if len(result) >= maxPromptMapKeys {
			break
		}
	}

	if len(result) < maxPromptMapKeys {
		keys := make([]string, 0, len(m))
		for key := range m {
			if _, ok := used[key]; ok {
				continue
			}
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			result[key] = compressPromptValue(m[key], depth+1, key)
			if len(result) >= maxPromptMapKeys {
				break
			}
		}
	}

	return result
}

func compressPromptArray(items []interface{}, depth int) interface{} {
	samples := min(maxPromptArraySamples, len(items))
	sampleItems := make([]interface{}, 0, samples)
	for i := 0; i < samples; i++ {
		sampleItems = append(sampleItems, compressPromptValue(items[i], depth+1, ""))
	}
	return map[string]interface{}{
		"count":          len(items),
		"sample":         sampleItems,
		"omitted_count":  max(0, len(items)-samples),
		"truncated_note": truncatedNoteForArray(len(items)),
	}
}

func truncatedNoteForArray(total int) string {
	if total <= maxPromptArraySamples {
		return ""
	}
	return "列表总条数大于默认展示上限（5 条），仅抽取前 5 条供分析；请在答复末尾用一句话提示用户可回复「查看更多」「下一页」，或在会话消息列表页面分页查看完整数据"
}

func truncatePromptString(s string) string {
	s = strings.TrimSpace(s)
	if len(s) <= maxPromptStringLen {
		return s
	}
	return s[:maxPromptStringLen] + "...[truncated]"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
