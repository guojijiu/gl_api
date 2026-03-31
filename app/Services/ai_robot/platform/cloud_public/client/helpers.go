package cloudclient

import (
	"regexp"
	"strings"
)

var (
	// 复用同一套 token 候选提取逻辑，避免每次调用都 regexp.MustCompile。
	candidateTokenRe = regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_-]{3,}`)
)

// ToInt、CollectIDs*、Extract*Candidates 供 domain 层解析列表与问题文本复用。
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
