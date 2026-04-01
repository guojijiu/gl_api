package parse

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"cloud-platform-api/app/Services/ai_robot/internal/filters"
)

var (
	explicitValuePatterns = []*regexp.Regexp{
		regexp.MustCompile(`^(?:为|是|叫|名称是|名为)[:：\s]*[\""'“”‘’]?([^，。,.\s\""'“”‘’]{2,64})[\""'“”‘’]?`),
		regexp.MustCompile(`^[:：\s]+[\""'“”‘’]?([^，。,.\s\""'“”‘’]{2,64})[\""'“”‘’]?`),
	}
	quotedValueRe  = regexp.MustCompile(`^\s*[\"“”‘']([^\"“”‘'\r\n]{2,64})[\"“”‘']`)
	recentDaysRe   = regexp.MustCompile(`最近\s*(\d{1,3})\s*天`)
	recentWeeksRe  = regexp.MustCompile(`最近\s*(\d{1,2})\s*周`)
	recentMonthsRe = regexp.MustCompile(`最近\s*(\d{1,2})\s*个?月`)
	monthRangeRe   = regexp.MustCompile(`(\d{1,2})\s*月`)
)

var aliasGroups = [][]string{
	{"项目", "单子", "项目单"},
	{"合同", "合约"},
	{"任务", "工单", "分析任务", "跑的任务"},
	{"任务编号", "uuid", "任务id", "工单号", "单号", "流水号"},
	{"项目编号", "项目号", "项目编码"},
	{"合同编号", "合同号", "合约号"},
	{"任务名称", "工单名称"},
	{"产品名称", "工作流名称"},
	{"工具名称", "分析工具"},
	{"项目标签", "标签"},
	{"物种分类", "物种大类"},
	{"样本类型", "样本类别"},
	{"研究方向", "研究方向-选择"},
	{"发表单位", "单位"},
	{"区域", "地区"},
}

func ExtractTaskUUIDsFromQuestion(question string) []string {
	re := regexp.MustCompile(`(?i)\b[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}\b`)
	raw := re.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	var out []string
	for _, v := range raw {
		key := strings.ToLower(v)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, v)
	}
	return out
}

func ExtractNumericIDsFromQuestion(question string) []int {
	re := regexp.MustCompile(`\b\d{3,}\b`)
	raw := re.FindAllString(question, -1)
	if len(raw) == 0 {
		return nil
	}
	seen := map[int]struct{}{}
	var out []int
	for _, v := range raw {
		var n int
		_, err := fmt.Sscanf(v, "%d", &n)
		if err != nil || n <= 0 {
			continue
		}
		if _, ok := seen[n]; ok {
			continue
		}
		seen[n] = struct{}{}
		out = append(out, n)
	}
	return out
}

func extractExplicitValue(question string, keywords ...string) string {
	for _, keyword := range keywords {
		idx := strings.Index(question, keyword)
		if idx < 0 {
			continue
		}
		rest := strings.TrimSpace(question[idx+len(keyword):])
		for _, re := range explicitValuePatterns {
			if matches := re.FindStringSubmatch(rest); len(matches) == 2 {
				return strings.TrimSpace(matches[1])
			}
		}
		if matches := quotedValueRe.FindStringSubmatch(rest); len(matches) == 2 {
			return strings.TrimSpace(matches[1])
		}
	}
	return ""
}

func normalizeQuestion(question string) string {
	normalized := question
	for _, group := range aliasGroups {
		canonical := group[0]
		for _, alias := range group[1:] {
			normalized = strings.ReplaceAll(normalized, alias, canonical)
		}
	}
	return normalized
}

