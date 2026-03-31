package client

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	candidateTokenRe = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_-]{3,}`)
)

func JsonRaw(b []byte) interface{} {
	if len(b) == 0 {
		return nil
	}
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return string(b)
	}
	return v
}

func ExtractFrontURL(raw []byte) string {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return ""
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].(map[string]interface{})
	url, _ := data["url"].(string)
	return strings.TrimSpace(url)
}

func ExtractFrontFilePath(raw []byte) string {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return ""
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].(map[string]interface{})
	fp, _ := data["file_path"].(string)
	return strings.TrimSpace(fp)
}

func ExtractAnyURL(raw []byte) string {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return ""
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].(map[string]interface{})
	candidates := []string{"file_path", "url", "pdf_url", "download_url"}
	for _, k := range candidates {
		if v, ok := data[k].(string); ok && strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

func ExtractFirstIDFromFrontList(raw []byte) int {
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return 0
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return 0
	}
	first, _ := data[0].(map[string]interface{})
	if first == nil {
		return 0
	}
	switch t := first["id"].(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	default:
		return 0
	}
}

func ToInt(v interface{}) (int, bool) {
	switch t := v.(type) {
	case int:
		return t, true
	case int64:
		return int(t), true
	case float64:
		return int(t), true
	default:
		return 0, false
	}
}

func CollectIDsByProjectNumber(items []interface{}, number string) []int {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil
	}
	var ids []int
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		n, _ := m["number"].(string)
		if strings.EqualFold(strings.TrimSpace(n), number) {
			if id, ok := ToInt(m["id"]); ok && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func ExtractNumberCandidates(question string) []string {
	raw := candidateTokenRe.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, v := range raw {
		key := strings.ToUpper(strings.TrimSpace(v))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func CollectIDsByContractNumber(items []interface{}, number string) []int {
	number = strings.TrimSpace(number)
	if number == "" {
		return nil
	}
	var ids []int
	for _, it := range items {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		n, _ := m["contract_number"].(string)
		if strings.EqualFold(strings.TrimSpace(n), number) {
			if id, ok := ToInt(m["id"]); ok && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func ExtractContractNumberCandidates(question string) []string {
	raw := candidateTokenRe.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, v := range raw {
		key := strings.ToUpper(strings.TrimSpace(v))
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}
