package aigateway

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

func (s *Service) FetchUserProjects(ctx context.Context, platform string, userToken string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.ProjectGetUserAllProjectURL(platform), userToken)
}

func (s *Service) FetchProjectZipURL(ctx context.Context, platform string, userToken string, projectID int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.ProjectGetZipURL(platform, projectID), userToken)
}

func (s *Service) FetchOriginalDataURL(ctx context.Context, platform string, userToken string, dataID int, accessCode string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.ProjectGetOriginalDataURL(platform, dataID, accessCode), userToken)
}

func (s *Service) FetchContracts(ctx context.Context, platform string, userToken string, contractNumber string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.ContractListOfProjectURL(platform, 1, 200, contractNumber), userToken)
}

func (s *Service) FetchProjectsByContractID(ctx context.Context, platform string, userToken string, contractID int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.ProjectListByContractURL(platform, contractID, 1, 1000), userToken)
}

func (s *Service) ResolveContractIDsFromContracts(ctx context.Context, userQuestion string, contractsJSON []byte) ([]int, string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(contractsJSON, &root); err != nil {
		return nil, "", fmt.Errorf("解析合同列表失败: %w", err)
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return nil, "", errors.New("未查询到可用合同")
	}
	candidates := extractContractNumberCandidates(userQuestion)
	if len(candidates) > 0 {
		for _, c := range candidates {
			ids := collectIDsByContractNumber(data, c)
			if len(ids) > 0 {
				return ids, c, nil
			}
		}
		if s.cfg.StrictMode {
			return nil, "", errors.New("严格模式已开启：未按合同编号匹配到数据，请确认合同编号后重试")
		}
	} else if s.cfg.StrictMode {
		return nil, "", errors.New("严格模式已开启：请在问题中明确提供合同编号")
	}
	if len(data) == 1 {
		if m, ok := data[0].(map[string]interface{}); ok {
			if id, ok := toInt(m["id"]); ok && id > 0 {
				return []int{id}, "", nil
			}
		}
	}
	return nil, "", errors.New("无法从问题中匹配合同，请补充合同编号")
}

func (s *Service) ResolveProjectIDsFromProjects(ctx context.Context, userQuestion string, projectsJSON []byte) ([]int, string, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(projectsJSON, &root); err != nil {
		return nil, "", fmt.Errorf("解析项目列表失败: %w", err)
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	if len(data) == 0 {
		return nil, "", errors.New("未查询到可用项目")
	}
	candidates := extractNumberCandidates(userQuestion)
	if len(candidates) > 0 {
		for _, c := range candidates {
			ids := collectIDsByProjectNumber(data, c)
			if len(ids) > 0 {
				return ids, c, nil
			}
		}
		if s.cfg.StrictMode {
			return nil, "", errors.New("严格模式已开启：未按项目编号匹配到数据，请确认项目编号后重试")
		}
	} else if s.cfg.StrictMode {
		return nil, "", errors.New("严格模式已开启：请在问题中明确提供项目编号")
	}
	if len(data) == 1 {
		if m, ok := data[0].(map[string]interface{}); ok {
			if id, ok := toInt(m["id"]); ok && id > 0 {
				return []int{id}, "", nil
			}
		}
	}
	prompt := fmt.Sprintf(`你是项目匹配器。根据用户问题，从给定项目列表中选出最可能的 project_id。
用户问题：%s
项目列表JSON：%s

只输出 JSON：
{"project_id":123,"reason":"不超过60字"}
若无法判断，project_id 输出 0。`, userQuestion, string(projectsJSON))
	text, err := s.chatCompletion(ctx, prompt)
	if err != nil {
		return nil, "", err
	}
	raw := extractJSONFromLLM(text)
	var parsed struct {
		ProjectID int    `json:"project_id"`
		Reason    string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, "", fmt.Errorf("解析 project_id 失败: %w", err)
	}
	if parsed.ProjectID <= 0 {
		return nil, "", errors.New("无法从问题中匹配项目，请在问题中补充项目编号")
	}
	return []int{parsed.ProjectID}, "", nil
}

func (s *Service) FilterExpiringProjects(apiBody []byte) ([]byte, error) {
	var root map[string]interface{}
	if err := json.Unmarshal(apiBody, &root); err != nil {
		return apiBody, nil
	}
	content, ok := root["content"].(map[string]interface{})
	if !ok {
		return apiBody, nil
	}
	data, ok := content["data"].([]interface{})
	if !ok || len(data) == 0 {
		return apiBody, nil
	}
	days := s.cfg.ExpiringWithinDays
	if days <= 0 {
		days = 30
	}
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	last := today.AddDate(0, 0, days)
	var kept []interface{}
	for _, it := range data {
		m, ok := it.(map[string]interface{})
		if !ok {
			continue
		}
		ends, _ := m["effective_end_at"].(string)
		if ends == "" {
			continue
		}
		endDay, err := time.ParseInLocation("2006-01-02", ends, time.Local)
		if err != nil {
			continue
		}
		if endDay.Before(today) || endDay.After(last) {
			continue
		}
		kept = append(kept, it)
	}
	content["data"] = kept
	root["content"] = content
	out, err := json.Marshal(root)
	if err != nil {
		return apiBody, nil
	}
	return out, nil
}

func toInt(v interface{}) (int, bool) {
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

func collectIDsByProjectNumber(items []interface{}, number string) []int {
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
			if id, ok := toInt(m["id"]); ok && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func extractNumberCandidates(question string) []string {
	re := regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_-]{3,}`)
	raw := re.FindAllString(question, -1)
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

func collectIDsByContractNumber(items []interface{}, number string) []int {
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
			if id, ok := toInt(m["id"]); ok && id > 0 {
				ids = append(ids, id)
			}
		}
	}
	return ids
}

func extractContractNumberCandidates(question string) []string {
	re := regexp.MustCompile(`[A-Za-z0-9][A-Za-z0-9_-]{3,}`)
	raw := re.FindAllString(question, -1)
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
