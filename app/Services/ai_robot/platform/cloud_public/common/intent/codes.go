package intent

// 本文件为与 LLM / 路由层约定的意图标识符及合法取值范围，属于「协议层常量」，
// 不是部署配置（如 AiGatewayConfig）；业务侧按 intent 字符串分支即可。

const (
	IntentUnsupported = "unsupported"

	IntentProjectList      = "project_list"
	IntentProjectExpiring  = "project_expiring"
	IntentContractList     = "contract_list"
	IntentContractProjects = "contract_projects"
	IntentDownloadFinal    = "download_final_report"
	IntentDownloadOriginal = "download_original_data"

	IntentTaskDownloadResult = "task_download_result"
	IntentTaskStatus         = "task_status"

	IntentProjectArticleList = "project_article_list"
)

var ProjectIntentWhitelist = []string{
	IntentProjectList,
	IntentProjectExpiring,
	IntentContractList,
	IntentContractProjects,
	IntentDownloadFinal,
	IntentDownloadOriginal,
	IntentUnsupported,
}

var TaskIntentWhitelist = []string{
	IntentTaskDownloadResult,
	IntentTaskStatus,
	IntentUnsupported,
}

var ProjectArticleIntentWhitelist = []string{
	IntentProjectArticleList,
	IntentUnsupported,
}