func extractDateRange(question string) (string, string) {
	now := time.Now()
	if matches := recentDaysRe.FindStringSubmatch(question); len(matches) == 2 {
		var days int
		_, _ = fmt.Sscanf(matches[1], "%d", &days)
		if days > 0 {
			start := now.AddDate(0, 0, -days+1).Format("2006-01-02")
			return start, now.Format("2006-01-02")
		}
	}
	if matches := recentWeeksRe.FindStringSubmatch(question); len(matches) == 2 {
		var weeks int
		_, _ = fmt.Sscanf(matches[1], "%d", &weeks)
		if weeks > 0 {
			start := now.AddDate(0, 0, -(weeks*7)+1).Format("2006-01-02")
			return start, now.Format("2006-01-02")
		}
	}
	if matches := recentMonthsRe.FindStringSubmatch(question); len(matches) == 2 {
		var months int
		_, _ = fmt.Sscanf(matches[1], "%d", &months)
		if months > 0 {
			start := now.AddDate(0, -months, 1).Format("2006-01-02")
			return start, now.Format("2006-01-02")
		}
	}
	if matches := monthRangeRe.FindStringSubmatch(question); len(matches) == 2 {
		var month int
		_, _ = fmt.Sscanf(matches[1], "%d", &month)
		if month >= 1 && month <= 12 {
			year := now.Year()
			start := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, now.Location())
			end := start.AddDate(0, 1, -1)
			return start.Format("2006-01-02"), end.Format("2006-01-02")
		}
	}
	return "", ""
}

func buildGuidance(prefix string, examples ...string) string {
	if len(examples) == 0 {
		return prefix
	}
	return prefix + "。例如：" + strings.Join(examples, "；")
}

func hasAny(question string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(question, keyword) {
			return true
		}
	}
	return false
}

func hasTaskSignals(question string) bool {
	return len(ExtractTaskUUIDsFromQuestion(question)) > 0 ||
		hasAny(normalizeQuestion(question), "任务", "任务编号", "状态", "进度", "失败原因", "有没有完成", "结果", "跑完没", "做完没", "出结果没")
}

func hasProjectSignals(question string) bool {
	return hasAny(normalizeQuestion(question), "项目", "项目编号", "结题报告", "原始数据", "质控报告", "合同")
}

func hasArticleSignals(question string) bool {
	return hasAny(normalizeQuestion(question), "文章", "文献", "期刊")
}

func ExtractProjectFiltersFromQuestion(question string) filters.ProjectFilters {
	question = normalizeQuestion(question)
	result := filters.ProjectFilters{IsFilterTime: 2}
	if v := extractExplicitValue(question, "项目编号"); v != "" {
		result.Number = v
	} else if strings.Contains(question, "项目") {
		if nums := ExtractTaskLikeCodes(question); len(nums) > 0 {
			result.Number = nums[0]
		}
	}
	result.Name = extractExplicitValue(question, "项目名称", "项目名")
	result.WorkflowNameCN = extractExplicitValue(question, "产品名称", "流程名称", "流程名", "工作流名称")
	return result
}

func ExtractContractFiltersFromQuestion(question string) filters.ContractFilters {
	question = normalizeQuestion(question)
	result := filters.ContractFilters{}
	if v := extractExplicitValue(question, "合同编号"); v != "" {
		result.ContractNumber = v
	} else if strings.Contains(question, "合同") {
		if nums := ExtractContractLikeCodes(question); len(nums) > 0 {
			result.ContractNumber = nums[0]
		}
	}
	result.Name = extractExplicitValue(question, "合同名称", "合同名")
	return result
}

func ExtractTaskFiltersFromQuestion(question string, uuid string) filters.TaskFilters {
	question = normalizeQuestion(question)
	result := filters.TaskFilters{UUID: uuid}
	result.Name = extractExplicitValue(question, "任务名称", "任务名")
	result.ToolName = extractExplicitValue(question, "工具名称", "工具名")
	result.ProjectNumber = extractExplicitValue(question, "项目编号")
	result.ProjectName = extractExplicitValue(question, "项目名称", "项目名")
	result.WorkflowNameCN = extractExplicitValue(question, "产品名称", "流程名称", "流程名")
	result.CreatedAtStart, result.CreatedAtEnd = extractDateRange(question)
	switch {
	case strings.Contains(question, "成功"), strings.Contains(question, "完成"):
		result.StatusValue = "success"
	case strings.Contains(question, "失败"), strings.Contains(question, "挂起"), strings.Contains(question, "异常"):
		result.StatusValue = "failed"
	case strings.Contains(question, "处理中"), strings.Contains(question, "运行中"), strings.Contains(question, "进度"):
		result.StatusValue = "running"
	}
	return result
}

