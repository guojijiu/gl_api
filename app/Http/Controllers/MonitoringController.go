package Controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"

	"github.com/gin-gonic/gin"
)

// MonitoringController 监控告警控制器
type MonitoringController struct {
	Controller
	monitoringService *Services.OptimizedMonitoringService
}

// NewMonitoringController 创建监控告警控制器
func NewMonitoringController() *MonitoringController {
	return &MonitoringController{}
}

// SetMonitoringService 设置监控告警服务
func (c *MonitoringController) SetMonitoringService(service *Services.OptimizedMonitoringService) {
	c.monitoringService = service
}

// GetMetrics 获取监控指标
// @Summary 获取监控指标
// @Description 获取系统监控指标数据
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param type query string false "指标类型" Enums(system,application,database,cache,business)
// @Param name query string false "指标名称"
// @Param limit query int false "限制返回数量" default(100)
// @Param start_time query string false "开始时间" format(date-time)
// @Param end_time query string false "结束时间" format(date-time)
// @Success 200 {object} Response "监控指标列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/metrics [get]
func (c *MonitoringController) GetMetrics(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取查询参数
	metricType := ctx.Query("type")
	// Prometheus/健康探测一般不会带 query 参数，这里给一个默认值，避免 metricType 为空导致 500
	if metricType == "" {
		metricType = "system"
	}
	name := ctx.Query("name")
	limitStr := ctx.DefaultQuery("limit", "100")
	// startTimeStr := ctx.Query("start_time")
	// endTimeStr := ctx.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的limit参数")
		return
	}

	// 获取监控指标
	metrics, err := c.monitoringService.GetMetrics(metricType)
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "获取监控指标失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"metrics": metrics,
		"total":   0,
		"type":    metricType,
		"name":    name,
		"limit":   limit,
	}, "获取监控指标成功")
}

// GetAlerts 获取告警记录
// @Summary 获取告警记录
// @Description 获取系统告警记录列表
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param status query string false "告警状态" Enums(active,acknowledged,resolved,suppressed)
// @Param severity query string false "严重程度" Enums(info,warning,critical,emergency)
// @Param limit query int false "限制返回数量" default(50)
// @Param page query int false "页码" default(1)
// @Param page_size query int false "每页数量" default(20)
// @Success 200 {object} Response "告警记录列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alerts [get]
func (c *MonitoringController) GetAlerts(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取查询参数
	status := ctx.Query("status")
	severity := ctx.Query("severity")
	limitStr := ctx.DefaultQuery("limit", "50")
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "20")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的limit参数")
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的page参数")
		return
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的page_size参数")
		return
	}

	// 获取告警记录
	alerts, err := c.monitoringService.GetAlerts(status, severity, limit)
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "获取告警记录失败: "+err.Error())
		return
	}

	// 分页处理
	total := len(alerts)
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= total {
		alerts = []interface{}{}
	} else if end > total {
		alerts = alerts[start:total]
	} else {
		alerts = alerts[start:end]
	}

	c.Success(ctx, gin.H{
		"alerts":      alerts,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
		"status":      status,
		"severity":    severity,
	}, "获取告警记录成功")
}

// AcknowledgeAlert 确认告警
// @Summary 确认告警
// @Description 确认指定的告警
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} Response "确认成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alerts/{id}/acknowledge [post]
func (c *MonitoringController) AcknowledgeAlert(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取告警ID
	alertIDStr := ctx.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警ID")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		c.Error(ctx, http.StatusUnauthorized, "用户未认证")
		return
	}

	userID := currentUserID.(uint)

	// 确认告警
	err = c.monitoringService.AcknowledgeAlert(uint(alertID), fmt.Sprintf("%d", userID))
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "确认告警失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message":         "告警确认成功",
		"alert_id":        alertID,
		"acknowledged_by": userID,
		"acknowledged_at": time.Now(),
	}, "告警确认成功")
}

// ResolveAlert 解决告警
// @Summary 解决告警
// @Description 解决指定的告警
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param id path int true "告警ID"
// @Success 200 {object} Response "解决成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alerts/{id}/resolve [post]
func (c *MonitoringController) ResolveAlert(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取告警ID
	alertIDStr := ctx.Param("id")
	alertID, err := strconv.ParseUint(alertIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警ID")
		return
	}

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		c.Error(ctx, http.StatusUnauthorized, "用户未认证")
		return
	}

	userID := currentUserID.(uint)

	// 解决告警
	err = c.monitoringService.ResolveAlert(uint(alertID), userID)
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "解决告警失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message":     "告警解决成功",
		"alert_id":    alertID,
		"resolved_by": userID,
		"resolved_at": time.Now(),
	}, "告警解决成功")
}

