package Requests

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"cloud-platform-api/app/Config"
)

// AiPlatformCloudPublic 平台：公网云平台（默认）
const AiPlatformCloudPublic = "cloud_public"

// AiPlatformCloudIntranet 平台：内网云平台
const AiPlatformCloudIntranet = "cloud_intranet"

// AiPlatformImageCompare 平台：图片对比工具（占位，后续实现）
const AiPlatformImageCompare = "image_compare"

// AiQuestionTypeProject 提问类型：项目（与前端/产品约定枚举值 1 一致）。
// 以后可在此增加常量，例如 AiQuestionTypeTool = 2，并在 Controller 里 switch 扩展。
const AiQuestionTypeProject = 1

// AiQuestionTypeTask 提问类型：任务（工具任务/重分析/模块化等统一归到任务域）
const AiQuestionTypeTask = 2

// AiQuestionTypeProjectArticle 提问类型：项目文章
const AiQuestionTypeProjectArticle = 3

type AiPlatformCapability struct {
	QuestionTypes map[int]struct{}
}

var aiPlatformCapabilities = map[string]AiPlatformCapability{
	AiPlatformCloudPublic: {
		QuestionTypes: map[int]struct{}{
			AiQuestionTypeProject:        {},
			AiQuestionTypeTask:           {},
			AiQuestionTypeProjectArticle: {},
		},
	},
	AiPlatformCloudIntranet: {
		QuestionTypes: map[int]struct{}{
			AiQuestionTypeProject: {},
			AiQuestionTypeTask:    {},
		},
	},
	AiPlatformImageCompare: {
		QuestionTypes: map[int]struct{}{
			AiQuestionTypeProject: {},
		},
	},
}

func normalizePlatform(platform string) string {
	return strings.TrimSpace(platform)
}

func supportedPlatformsHint() string {
	platforms := make([]string, 0, len(aiPlatformCapabilities))
	for p := range aiPlatformCapabilities {
		platforms = append(platforms, p)
	}
	sort.Strings(platforms)
	return strings.Join(platforms, "/")
}

func supportedQuestionTypesHint(platform string) string {
	capability, ok := aiPlatformCapabilities[normalizePlatform(platform)]
	if !ok {
		return ""
	}
	values := make([]int, 0, len(capability.QuestionTypes))
	for v := range capability.QuestionTypes {
		values = append(values, v)
	}
	sort.Ints(values)
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, strconv.Itoa(v))
	}
	return strings.Join(parts, ",")
}

func IsSupportedPlatform(platform string) bool {
	platform = normalizePlatform(platform)
	cfg := Config.GetAiGatewayConfig()
	if cfg != nil {
		return cfg.IsPlatformSupported(platform)
	}
	_, ok := aiPlatformCapabilities[platform]
	return ok
}

func IsSupportedQuestionType(platform string, questionType int) bool {
	platform = normalizePlatform(platform)
	cfg := Config.GetAiGatewayConfig()
	if cfg != nil {
		return cfg.IsQuestionTypeAllowed(platform, questionType)
	}
	capability, ok := aiPlatformCapabilities[platform]
	if !ok {
		return false
	}
	_, ok = capability.QuestionTypes[questionType]
	return ok
}

// AiRobotChatRequest 对应接口 POST /api/v1/ai_robot/chat 的 JSON 体。
//
// 业务上建议按顺序理解字段：platform（平台）→ question_type（类型）→ question（自然语言，用于第三层解析具体接口）。
//
// binding 标签：Gin 用 go-playground/validator 做基础校验（如 required）；
// 更复杂的规则写在下面的 Validate() 方法里。
type AiRobotChatRequest struct {
	// Question 用户自然语言问题
	Question string `json:"question" binding:"required"`
	// Platform 平台标识：用于把同一套 AI 中转能力复用到不同后端（公网云平台/内网云平台/第三方工具等）。
	// 当前支持：cloud_public（默认）、cloud_intranet；image_compare 等后续扩展。
	Platform string `json:"platform" binding:"required"`
	// QuestionType 业务分类，当前仅实现 1=项目 2=任务 3=项目文章
	QuestionType int `json:"question_type" binding:"required"`
	// ConversationID 可选，会话标识；为空时后端会自动生成。
	ConversationID string `json:"conversation_id"`
	// MessageID 可选，消息标识；为空时后端会自动生成。
	MessageID string `json:"message_id"`
	// EnableContext 是否启用上下文增强；当传入 user_id 时后端会默认开启。
	EnableContext bool `json:"enable_context"`
	// UserID 当前对话所属用户标识，由接口显式传入；允许字符串用户ID。
	UserID string `json:"user_id"`
	// ResolvedQuestion 运行期生成的增强问题文本，不参与外部入参绑定。
	ResolvedQuestion string `json:"-"`
}

