package cloudapi

import (
	"fmt"
	"net/url"
	"strings"
)

func ProjectGetUserAllProjectURL(base string) string {
	return base + "/api/front/project/get_user_all_project"
}

func ProjectGetZipURL(base string, projectID int) string {
	return base + "/api/front/project/zip_url?id=" + fmt.Sprintf("%d", projectID)
}

func ProjectGetOriginalDataURL(base string, dataID int, accessCode string) string {
	return base + "/api/front/project/original_data_url?id=" + fmt.Sprintf("%d", dataID) +
		"&access_code=" + url.QueryEscape(accessCode)
}

func ContractListOfProjectURL(base string, page, size int, contractNumber string) string {
	u := base + "/api/front/contract/list_of_project?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(contractNumber) != "" {
		u += "&contract_number=" + url.QueryEscape(contractNumber)
	}
	return u
}

func ProjectListByContractURL(base string, contractID, page, size int) string {
	return base + "/api/front/project/list?contract_id=" + fmt.Sprintf("%d", contractID) +
		"&page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
}

func TaskListURL(base string, page, size int, uuid string) string {
	u := base + "/api/front/task/list?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		u += "&uuid=" + url.QueryEscape(uuid)
	}
	return u
}

func TaskStatusByUUIDsURL(base, uuids string) string {
	return base + "/api/front/task/status_by_uuids?uuids=" + url.QueryEscape(uuids)
}

func TaskGetResultURL(base string, taskID int) string {
	return base + "/api/front/task/result?id=" + fmt.Sprintf("%d", taskID)
}

func TaskDownloadResultURL(base string) string {
	return base + "/api/front/task/download_result"
}

func TaskListOfWorkflowURL(base string, page, size int, uuid string) string {
	u := base + "/api/front/task/list_of_workflow?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		u += "&uuid=" + url.QueryEscape(uuid)
	}
	return u
}

func TaskDetailOfModuleToolURL(base string, id int) string {
	return base + "/api/front/task/detail_of_module_tool?id=" + fmt.Sprintf("%d", id)
}

func TaskPdfURLOfModuleToolURL(base string, taskID int) string {
	return base + "/api/front/task/pdf_url_of_module_tool?type=2&source_id=" + fmt.Sprintf("%d", taskID)
}

func ProjectArticleListURL(base string) string {
	return base + "/api/front/project_article/list"
}
