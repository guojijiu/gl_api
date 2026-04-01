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
			return map[string]interface{}{
				"count":          len(t),
				"omitted_count":  max(0, len(t)),
				"truncated_note": "depth_limit_reached",
			}
		case map[string]interface{}:
			return map[string]interface{}{
				"keys":           len(t),
				"truncated_note": "depth_limit_reached",
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
		return compressPromptArray(t, depth, parentKey)
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

	if len(m) > len(result) {
		result["_omitted_keys_count"] = len(m) - len(result)
	}
	return result
}

func compressPromptArray(items []interface{}, depth int, parentKey string) interface{} {
	samples := min(maxPromptArraySamples, len(items))
	sampleItems := make([]interface{}, 0, samples)
	for i := 0; i < samples; i++ {
		sampleItems = append(sampleItems, compressPromptValue(items[i], depth+1, parentKey))
	}
	return map[string]interface{}{
		"count":          len(items),
		"sample":         sampleItems,
		"omitted_count":  max(0, len(items)-samples),
		"truncated_note": truncatedNoteForArray(parentKey, len(items)),
	}
}

func truncatedNoteForArray(parentKey string, total int) string {
	if total <= maxPromptArraySamples {
		return ""
	}
	switch parentKey {
	case "data":
		return "data_array_compressed_for_prompt"
	case "details":
		return "details_array_compressed_for_prompt"
	case "links":
		return "links_array_compressed_for_prompt"
	default:
		return "array_compressed_for_prompt"
	}
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
