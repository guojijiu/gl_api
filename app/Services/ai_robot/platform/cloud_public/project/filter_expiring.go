package project

import (
	"encoding/json"
	"time"

	"cloud-platform-api/app/Config"
)

// FilterExpiringProjects 按配置天数过滤即将到期的项目列表 JSON。
func FilterExpiringProjects(cfg *Config.AiGatewayConfig, apiBody []byte) ([]byte, error) {
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
	days := 30
	if cfg != nil && cfg.ExpiringWithinDays > 0 {
		days = cfg.ExpiringWithinDays
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