// AiRobotConversationManageRequest 用于读取/删除单条 ai_robot 会话。
type AiRobotConversationManageRequest struct {
	UserID         string `form:"user_id" json:"user_id"`
	Platform       string `form:"platform" json:"platform"`
	ConversationID string `form:"conversation_id" json:"conversation_id"`
}

// AiRobotConversationListRequest 用于分页查询 ai_robot 会话列表。
type AiRobotConversationListRequest struct {
	UserID         string `form:"user_id" json:"user_id"`
	Platform       string `form:"platform" json:"platform"`
	ConversationID string `form:"conversation_id" json:"conversation_id"`
	QuestionType   int    `form:"question_type" json:"question_type"`
	StartTime      string `form:"start_time" json:"start_time"`
	EndTime        string `form:"end_time" json:"end_time"`
	Page           int    `form:"page" json:"page"`
	Limit          int    `form:"limit" json:"limit"`
}

// AiRobotConversationCleanupRequest 用于按条件批量清理 ai_robot 会话。
type AiRobotConversationCleanupRequest struct {
	UserID       string `form:"user_id" json:"user_id"`
	Platform     string `form:"platform" json:"platform"`
	QuestionType int    `form:"question_type" json:"question_type"`
	StartTime    string `form:"start_time" json:"start_time"`
	EndTime      string `form:"end_time" json:"end_time"`
}

// Validate 补充校验（比 binding 更易读，错误信息可直接给前端展示）。
func (r *AiRobotChatRequest) Validate() error {
	q := strings.TrimSpace(r.Question)
	if q == "" {
		return errors.New("question 不能为空")
	}
	p := normalizePlatform(r.Platform)
	if p == "" {
		return errors.New("platform 不能为空")
	}
	if !IsSupportedPlatform(p) {
		return fmt.Errorf("platform 不支持（可选：%s）", supportedPlatformsHint())
	}
	if !IsSupportedQuestionType(p, r.QuestionType) {
		hint := supportedQuestionTypesHint(p)
		if hint != "" {
			return fmt.Errorf("%s 暂不支持 question_type=%d（可选：%s）", p, r.QuestionType, hint)
		}
		return fmt.Errorf("%s 暂不支持 question_type=%d", p, r.QuestionType)
	}
	r.UserID = strings.TrimSpace(r.UserID)
	r.ConversationID = strings.TrimSpace(r.ConversationID)
	if r.EnableContext {
		if r.UserID == "" {
			return errors.New("enable_context=true 时 user_id 不能为空")
		}
		if r.ConversationID == "" {
			return errors.New("enable_context=true 时 conversation_id 不能为空")
		}
	}

	return nil
}

func (r *AiRobotConversationManageRequest) Validate() error {
	r.UserID = strings.TrimSpace(r.UserID)
	if r.UserID == "" {
		return errors.New("user_id 不能为空")
	}
	p := normalizePlatform(r.Platform)
	if p == "" {
		return errors.New("platform 不能为空")
	}
	if !IsSupportedPlatform(p) {
		return fmt.Errorf("platform 不支持（可选：%s）", supportedPlatformsHint())
	}
	if strings.TrimSpace(r.ConversationID) == "" {
		return errors.New("conversation_id 不能为空")
	}
	return nil
}

