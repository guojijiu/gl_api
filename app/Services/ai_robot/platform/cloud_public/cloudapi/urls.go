// Package cloudapi 拼装 Laravel 云平台 front 路由完整 URL（与 routes/front.php 对应）。
// 公网 cloud_public 与内网 cloud_intranet 共用同一套路径，仅站点根地址（base）不同；
// base 须为已去掉末尾斜杠的根 URL，与 Config.AiGatewayConfig.CloudAPIBaseByPlatform 返回值一致。
package cloudapi

import (
	"fmt"
	"net/url"
	"strings"
)

// ProjectGetUserAllProjectURL Laravel：project/get_user_all_project。
func ProjectGetUserAllProjectURL(base string) string {
	return base + "/api/front/project/get_user_all_project"
}

// ProjectGetZipURL 结题报告下载（front.php: project/zip_url），query：id。
func ProjectGetZipURL(base string, projectID int) string {
	return base + "/api/front/project/zip_url?id=" + fmt.Sprintf("%d", projectID)
}

// ProjectGetOriginalDataURL 原始数据下载（front.php: project/original_data_url）。
func ProjectGetOriginalDataURL(base string, dataID int, accessCode string) string {
	return base + "/api/front/project/original_data_url?id=" + fmt.Sprintf("%d", dataID) +
		"&access_code=" + url.QueryEscape(accessCode)
}

// ContractListOfProjectURL 合同列表（front.php: contract/list_of_project）。
func ContractListOfProjectURL(base string, page, size int, contractNumber string) string {
	u := base + "/api/front/contract/list_of_project?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(contractNumber) != "" {
		u += "&contract_number=" + url.QueryEscape(contractNumber)
	}
	return u
}

// ProjectListByContractURL 项目列表（front.php: project/list）。
func ProjectListByContractURL(base string, contractID, page, size int) string {
	return base + "/api/front/project/list?contract_id=" + fmt.Sprintf("%d", contractID) +
		"&page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
}

// TaskListURL 任务列表（front.php: task/list）。
func TaskListURL(base string, page, size int, uuid string) string {
	u := base + "/api/front/task/list?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		u += "&uuid=" + url.QueryEscape(uuid)
	}
	return u
}

// TaskStatusByUUIDsURL task/status_by_uuids。
func TaskStatusByUUIDsURL(base, uuids string) string {
	return base + "/api/front/task/status_by_uuids?uuids=" + url.QueryEscape(uuids)
}

// TaskGetResultURL task/result。
func TaskGetResultURL(base string, taskID int) string {
	return base + "/api/front/task/result?id=" + fmt.Sprintf("%d", taskID)
}

// TaskDownloadResultURL task/download_result（POST）。
func TaskDownloadResultURL(base string) string {
	return base + "/api/front/task/download_result"
}

// TaskListOfWorkflowURL task/list_of_workflow。
func TaskListOfWorkflowURL(base string, page, size int, uuid string) string {
	u := base + "/api/front/task/list_of_workflow?page=" + fmt.Sprintf("%d", page) +
		"&size=" + fmt.Sprintf("%d", size)
	if strings.TrimSpace(uuid) != "" {
		u += "&uuid=" + url.QueryEscape(uuid)
	}
	return u
}

// TaskDetailOfModuleToolURL task/detail_of_module_tool。
func TaskDetailOfModuleToolURL(base string, id int) string {
	return base + "/api/front/task/detail_of_module_tool?id=" + fmt.Sprintf("%d", id)
}

// TaskPdfURLOfModuleToolURL task/pdf_url_of_module_tool，type=2 表示 source_type=task。
func TaskPdfURLOfModuleToolURL(base string, taskID int) string {
	return base + "/api/front/task/pdf_url_of_module_tool?type=2&source_id=" + fmt.Sprintf("%d", taskID)
}

// ProjectArticleListURL project_article/list。
func ProjectArticleListURL(base string) string {
	return base + "/api/front/project_article/list"
}
