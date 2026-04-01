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

func ProjectGetZipURL(base string, projectID int) string {
	return base + "/api/front/project/zip_url?id=" + fmt.Sprintf("%d", projectID)
}

func ProjectGetOriginalDataURL(base string, dataID int, accessCode string) string {
	return base + "/api/front/project/original_data_url?id=" + fmt.Sprintf("%d", dataID) +
		"&access_code=" + url.QueryEscape(accessCode)
}

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

func TaskStatusByUUIDsURL(base, uuids string) string {
	return base + "/api/front/task/status_by_uuids?uuids=" + url.QueryEscape(uuids)
}

func TaskGetResultURL(base string, taskID int) string {
	return base + "/api/front/task/result?id=" + fmt.Sprintf("%d", taskID)
}

func TaskDownloadResultURL(base string) string {
	return base + "/api/front/task/download_result"
}

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

func TaskDetailOfModuleToolURL(base string, id int) string {
	return base + "/api/front/task/detail_of_module_tool?id=" + fmt.Sprintf("%d", id)
}

func TaskPdfURLOfModuleToolURL(base string, taskID int) string {
	return base + "/api/front/task/pdf_url_of_module_tool?type=2&source_id=" + fmt.Sprintf("%d", taskID)
}

func ProjectArticleListURL(base string) string {
	return base + "/api/front/project_article/list"
}