func ExtractProjectArticleFiltersFromQuestion(question string) filters.ProjectArticleFilters {
	question = normalizeQuestion(question)
	var items []filters.ProjectArticleFilterItem
	if v := extractExplicitValue(question, "项目编号"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "number", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "中文名称", "中文名"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "name_cn", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "英文名称", "英文名"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "name_en", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "期刊名称", "期刊名", "期刊"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "journal_name", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "产品分类"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "product_category", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "项目标签", "标签"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "product_label", Operator: "json_in", Value: v})
	}
	if v := extractExplicitValue(question, "物种分类"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "species_category", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "物种名称", "物种名"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "species_name", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "样本类型", "样本"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "sample_type", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "研究方向"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "research_direction_select", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "发表单位", "单位"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "publish_unit", Operator: "like", Value: v})
	}
	if v := extractExplicitValue(question, "区域", "地区"); v != "" {
		items = append(items, filters.ProjectArticleFilterItem{Column: "region", Operator: "like", Value: v})
	}
	return filters.ProjectArticleFilters{SearchFilter: items}
}

func ExtractTaskLikeCodes(question string) []string {
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

func ExtractContractLikeCodes(question string) []string {
	return ExtractTaskLikeCodes(question)
}

func ProjectMatchGuidance() string {
	return buildGuidance("我暂时没能准确定位到你说的项目，请尽量补充项目编号或项目名称", "查询项目编号 MWXS-25-10500-a", "项目名称为 乳酸菌代谢组 的项目")
}

func ProjectMatchGuidanceForQuestion(question string) string {
	if hasTaskSignals(question) && !hasProjectSignals(question) {
		return buildGuidance("你现在这句话更像是在查任务，不是查项目。如果你要查任务，请补充任务编号", "查询任务编号 86ce0f8b-ee05-47af-9495-6a48addd8a39 的状态", "下载这个工单 86ce0f8b-ee05-47af-9495-6a48addd8a39 的结果")
	}
	if hasArticleSignals(question) && !hasProjectSignals(question) {
		return buildGuidance("你现在这句话更像是在查项目文章。如果你要查文章，请补充期刊名、标签、物种或文章名称", "查期刊名为 Nature 的文章", "查标签为 植物测试 的项目文章")
	}
	return ProjectMatchGuidance()
}

func ContractMatchGuidance() string {
	return buildGuidance("我暂时没能准确定位到你说的合同，请尽量补充合同编号或合同名称", "查询合同编号 HT20250301", "合同名称为 玉米代谢组服务 的合同")
}

func ContractMatchGuidanceForQuestion(question string) string {
	if hasTaskSignals(question) && !hasProjectSignals(question) {
		return TaskUUIDGuidanceForQuestion(question)
	}
	return ContractMatchGuidance()
}

func TaskUUIDGuidance() string {
	return buildGuidance("我还不能直接判断你说的是哪一个任务，请补充任务编号", "查询任务编号 86ce0f8b-ee05-47af-9495-6a48addd8a39 的状态", "下载这个工单 86ce0f8b-ee05-47af-9495-6a48addd8a39 的结果")
}

func TaskUUIDGuidanceForQuestion(question string) string {
	if hasProjectSignals(question) && !hasTaskSignals(question) {
		return buildGuidance("你现在这句话更像是在查项目资料。如果你要查项目，请补充项目编号或项目名称", "下载项目编号 MWXS-25-10500-a 的结题报告", "下载项目编号 MWXS-25-10500-a 的原始数据")
	}
	if hasArticleSignals(question) && !hasTaskSignals(question) {
		return buildGuidance("你现在这句话更像是在查项目文章，不是任务。如果你要查文章，请补充文章名称、期刊名或标签", "查期刊名为 Cell 的项目文章", "查标签为 植物测试 的项目文章")
	}
	return TaskUUIDGuidance()
}
