package limitcap

import "encoding/json"

// CapCloudListResponseJSON 若 JSON 含 content.data 数组且长度大于 limit，则仅保留前 limit 条后重新序列化；
// 供接口默认展示。返回 (展示用字节, data 总条数)；无法识别列表结构时返回原始 raw 与 total=0。
func CapCloudListResponseJSON(raw []byte, limit int) ([]byte, int) {
	if limit <= 0 {
		limit = DefaultVisibleItems
	}
	if len(raw) == 0 {
		return raw, 0
	}
	var root map[string]interface{}
	if err := json.Unmarshal(raw, &root); err != nil {
		return raw, 0
	}
	content, _ := root["content"].(map[string]interface{})
	if content == nil {
		return raw, 0
	}
	data, ok := content["data"].([]interface{})
	if !ok {
		return raw, 0
	}
	total := len(data)
	if total <= limit {
		return raw, total
	}
	content["data"] = data[:limit]
	out, err := json.Marshal(root)
	if err != nil {
		return raw, total
	}
	return out, total
}
