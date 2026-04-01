package intent

import "strings"

func normalizeIntentQuestion(question string) string {
	replacer := strings.NewReplacer(
		"工单", "任务",
		"单号", "任务编号",
		"流水号", "任务编号",
		"uuid", "任务编号",
		"合约", "合同",
		"源数据", "原始数据",
		"原始文件", "原始数据",
		"原文件", "原始数据",
		"结果包", "结果",
		"压缩包", "zip",
		"报告包", "结题报告",
		"跑完没", "有没有完成",
		"做完没", "有没有完成",
		"出结果没", "有没有完成",
		"挂了", "失败",
		"报错", "失败原因",
	)
	return replacer.Replace(strings.ToLower(strings.TrimSpace(question)))
}

func heuristicProjectIntent(question string) *IntentPlan {
	q := normalizeIntentQuestion(question)
	switch {
	case strings.Contains(q, "原始数据"), strings.Contains(q, "提取码"), strings.Contains(q, "授权码"):
		return &IntentPlan{Intent: IntentDownloadOriginal, Reason: "命中原始数据下载口语规则"}
	case strings.Contains(q, "结题报告"), strings.Contains(q, "质控报告"), strings.Contains(q, "zip"), strings.Contains(q, "报告下载"), strings.Contains(q, "下载报告"):
		return &IntentPlan{Intent: IntentDownloadFinal, Reason: "命中报告下载口语规则"}
	case strings.Contains(q, "合同") && strings.Contains(q, "项目"):
		return &IntentPlan{Intent: IntentContractProjects, Reason: "命中合同下项目查询规则"}
	case strings.Contains(q, "合同"):
		return &IntentPlan{Intent: IntentContractList, Reason: "命中合同查询规则"}
	case strings.Contains(q, "到期"), strings.Contains(q, "快过期"), strings.Contains(q, "剩余期限"):
		return &IntentPlan{Intent: IntentProjectExpiring, Reason: "命中项目到期规则"}
	case strings.Contains(q, "项目"), strings.Contains(q, "我的单子"), strings.Contains(q, "我的项目"):
		return &IntentPlan{Intent: IntentProjectList, Reason: "命中项目列表规则"}
	default:
		return nil
	}
}

func heuristicTaskIntent(question string) *IntentPlan {
	q := normalizeIntentQuestion(question)
	switch {
	case strings.Contains(q, "下载"), strings.Contains(q, "导出"), strings.Contains(q, "结果"), strings.Contains(q, "zip"):
		return &IntentPlan{Intent: IntentTaskDownloadResult, Reason: "命中任务结果下载口语规则"}
	case strings.Contains(q, "有没有完成"), strings.Contains(q, "成功还是失败"), strings.Contains(q, "失败原因"), strings.Contains(q, "状态"), strings.Contains(q, "进度"):
		return &IntentPlan{Intent: IntentTaskStatus, Reason: "命中任务状态口语规则"}
	default:
		return nil
	}
}

func heuristicProjectArticleIntent(question string) *IntentPlan {
	q := normalizeIntentQuestion(question)
	switch {
	case strings.Contains(q, "文章"), strings.Contains(q, "文献"), strings.Contains(q, "期刊"):
		return &IntentPlan{Intent: IntentProjectArticleList, Reason: "命中文章查询口语规则"}
	default:
		return nil
	}
}