func (r *AiRobotConversationListRequest) Normalize() {
	if r.Page <= 0 {
		r.Page = 1
	}
	if r.Limit <= 0 {
		r.Limit = 20
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	r.UserID = strings.TrimSpace(r.UserID)
	r.Platform = normalizePlatform(r.Platform)
	r.ConversationID = strings.TrimSpace(r.ConversationID)
	r.StartTime = strings.TrimSpace(r.StartTime)
	r.EndTime = strings.TrimSpace(r.EndTime)
}

func (r *AiRobotConversationListRequest) Validate() error {
	r.Normalize()
	if r.UserID == "" {
		return errors.New("user_id 不能为空")
	}
	if r.Platform != "" && !IsSupportedPlatform(r.Platform) {
		return fmt.Errorf("platform 不支持（可选：%s）", supportedPlatformsHint())
	}
	if r.QuestionType != 0 {
		if r.Platform == "" {
			return errors.New("指定 question_type 时必须同时传 platform")
		}
		if !IsSupportedQuestionType(r.Platform, r.QuestionType) {
			hint := supportedQuestionTypesHint(r.Platform)
			if hint != "" {
				return fmt.Errorf("%s 暂不支持 question_type=%d（可选：%s）", r.Platform, r.QuestionType, hint)
			}
			return fmt.Errorf("%s 暂不支持 question_type=%d", r.Platform, r.QuestionType)
		}
	}
	startAt, err := parseConversationListTime(r.StartTime)
	if err != nil {
		return fmt.Errorf("start_time 格式错误: %w", err)
	}
	endAt, err := parseConversationListTime(r.EndTime)
	if err != nil {
		return fmt.Errorf("end_time 格式错误: %w", err)
	}
	if !startAt.IsZero() && !endAt.IsZero() && startAt.After(endAt) {
		return errors.New("start_time 不能晚于 end_time")
	}
	return nil
}

func parseConversationListTime(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, nil
	}
	layouts := []string{
		time.RFC3339,
		"2006-01-02 15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("仅支持 RFC3339、2006-01-02 15:04:05、2006-01-02")
}

func ParseConversationListTimeForController(value string) (time.Time, error) {
	return parseConversationListTime(value)
}

func (r *AiRobotConversationCleanupRequest) Normalize() {
	r.UserID = strings.TrimSpace(r.UserID)
	r.Platform = normalizePlatform(r.Platform)
	r.StartTime = strings.TrimSpace(r.StartTime)
	r.EndTime = strings.TrimSpace(r.EndTime)
}

func (r *AiRobotConversationCleanupRequest) Validate() error {
	r.Normalize()
	if r.UserID == "" {
		return errors.New("user_id 不能为空")
	}
	if r.Platform != "" && !IsSupportedPlatform(r.Platform) {
		return fmt.Errorf("platform 不支持（可选：%s）", supportedPlatformsHint())
	}
	if r.QuestionType != 0 {
		if r.Platform == "" {
			return errors.New("指定 question_type 时必须同时传 platform")
		}
		if !IsSupportedQuestionType(r.Platform, r.QuestionType) {
			hint := supportedQuestionTypesHint(r.Platform)
			if hint != "" {
				return fmt.Errorf("%s 暂不支持 question_type=%d（可选：%s）", r.Platform, r.QuestionType, hint)
			}
			return fmt.Errorf("%s 暂不支持 question_type=%d", r.Platform, r.QuestionType)
		}
	}
	startAt, err := parseConversationListTime(r.StartTime)
	if err != nil {
		return fmt.Errorf("start_time 格式错误: %w", err)
	}
	endAt, err := parseConversationListTime(r.EndTime)
	if err != nil {
		return fmt.Errorf("end_time 格式错误: %w", err)
	}
	if !startAt.IsZero() && !endAt.IsZero() && startAt.After(endAt) {
		return errors.New("start_time 不能晚于 end_time")
	}
	if r.Platform == "" && r.QuestionType == 0 && startAt.IsZero() && endAt.IsZero() {
		return errors.New("批量删除至少需要一个筛选条件")
	}
	return nil
}
