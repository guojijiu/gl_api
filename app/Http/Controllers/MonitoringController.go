package Controllers

import (
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
	monitoringService *Services.MonitoringService
}

// NewMonitoringController 创建监控告警控制器
func NewMonitoringController() *MonitoringController {
	return &MonitoringController{}
}

// SetMonitoringService 设置监控告警服务
func (c *MonitoringController) SetMonitoringService(service *Services.MonitoringService) {
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
	name := ctx.Query("name")
	limitStr := ctx.DefaultQuery("limit", "100")
	startTimeStr := ctx.Query("start_time")
	endTimeStr := ctx.Query("end_time")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的limit参数")
		return
	}

	// 获取监控指标
	metrics, err := c.monitoringService.GetMetrics(metricType, name, limit)
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "获取监控指标失败: "+err.Error())
		return
	}

	// 时间范围过滤
	if startTimeStr != "" || endTimeStr != "" {
		filteredMetrics := make([]Models.MonitoringMetric, 0)
		for _, metric := range metrics {
			if startTimeStr != "" {
				if startTime, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
					if metric.Timestamp.Before(startTime) {
						continue
					}
				}
			}
			if endTimeStr != "" {
				if endTime, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
					if metric.Timestamp.After(endTime) {
						continue
					}
				}
			}
			filteredMetrics = append(filteredMetrics, metric)
		}
		metrics = filteredMetrics
	}

	c.Success(ctx, gin.H{
		"metrics": metrics,
		"total":   len(metrics),
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
		alerts = []Models.Alert{}
	} else if end > total {
		alerts = alerts[start:total]
	} else {
		alerts = alerts[start:end]
	}

	c.Success(ctx, gin.H{
		"alerts":     alerts,
		"total":      total,
		"page":       page,
		"page_size":  pageSize,
		"total_pages": (total + pageSize - 1) / pageSize,
		"status":     status,
		"severity":   severity,
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
	err = c.monitoringService.AcknowledgeAlert(uint(alertID), userID)
	if err != nil {
		c.Error(ctx, http.StatusInternalServerError, "确认告警失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message": "告警确认成功",
		"alert_id": alertID,
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
		"message": "告警解决成功",
		"alert_id": alertID,
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

	// 构建查询条件
	query := c.monitoringService.GetDB().Model(&Models.AlertRule{})
	
	if enabledStr != "" {
		if enabled, err := strconv.ParseBool(enabledStr); err == nil {
			query = query.Where("enabled = ?", enabled)
		}
	}
	
	if ruleType != "" {
		query = query.Where("type = ?", ruleType)
	}

	// 执行查询
	var rules []Models.AlertRule
	if err := query.Find(&rules).Error; err != nil {
		c.Error(ctx, http.StatusInternalServerError, "获取告警规则失败: "+err.Error())
		return
	}

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
		Name:        req.Name,
		Description: req.Description,
		Type:        req.Type,
		MetricType:  req.MetricType,
		MetricName:  req.MetricName,
		Condition:   req.Condition,
		Threshold:   req.Threshold,
		Duration:    req.Duration,
		Severity:    req.Severity,
		Enabled:     req.Enabled,
		Suppression: req.Suppression,
		SuppressionWindow: req.SuppressionWindow,
		Escalation:  req.Escalation,
		EscalationDelay: req.EscalationDelay,
		MaxEscalationLevel: req.MaxEscalationLevel,
		NotificationChannels: req.NotificationChannels,
		Tags:        req.Tags,
		CreatedBy:   userID,
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
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警规则ID")
		return
	}

	var req UpdateAlertRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "参数绑定失败: "+err.Error())
		return
	}

	// 查找现有规则
	var rule Models.AlertRule
	if err := c.monitoringService.GetDB().First(&rule, ruleID).Error; err != nil {
		c.Error(ctx, http.StatusNotFound, "告警规则不存在")
		return
	}

	// 更新规则字段
	if req.Name != "" {
		rule.Name = req.Name
	}
	if req.Description != "" {
		rule.Description = req.Description
	}
	if req.Type != "" {
		rule.Type = req.Type
	}
	if req.MetricType != "" {
		rule.MetricType = req.MetricType
	}
	if req.MetricName != "" {
		rule.MetricName = req.MetricName
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Threshold != 0 {
		rule.Threshold = req.Threshold
	}
	if req.Duration != 0 {
		rule.Duration = req.Duration
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	rule.Enabled = req.Enabled
	rule.Suppression = req.Suppression
	if req.SuppressionWindow != 0 {
		rule.SuppressionWindow = req.SuppressionWindow
	}
	rule.Escalation = req.Escalation
	if req.EscalationDelay != 0 {
		rule.EscalationDelay = req.EscalationDelay
	}
	if req.MaxEscalationLevel != 0 {
		rule.MaxEscalationLevel = req.MaxEscalationLevel
	}
	if req.NotificationChannels != "" {
		rule.NotificationChannels = req.NotificationChannels
	}
	if req.Tags != "" {
		rule.Tags = req.Tags
	}

	if err := c.monitoringService.UpdateAlertRule(&rule); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "更新告警规则失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message": "告警规则更新成功",
		"rule_id": rule.ID,
		"rule":    rule,
	}, "告警规则更新成功")
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
	ruleID, err := strconv.ParseUint(ruleIDStr, 10, 32)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的告警规则ID")
		return
	}

	if err := c.monitoringService.DeleteAlertRule(uint(ruleID)); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "删除告警规则失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message": "告警规则删除成功",
		"rule_id": ruleID,
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

	// 构建查询条件
	query := c.monitoringService.GetDB().Model(&Models.NotificationRecord{})
	
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// 执行查询
	var notifications []Models.NotificationRecord
	if err := query.Order("created_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		c.Error(ctx, http.StatusInternalServerError, "获取通知记录失败: "+err.Error())
		return
	}

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

	c.monitoringService.GetDB().Model(&Models.MonitoringMetric{}).Count(&totalMetrics)
	c.monitoringService.GetDB().Model(&Models.Alert{}).Count(&totalAlerts)
	c.monitoringService.GetDB().Model(&Models.AlertRule{}).Count(&totalRules)
	c.monitoringService.GetDB().Model(&Models.NotificationRecord{}).Count(&totalNotifications)
	c.monitoringService.GetDB().Model(&Models.Alert{}).Where("status = ?", "active").Count(&activeAlerts)
	c.monitoringService.GetDB().Model(&Models.Alert{}).Where("status = ? AND severity = ?", "active", "critical").Count(&criticalAlerts)

	// 获取最近24小时的指标数量
	var recentMetrics int64
	c.monitoringService.GetDB().Model(&Models.MonitoringMetric{}).
		Where("timestamp > ?", time.Now().Add(-24*time.Hour)).
		Count(&recentMetrics)

	// 获取最近24小时的告警数量
	var recentAlerts int64
	c.monitoringService.GetDB().Model(&Models.Alert{}).
		Where("fired_at > ?", time.Now().Add(-24*time.Hour)).
		Count(&recentAlerts)

	stats := gin.H{
		"total_metrics":      totalMetrics,
		"total_alerts":       totalAlerts,
		"total_rules":        totalRules,
		"total_notifications": totalNotifications,
		"active_alerts":      activeAlerts,
		"critical_alerts":    criticalAlerts,
		"recent_metrics_24h": recentMetrics,
		"recent_alerts_24h":  recentAlerts,
		"timestamp":          time.Now(),
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
