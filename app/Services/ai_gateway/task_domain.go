package aigateway

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func (s *Service) FetchTaskListByUUID(ctx context.Context, platform string, userToken string, uuid string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskListURL(platform, 1, 1000, uuid), userToken)
}

func (s *Service) FetchTaskStatusByUUIDs(ctx context.Context, platform string, userToken string, uuidsCSV string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskStatusByUUIDsURL(platform, uuidsCSV), userToken)
}

func (s *Service) FetchTaskResult(ctx context.Context, platform string, userToken string, taskID int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskGetResultURL(platform, taskID), userToken)
}

func (s *Service) DownloadTaskResult(ctx context.Context, platform string, userToken string, taskID int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformPOSTJSON(ctx, s.cfg.TaskDownloadResultURL(platform), userToken, map[string]interface{}{"id": taskID})
}

func (s *Service) FetchWorkflowTaskListByUUID(ctx context.Context, platform string, userToken string, uuid string) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskListOfWorkflowURL(platform, 1, 1000, uuid), userToken)
}

func (s *Service) FetchModuleTaskDetail(ctx context.Context, platform string, userToken string, id int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskDetailOfModuleToolURL(platform, id), userToken)
}

func (s *Service) FetchModuleTaskResultURL(ctx context.Context, platform string, userToken string, id int) ([]byte, int, error) {
	base := strings.TrimSpace(s.cfg.CloudAPIBaseByPlatform(platform))
	if base == "" {
		if platform == "cloud_intranet" {
			return nil, 0, errors.New("未配置 CLOUD_PLATFORM_INTRANET_BASE_URL，无法请求内网云平台接口")
		}
		return nil, 0, errors.New("未配置 CLOUD_PLATFORM_BASE_URL，无法请求云平台接口")
	}
	return s.CallCloudPlatformGET(ctx, s.cfg.TaskPdfURLOfModuleToolURL(platform, id), userToken)
}

func (s *Service) ExtractTaskUUIDsFromQuestion(question string) []string {
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

func (s *Service) ExtractNumericIDsFromQuestion(question string) []int {
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