// GetAlertRules 获取告警规则
// @Summary 获取告警规则
// @Description 获取系统告警规则列表
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param enabled query bool false "是否启用"
// @Param type query string false "规则类型" Enums(threshold,trend,anomaly)
// @Success 200 {object} Response "告警规则列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alert-rules [get]
func (c *MonitoringController) GetAlertRules(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取查询参数
	enabledStr := ctx.Query("enabled")
	ruleType := ctx.Query("type")

	// 构建查询条件 - 暂时返回空结果
	var rules []interface{}

	c.Success(ctx, gin.H{
		"alert_rules": rules,
		"total":       len(rules),
		"enabled":     enabledStr,
		"type":        ruleType,
	}, "获取告警规则成功")
}

// CreateAlertRule 创建告警规则
// @Summary 创建告警规则
// @Description 创建新的告警规则
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param rule body CreateAlertRuleRequest true "告警规则信息"
// @Success 200 {object} Response "创建成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alert-rules [post]
func (c *MonitoringController) CreateAlertRule(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	var req CreateAlertRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "参数绑定失败: "+err.Error())
		return
	}

	// 获取当前用户ID
	currentUserID, exists := ctx.Get("user_id")
	if !exists {
		c.Error(ctx, http.StatusUnauthorized, "用户未认证")
		return
	}

	userID := currentUserID.(uint)

	// 创建告警规则
	rule := &Models.AlertRule{
		Name:                 req.Name,
		Description:          req.Description,
		Type:                 req.Type,
		MetricType:           req.MetricType,
		MetricName:           req.MetricName,
		Condition:            req.Condition,
		Threshold:            req.Threshold,
		Duration:             req.Duration,
		Severity:             req.Severity,
		Enabled:              req.Enabled,
		Suppression:          req.Suppression,
		SuppressionWindow:    req.SuppressionWindow,
		Escalation:           req.Escalation,
		EscalationDelay:      req.EscalationDelay,
		MaxEscalationLevel:   req.MaxEscalationLevel,
		NotificationChannels: req.NotificationChannels,
		Tags:                 req.Tags,
		CreatedBy:            userID,
	}

	if err := c.monitoringService.CreateAlertRule(rule); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "创建告警规则失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message": "告警规则创建成功",
		"rule_id": rule.ID,
		"rule":    rule,
	}, "告警规则创建成功")
}

// UpdateAlertRule 更新告警规则
// @Summary 更新告警规则
// @Description 更新指定的告警规则
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param id path int true "告警规则ID"
// @Param rule body UpdateAlertRuleRequest true "告警规则信息"
// @Success 200 {object} Response "更新成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alert-rules/{id} [put]
func (c *MonitoringController) UpdateAlertRule(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取告警规则ID
	ruleIDStr := ctx.Param("id")
	_, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警规则ID")
		return
	}

	var req UpdateAlertRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "参数绑定失败: "+err.Error())
		return
	}

	// 暂时跳过数据库操作
	c.Success(ctx, gin.H{
		"message": "告警规则更新功能暂时不可用",
	}, "告警规则更新成功")
	return
}

// DeleteAlertRule 删除告警规则
// @Summary 删除告警规则
// @Description 删除指定的告警规则
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param id path int true "告警规则ID"
// @Success 200 {object} Response "删除成功"
// @Failure 400 {object} Response "参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/alert-rules/{id} [delete]
func (c *MonitoringController) DeleteAlertRule(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取告警规则ID
	ruleIDStr := ctx.Param("id")
	_, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警规则ID")
		return
	}

	// 暂时跳过数据库操作
	c.Success(ctx, gin.H{
		"message": "告警规则删除功能暂时不可用",
	}, "告警规则删除成功")
}

// GetSystemHealth 获取系统健康状态
// @Summary 获取系统健康状态
// @Description 获取系统整体健康状态信息
// @Tags 监控告警
// @Accept json
// @Produce json
// @Success 200 {object} Response "系统健康状态"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/health [get]
func (c *MonitoringController) GetSystemHealth(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	health := c.monitoringService.GetSystemHealth()

	c.Success(ctx, gin.H{
		"health": health,
	}, "获取系统健康状态成功")
}

