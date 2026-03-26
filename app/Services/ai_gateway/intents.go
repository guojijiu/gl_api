package aigateway

import "strings"

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

func BuildIntentEnumForPrompt(intents []string) string {
	if len(intents) == 0 {
		return ""
	}
	quoted := make([]string, 0, len(intents))
	for _, v := range intents {
		quoted = append(quoted, `"`+v+`"`)
	}
	return strings.Join(quoted, "|")
}
