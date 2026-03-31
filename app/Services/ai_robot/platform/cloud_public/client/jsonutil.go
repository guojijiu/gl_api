package cloudclient

import (
	"encoding/json"
	"strings"
)

// JsonRaw 将云平台 JSON 字节解析为 interface{}，供 gin 等序列化；失败时退回原始字符串。
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

// ExtractFrontURL 从典型 content.data.url 结构取 URL。
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

// ExtractFrontFilePath 从典型 content.data.file_path 取路径。
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

// ExtractAnyURL 从 content.data 中常见下载相关字段取首个非空 URL。
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

// ExtractFirstIDFromFrontList 从 content.data 列表首项取 id。
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
