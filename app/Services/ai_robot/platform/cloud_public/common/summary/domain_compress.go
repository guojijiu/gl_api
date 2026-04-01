package summary

import "encoding/json"

func compressProjectJSONForPrompt(apiJSON []byte) string {
	var root map[string]interface{}
	if err := json.Unmarshal(apiJSON, &root); err != nil {
		return compressAPIJSONForPrompt(apiJSON)
	}
	if cps, ok := root["contract_projects"].([]interface{}); ok {
		out := map[string]interface{}{
			"domain":            "project_contract_projects",
			"contract_projects": compressPromptArray(cps, 0, "contract_projects"),
		}
		return mustJSON(out, apiJSON)
	}
	if links, ok := root["links"].([]interface{}); ok {
		out := map[string]interface{}{
			"domain": "project_download",
			"links":  compressPromptArray(links, 0, "links"),
		}
		if raw, exists := root["raw"]; exists {
			out["raw"] = compressPromptValue(raw, 0, "raw")
		}
		return mustJSON(out, apiJSON)
	}
	return compressListResponseForPrompt(apiJSON, "project", []string{"id", "number", "name", "workflow_name_cn", "effective_end_at", "contract_id"})
}

func compressTaskJSONForPrompt(apiJSON []byte) string {
	var root map[string]interface{}
	if err := json.Unmarshal(apiJSON, &root); err != nil {
		return compressAPIJSONForPrompt(apiJSON)
	}
	if details, ok := root["details"].([]interface{}); ok {
		out := map[string]interface{}{
			"domain":       "task_status",
			"uuids":        root["uuids"],
			"status_batch": compressPromptValue(root["status_batch"], 0, "status_batch"),
			"details":      compressPromptArray(details, 0, "details"),
		}
		return mustJSON(out, apiJSON)
	}
	if links, ok := root["links"].([]interface{}); ok {
		out := map[string]interface{}{
			"domain": "task_download",
			"uuids":  root["uuids"],
			"links":  compressPromptArray(links, 0, "links"),
		}
		if raw, exists := root["raw"]; exists {
			out["raw"] = compressPromptValue(raw, 0, "raw")
		}
		return mustJSON(out, apiJSON)
	}
	return compressListResponseForPrompt(apiJSON, "task", []string{"id", "uuid", "name", "status", "status_value", "project_id", "workflow_name_cn", "file_path"})
}

func compressProjectArticleJSONForPrompt(apiJSON []byte) string {
	return compressListResponseForPrompt(apiJSON, "project_article", []string{
		"name_cn", "name_en", "journal_name", "publish_date", "product_category",
		"product_label", "species_name", "species_category", "sample_type", "region",
	})
}

func compressListResponseForPrompt(apiJSON []byte, domain string, fields []string) string {
	var root map[string]interface{}
	if err := json.Unmarshal(apiJSON, &root); err != nil {
		return compressAPIJSONForPrompt(apiJSON)
	}
	content, _ := root["content"].(map[string]interface{})
	data, _ := content["data"].([]interface{})
	items := make([]map[string]interface{}, 0, min(maxPromptArraySamples, len(data)))
	for i := 0; i < min(maxPromptArraySamples, len(data)); i++ {
		item, _ := data[i].(map[string]interface{})
		if item == nil {
			continue
		}
		row := make(map[string]interface{})
		for _, field := range fields {
			if value, ok := item[field]; ok {
				row[field] = compressPromptValue(value, 0, field)
			}
		}
		items = append(items, row)
	}
	out := map[string]interface{}{
		"domain":  domain,
		"code":    root["code"],
		"showMsg": root["showMsg"],
		"total": firstNonNil(
			content["total"],
			root["total"],
			len(data),
		),
		"items_sample":   items,
		"omitted_count":  max(0, len(data)-len(items)),
		"truncated_note": truncatedNoteForArray("data", len(data)),
	}
	return mustJSON(out, apiJSON)
}

func firstNonNil(values ...interface{}) interface{} {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}
