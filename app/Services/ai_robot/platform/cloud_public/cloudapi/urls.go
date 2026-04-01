// Package cloudapi 拼装 Laravel 云平台 front 路由完整 URL（与 routes/front.php 对应）。
// 公网 cloud_public 与内网 cloud_intranet 共用同一套路径，仅站点根地址（base）不同；
// base 须为已去掉末尾斜杠的根 URL，与 Config.AiGatewayConfig.CloudAPIBaseByPlatform 返回值一致。
package cloudapi

import (
	"fmt"
	"net/url"
	"strings"
)

type ProjectGetUserAllProjectQuery struct {
	IsFilterTime int
	IsUsedFree   int
}

type ContractListOfProjectQuery struct {
	Page           int
	Size           int
	ContractNumber string
	Name           string
}

type ProjectListByContractQuery struct {
	ContractID     int
	Page           int
	Size           int
	Number         string
	Name           string
	WorkflowNameCN string
}

type TaskListQuery struct {
	Page           int
	Size           int
	ProjectNumber  string
	ProjectName    string
	UUID           string
	Name           string
	ToolName       string
	StatusValue    string
	WorkflowNameCN string
	CreatedAtStart string
	CreatedAtEnd   string
}

func appendQuery(basePath string, values url.Values) string {
	encoded := values.Encode()
	if encoded == "" {
		return basePath
	}
	return basePath + "?" + encoded
}

// ProjectGetUserAllProjectURL Laravel：project/get_user_all_project。
func ProjectGetUserAllProjectURL(base string, query ProjectGetUserAllProjectQuery) string {
	values := url.Values{}
	if query.IsFilterTime > 0 {
		values.Set("is_filter_time", fmt.Sprintf("%d", query.IsFilterTime))
	}
	if query.IsUsedFree > 0 {
		values.Set("is_used_free", fmt.Sprintf("%d", query.IsUsedFree))
	}
	return appendQuery(base+"/api/front/project/get_user_all_project", values)
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
func ContractListOfProjectURL(base string, query ContractListOfProjectQuery) string {
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", query.Page))
	values.Set("size", fmt.Sprintf("%d", query.Size))
	if strings.TrimSpace(query.ContractNumber) != "" {
		values.Set("contract_number", query.ContractNumber)
	}
	if strings.TrimSpace(query.Name) != "" {
		values.Set("name", query.Name)
	}
	return appendQuery(base+"/api/front/contract/list_of_project", values)
}

// ProjectListByContractURL 项目列表（front.php: project/list）。
func ProjectListByContractURL(base string, query ProjectListByContractQuery) string {
	values := url.Values{}
	values.Set("contract_id", fmt.Sprintf("%d", query.ContractID))
	values.Set("page", fmt.Sprintf("%d", query.Page))
	values.Set("size", fmt.Sprintf("%d", query.Size))
	if strings.TrimSpace(query.Number) != "" {
		values.Set("number", query.Number)
	}
	if strings.TrimSpace(query.Name) != "" {
		values.Set("name", query.Name)
	}
	if strings.TrimSpace(query.WorkflowNameCN) != "" {
		values.Set("workflow_name_cn", query.WorkflowNameCN)
	}
	return appendQuery(base+"/api/front/project/list", values)
}

// TaskListURL 任务列表（front.php: task/list）。
func TaskListURL(base string, query TaskListQuery) string {
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", query.Page))
	values.Set("size", fmt.Sprintf("%d", query.Size))
	if strings.TrimSpace(query.ProjectNumber) != "" {
		values.Set("project_number", query.ProjectNumber)
	}
	if strings.TrimSpace(query.ProjectName) != "" {
		values.Set("project_name", query.ProjectName)
	}
	if strings.TrimSpace(query.UUID) != "" {
		values.Set("uuid", query.UUID)
	}
	if strings.TrimSpace(query.Name) != "" {
		values.Set("name", query.Name)
	}
	if strings.TrimSpace(query.ToolName) != "" {
		values.Set("tool_name", query.ToolName)
	}
	if strings.TrimSpace(query.StatusValue) != "" {
		values.Set("status_value", query.StatusValue)
	}
	if strings.TrimSpace(query.WorkflowNameCN) != "" {
		values.Set("workflow_name_cn", query.WorkflowNameCN)
	}
	if strings.TrimSpace(query.CreatedAtStart) != "" {
		values.Set("created_at_start", query.CreatedAtStart)
	}
	if strings.TrimSpace(query.CreatedAtEnd) != "" {
		values.Set("created_at_end", query.CreatedAtEnd)
	}
	return appendQuery(base+"/api/front/task/list", values)
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
func TaskListOfWorkflowURL(base string, query TaskListQuery) string {
	values := url.Values{}
	values.Set("page", fmt.Sprintf("%d", query.Page))
	values.Set("size", fmt.Sprintf("%d", query.Size))
	if strings.TrimSpace(query.ProjectNumber) != "" {
		values.Set("project_number", query.ProjectNumber)
	}
	if strings.TrimSpace(query.ProjectName) != "" {
		values.Set("project_name", query.ProjectName)
	}
	if strings.TrimSpace(query.UUID) != "" {
		values.Set("uuid", query.UUID)
	}
	if strings.TrimSpace(query.Name) != "" {
		values.Set("name", query.Name)
	}
	if strings.TrimSpace(query.StatusValue) != "" {
		values.Set("status_value", query.StatusValue)
	}
	if strings.TrimSpace(query.WorkflowNameCN) != "" {
		values.Set("workflow_name_cn", query.WorkflowNameCN)
	}
	if strings.TrimSpace(query.CreatedAtStart) != "" {
		values.Set("created_at_start", query.CreatedAtStart)
	}
	if strings.TrimSpace(query.CreatedAtEnd) != "" {
		values.Set("created_at_end", query.CreatedAtEnd)
	}
	return appendQuery(base+"/api/front/task/list_of_workflow", values)
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