// GetNotificationRecords 获取通知记录
// @Summary 获取通知记录
// @Description 获取系统通知发送记录
// @Tags 监控告警
// @Accept json
// @Produce json
// @Param channel query string false "通知渠道" Enums(email,webhook,slack,dingtalk,sms)
// @Param status query string false "发送状态" Enums(pending,sent,failed,retrying)
// @Param limit query int false "限制返回数量" default(50)
// @Success 200 {object} Response "通知记录列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/notifications [get]
func (c *MonitoringController) GetNotificationRecords(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取查询参数
	channel := ctx.Query("channel")
	status := ctx.Query("status")
	limitStr := ctx.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的limit参数")
		return
	}

	// 暂时跳过数据库查询
	var notifications []interface{}

	c.Success(ctx, gin.H{
		"notifications": notifications,
		"total":         len(notifications),
		"channel":       channel,
		"status":        status,
		"limit":         limit,
	}, "获取通知记录成功")
}

// GetMonitoringStats 获取监控统计信息
// @Summary 获取监控统计信息
// @Description 获取监控系统统计信息
// @Tags 监控告警
// @Accept json
// @Produce json
// @Success 200 {object} Response "监控统计信息"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/monitoring/stats [get]
func (c *MonitoringController) GetMonitoringStats(ctx *gin.Context) {
	if c.monitoringService == nil {
		c.Error(ctx, http.StatusInternalServerError, "监控服务未初始化")
		return
	}

	// 获取各种统计信息
	var totalMetrics, totalAlerts, totalRules, totalNotifications int64
	var activeAlerts, criticalAlerts int64

	// 暂时使用默认值
	totalMetrics = 0
	totalAlerts = 0
	totalRules = 0
	totalNotifications = 0
	activeAlerts = 0
	criticalAlerts = 0

	// 暂时使用默认值
	var recentMetrics int64 = 0
	var recentAlerts int64 = 0

	stats := gin.H{
		"total_metrics":       totalMetrics,
		"total_alerts":        totalAlerts,
		"total_rules":         totalRules,
		"total_notifications": totalNotifications,
		"active_alerts":       activeAlerts,
		"critical_alerts":     criticalAlerts,
		"recent_metrics_24h":  recentMetrics,
		"recent_alerts_24h":   recentAlerts,
		"timestamp":           time.Now(),
	}

	c.Success(ctx, gin.H{
		"stats": stats,
	}, "获取监控统计信息成功")
}

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	Name                 string  `json:"name" binding:"required"`
	Description          string  `json:"description"`
	Type                 string  `json:"type" binding:"required"`
	MetricType           string  `json:"metric_type" binding:"required"`
	MetricName           string  `json:"metric_name" binding:"required"`
	Condition            string  `json:"condition" binding:"required"`
	Threshold            float64 `json:"threshold" binding:"required"`
	Duration             int     `json:"duration"`
	Severity             string  `json:"severity" binding:"required"`
	Enabled              bool    `json:"enabled"`
	Suppression          bool    `json:"suppression"`
	SuppressionWindow    int     `json:"suppression_window"`
	Escalation           bool    `json:"escalation"`
	EscalationDelay      int     `json:"escalation_delay"`
	MaxEscalationLevel   int     `json:"max_escalation_level"`
	NotificationChannels string  `json:"notification_channels"`
	Tags                 string  `json:"tags"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	Name                 string  `json:"name"`
	Description          string  `json:"description"`
	Type                 string  `json:"type"`
	MetricType           string  `json:"metric_type"`
	MetricName           string  `json:"metric_name"`
	Condition            string  `json:"condition"`
	Threshold            float64 `json:"threshold"`
	Duration             int     `json:"duration"`
	Severity             string  `json:"severity"`
	Enabled              bool    `json:"enabled"`
	Suppression          bool    `json:"suppression"`
	SuppressionWindow    int     `json:"suppression_window"`
	Escalation           bool    `json:"escalation"`
	EscalationDelay      int     `json:"escalation_delay"`
	MaxEscalationLevel   int     `json:"max_escalation_level"`
	NotificationChannels string  `json:"notification_channels"`
	Tags                 string  `json:"tags"`
}
