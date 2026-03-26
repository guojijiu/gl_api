package aigateway

import (
	"encoding/json"
	"strings"
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
