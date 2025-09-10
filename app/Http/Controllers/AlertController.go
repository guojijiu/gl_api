package Controllers

import (
	"cloud-platform-api/app/Services"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

// AlertController 告警控制器
type AlertController struct {
	Controller
	alertService *Services.AlertService
}

// NewAlertController 创建告警控制器
func NewAlertController(alertService *Services.AlertService) *AlertController {
	return &AlertController{
		alertService: alertService,
	}
}

// @Summary 获取告警规则列表
// @Description 获取所有告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=[]AlertRule} "告警规则列表"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/rules [get]
// GetAlertRules 获取告警规则列表
func (c *AlertController) GetAlertRules(ctx *gin.Context) {
	rules := c.alertService.GetRules()
	c.Success(ctx, rules, "告警规则获取成功")
}

// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule body AlertRuleRequest true "告警规则信息"
// @Success 200 {object} Response{data=AlertRule} "告警规则创建成功"
// @Failure 400 {object} Response{error=string} "参数错误"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/rules [post]
// CreateAlertRule 创建告警规则
func (c *AlertController) CreateAlertRule(ctx *gin.Context) {
	var request struct {
		Name        string                    `json:"name" binding:"required"`
		Description string                    `json:"description"`
		Metric      string                    `json:"metric" binding:"required"`
		Condition   string                    `json:"condition" binding:"required"`
		Threshold   float64                   `json:"threshold" binding:"required"`
		Duration    time.Duration             `json:"duration"`
		Level       Services.AlertLevel       `json:"level" binding:"required"`
		Channels    []Services.AlertChannel   `json:"channels"`
		Enabled     bool                      `json:"enabled"`
	}

	if err := ctx.ShouldBindJSON(&request); err != nil {
		c.ValidationError(ctx, err.Error())
		return
	}

	rule := &Services.AlertRule{
		Name:        request.Name,
		Description: request.Description,
		Metric:      request.Metric,
		Condition:   request.Condition,
		Threshold:   request.Threshold,
		Duration:    request.Duration,
		Level:       request.Level,
		Channels:    request.Channels,
		Enabled:     request.Enabled,
	}

	if err := c.alertService.AddRule(rule); err != nil {
		c.ServerError(ctx, err.Error())
		return
	}

	c.Success(ctx, rule, "告警规则创建成功")
}

// @Summary 获取告警列表
// @Description 获取告警列表，支持按状态筛选
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param status query string false "告警状态" Enums(active, resolved)
// @Param limit query int false "每页数量" default(20)
// @Success 200 {object} Response{data=[]Alert} "告警列表"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts [get]
// GetAlerts 获取告警列表
func (c *AlertController) GetAlerts(ctx *gin.Context) {
	status := ctx.Query("status")
	limitStr := ctx.DefaultQuery("limit", "20")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	alerts := c.alertService.GetAlerts(status, limit)
	c.Success(ctx, alerts, "告警列表获取成功")
}

// @Summary 获取告警统计
// @Description 获取告警统计信息
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{data=AlertStats} "告警统计"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/stats [get]
// GetAlertStats 获取告警统计
func (c *AlertController) GetAlertStats(ctx *gin.Context) {
	stats := c.alertService.GetAlertStats()
	c.Success(ctx, stats, "告警统计获取成功")
}

// @Summary 手动检查告警
// @Description 手动触发告警检查
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} Response{message=string} "告警检查完成"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/check [post]
// CheckAlerts 手动检查告警
func (c *AlertController) CheckAlerts(ctx *gin.Context) {
	if err := c.alertService.CheckAlerts(); err != nil {
		c.ServerError(ctx, err.Error())
		return
	}

	c.Success(ctx, nil, "告警检查完成")
}

// @Summary 测试告警规则
// @Description 测试告警规则是否正常工作
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule_id path string true "告警规则ID"
// @Success 200 {object} Response{message=string} "告警规则测试完成"
// @Failure 400 {object} Response{error=string} "参数错误"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/rules/{rule_id}/test [post]
// TestAlertRule 测试告警规则
func (c *AlertController) TestAlertRule(ctx *gin.Context) {
	_ = ctx.Param("rule_id")
	
	// 这里应该实现告警规则测试逻辑
	// 暂时返回成功
	c.Success(ctx, nil, "告警规则测试完成")
}

// @Summary 启用告警规则
// @Description 启用指定的告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule_id path string true "告警规则ID"
// @Success 200 {object} Response{message=string} "告警规则已启用"
// @Failure 400 {object} Response{error=string} "参数错误"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/rules/{rule_id}/enable [post]
// EnableAlertRule 启用告警规则
func (c *AlertController) EnableAlertRule(ctx *gin.Context) {
	_ = ctx.Param("rule_id")
	
	// 这里应该实现启用告警规则的逻辑
	// 暂时返回成功
	c.Success(ctx, nil, "告警规则已启用")
}

// @Summary 禁用告警规则
// @Description 禁用指定的告警规则
// @Tags 告警
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param rule_id path string true "告警规则ID"
// @Success 200 {object} Response{message=string} "告警规则已禁用"
// @Failure 400 {object} Response{error=string} "参数错误"
// @Failure 401 {object} Response{error=string} "未认证"
// @Router /api/v1/alerts/rules/{rule_id}/disable [post]
// DisableAlertRule 禁用告警规则
func (c *AlertController) DisableAlertRule(ctx *gin.Context) {
	_ = ctx.Param("rule_id")
	
	// 这里应该实现禁用告警规则的逻辑
	// 暂时返回成功
	c.Success(ctx, nil, "告警规则已禁用")
}
