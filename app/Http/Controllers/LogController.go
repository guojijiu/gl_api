package Controllers

import (
	"cloud-platform-api/app/Config"
	"cloud-platform-api/app/Services"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// LogController 日志控制器
//
// 重要功能说明：
// 1. 日志统计查询：获取各类型日志的统计信息
// 2. 日志监控管理：管理监控规则、查看告警状态
// 3. 日志配置管理：查看和更新日志配置
// 4. 日志搜索查询：支持多条件日志搜索
// 5. 日志报告生成：生成日志分析报告
// 6. 系统健康检查：检查日志系统状态
//
// 安全特性：
// - 所有接口都需要管理员权限
// - 支持操作审计日志
// - 防止敏感信息泄露
// - 支持操作频率限制
type LogController struct {
	Controller
	logManager *Services.LogManagerService
	logMonitor *Services.LogMonitorService
	config     *Config.LogConfig
}

// NewLogController 创建日志控制器
func NewLogController() *LogController {
	// 创建默认日志配置
	config := &Config.LogConfig{}
	config.SetDefaults()
	
	logManager := Services.NewLogManagerService(config)
	logMonitor := Services.NewLogMonitorService(logManager, config)
	
	return &LogController{
		logManager: logManager,
		logMonitor: logMonitor,
		config:     config,
	}
}

// GetLogStats 获取日志统计信息
// @Summary 获取日志统计信息
// @Description 获取各类型日志的统计信息，包括数量、级别分布等
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param logger query string false "日志记录器名称"
// @Param level query string false "日志级别"
// @Param time_range query string false "时间范围(1h, 24h, 7d, 30d)"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/stats [get]
func (c *LogController) GetLogStats(ctx *gin.Context) {
	loggerName := ctx.Query("logger")
	level := ctx.Query("level")
	timeRange := ctx.Query("time_range")
	
	// 获取基础统计信息
	stats := c.logManager.GetStats()
	
	// 根据查询条件过滤
	var filteredStats *Services.LogStats
	if loggerName != "" || level != "" || timeRange != "" {
		filteredStats = c.filterStats(stats, loggerName, level, timeRange)
	} else {
		filteredStats = stats
	}
	
	// 获取各日志记录器的详细统计
	loggerStats := make(map[string]*Services.LoggerStats)
	if loggerName != "" {
		if stats := c.logManager.GetLoggerStats(loggerName); stats != nil {
			loggerStats[loggerName] = stats
		}
	} else {
		// 获取所有日志记录器的统计
		loggers := []string{"request", "sql", "error", "audit", "security", "business", "access", "system"}
		for _, name := range loggers {
			if stats := c.logManager.GetLoggerStats(name); stats != nil {
				loggerStats[name] = stats
			}
		}
	}
	
	c.Success(ctx, gin.H{
		"overview":     filteredStats,
		"logger_stats": loggerStats,
		"timestamp":    time.Now(),
	}, "获取日志统计信息成功")
}

// GetLogMonitorStats 获取日志监控统计
// @Summary 获取日志监控统计
// @Description 获取日志监控系统的统计信息，包括规则、告警等
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/stats [get]
func (c *LogController) GetLogMonitorStats(ctx *gin.Context) {
	stats := c.logMonitor.GetStats()
	
	c.Success(ctx, gin.H{
		"monitor_stats": stats,
		"timestamp":     time.Now(),
	}, "获取日志监控统计成功")
}

// GetLogRules 获取监控规则列表
// @Summary 获取监控规则列表
// @Description 获取所有日志监控规则
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param enabled query bool false "是否启用"
// @Param logger query string false "日志记录器名称"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/rules [get]
func (c *LogController) GetLogRules(ctx *gin.Context) {
	enabledStr := ctx.Query("enabled")
	logger := ctx.Query("logger")
	
	rules := c.logMonitor.GetRules()
	
	// 过滤规则
	var filteredRules []*Services.LogRule
	for _, rule := range rules {
		// 按启用状态过滤
		if enabledStr != "" {
			enabled, err := strconv.ParseBool(enabledStr)
			if err == nil && rule.Enabled != enabled {
				continue
			}
		}
		
		// 按日志记录器过滤
		if logger != "" && rule.Logger != logger && rule.Logger != "*" {
			continue
		}
		
		filteredRules = append(filteredRules, rule)
	}
	
	c.Success(ctx, gin.H{
		"rules":     filteredRules,
		"total":     len(filteredRules),
		"timestamp": time.Now(),
	}, "获取监控规则列表成功")
}

// GetLogRule 获取单个监控规则
// @Summary 获取单个监控规则
// @Description 根据ID获取指定的监控规则
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rule_id path string true "规则ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/rules/{rule_id} [get]
func (c *LogController) GetLogRule(ctx *gin.Context) {
	ruleID := ctx.Param("rule_id")
	
	rule := c.logMonitor.GetRule(ruleID)
	if rule == nil {
		c.Error(ctx, http.StatusNotFound, "监控规则不存在")
		return
	}
	
	c.Success(ctx, gin.H{
		"rule":      rule,
		"timestamp": time.Now(),
	}, "获取监控规则成功")
}

// CreateLogRule 创建监控规则
// @Summary 创建监控规则
// @Description 创建新的日志监控规则
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rule body CreateLogRuleRequest true "规则信息"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/rules [post]
func (c *LogController) CreateLogRule(ctx *gin.Context) {
	var req CreateLogRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	// 验证必填字段
	if req.Name == "" || req.Logger == "" || req.Threshold <= 0 {
		c.Error(ctx, http.StatusBadRequest, "规则名称、日志记录器和阈值不能为空")
		return
	}
	
	// 创建规则
	rule := &Services.LogRule{
		ID:          generateRuleID(),
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Logger:      req.Logger,
		Level:       req.Level,
		Pattern:     req.Pattern,
		Keywords:    req.Keywords,
		Fields:      req.Fields,
		Threshold:   req.Threshold,
		TimeWindow:  req.TimeWindow,
		AlertLevel:  req.AlertLevel,
		Message:     req.Message,
		Actions:     req.Actions,
	}
	
	// 添加规则
	if err := c.logMonitor.AddRule(rule); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "创建监控规则失败: "+err.Error())
		return
	}
	
	// 记录审计日志
	c.logManager.LogAudit(ctx, "create_log_rule", "log_rule", rule.ID, map[string]interface{}{
		"rule_name": rule.Name,
		"logger":    rule.Logger,
		"threshold": rule.Threshold,
	})
	
	c.Success(ctx, gin.H{
		"rule":      rule,
		"message":   "监控规则创建成功",
		"timestamp": time.Now(),
	}, "监控规则创建成功")
}

// UpdateLogRule 更新监控规则
// @Summary 更新监控规则
// @Description 更新指定的监控规则
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rule_id path string true "规则ID"
// @Param updates body UpdateLogRuleRequest true "更新内容"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/rules/{rule_id} [put]
func (c *LogController) UpdateLogRule(ctx *gin.Context) {
	ruleID := ctx.Param("rule_id")
	
	var req UpdateLogRuleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		c.Error(ctx, http.StatusBadRequest, "请求参数错误: "+err.Error())
		return
	}
	
	// 构建更新字段
	updates := make(map[string]interface{})
	if req.Enabled != nil {
		updates["enabled"] = *req.Enabled
	}
	if req.Threshold > 0 {
		updates["threshold"] = req.Threshold
	}
	if req.TimeWindow > 0 {
		updates["time_window"] = req.TimeWindow
	}
	if req.Message != "" {
		updates["message"] = req.Message
	}
	
	// 更新规则
	if err := c.logMonitor.UpdateRule(ruleID, updates); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "更新监控规则失败: "+err.Error())
		return
	}
	
	// 记录审计日志
	c.logManager.LogAudit(ctx, "update_log_rule", "log_rule", ruleID, map[string]interface{}{
		"updates": updates,
	})
	
	c.Success(ctx, gin.H{
		"message":   "监控规则更新成功",
		"rule_id":   ruleID,
		"timestamp": time.Now(),
	}, "监控规则更新成功")
}

// DeleteLogRule 删除监控规则
// @Summary 删除监控规则
// @Description 删除指定的监控规则
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param rule_id path string true "规则ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/rules/{rule_id} [delete]
func (c *LogController) DeleteLogRule(ctx *gin.Context) {
	ruleID := ctx.Param("rule_id")
	
	// 删除规则
	if err := c.logMonitor.RemoveRule(ruleID); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "删除监控规则失败: "+err.Error())
		return
	}
	
	// 记录审计日志
	c.logManager.LogAudit(ctx, "delete_log_rule", "log_rule", ruleID, map[string]interface{}{
		"rule_id": ruleID,
	})
	
	c.Success(ctx, gin.H{
		"message":   "监控规则删除成功",
		"rule_id":   ruleID,
		"timestamp": time.Now(),
	}, "监控规则删除成功")
}

// GetLogAlerts 获取告警列表
// @Summary 获取告警列表
// @Description 获取日志监控告警列表
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param status query string false "告警状态(active, resolved, acknowledged)"
// @Param limit query int false "返回数量限制" default(50)
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/alerts [get]
func (c *LogController) GetLogAlerts(ctx *gin.Context) {
	status := ctx.Query("status")
	limitStr := ctx.DefaultQuery("limit", "50")
	
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	
	alerts := c.logMonitor.GetAlerts(status, limit)
	
	c.Success(ctx, gin.H{
		"alerts":    alerts,
		"total":     len(alerts),
		"status":    status,
		"timestamp": time.Now(),
	}, "获取告警列表成功")
}

// ResolveLogAlert 解决告警
// @Summary 解决告警
// @Description 将指定的告警标记为已解决
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param alert_id path string true "告警ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/alerts/{alert_id}/resolve [post]
func (c *LogController) ResolveLogAlert(ctx *gin.Context) {
	alertID := ctx.Param("alert_id")
	
	if err := c.logMonitor.ResolveAlert(alertID); err != nil {
		c.Error(ctx, http.StatusBadRequest, "解决告警失败: "+err.Error())
		return
	}
	
	// 记录审计日志
	_, _ = ctx.Get("user_id")
	_ = ctx.GetString("username")
	c.logManager.LogAudit(ctx, "resolve_log_alert", "log_alert", alertID, map[string]interface{}{
		"alert_id": alertID,
	})
	
	c.Success(ctx, gin.H{
		"message":   "告警已解决",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	}, "告警已解决")
}

// AcknowledgeLogAlert 确认告警
// @Summary 确认告警
// @Description 确认指定的告警
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param alert_id path string true "告警ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/monitor/alerts/{alert_id}/acknowledge [post]
func (c *LogController) AcknowledgeLogAlert(ctx *gin.Context) {
	alertID := ctx.Param("alert_id")
	
	if err := c.logMonitor.AcknowledgeAlert(alertID); err != nil {
		c.Error(ctx, http.StatusBadRequest, "确认告警失败: "+err.Error())
		return
	}
	
	// 记录审计日志
	_, _ = ctx.Get("user_id")
	_ = ctx.GetString("username")
	c.logManager.LogAudit(ctx, "acknowledge_log_alert", "log_alert", alertID, map[string]interface{}{
		"alert_id": alertID,
	})
	
	c.Success(ctx, gin.H{
		"message":   "告警已确认",
		"alert_id":  alertID,
		"timestamp": time.Now(),
	}, "告警已确认")
}

// GetLogConfig 获取日志配置
// @Summary 获取日志配置
// @Description 获取当前日志系统配置
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/config [get]
func (c *LogController) GetLogConfig(ctx *gin.Context) {
	// 返回配置信息（隐藏敏感信息）
	safeConfig := map[string]interface{}{
		"base_path":    c.config.BasePath,
		"level":        c.config.Level,
		"format":       c.config.Format,
		"output":       c.config.Output,
		"rotation":     c.config.Rotation,
		"request_log":  c.config.RequestLog.Enabled,
		"sql_log":      c.config.SQLLog.Enabled,
		"error_log":    c.config.ErrorLog.Enabled,
		"audit_log":    c.config.AuditLog.Enabled,
		"security_log": c.config.SecurityLog.Enabled,
		"business_log": c.config.BusinessLog.Enabled,
		"access_log":   c.config.AccessLog.Enabled,
	}
	
	c.Success(ctx, gin.H{
		"config":    safeConfig,
		"timestamp": time.Now(),
	}, "获取日志配置成功")
}

// GetSystemHealth 获取系统健康状态
// @Summary 获取系统健康状态
// @Description 检查日志系统的健康状态
// @Tags 日志管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Router /logs/health [get]
func (c *LogController) GetSystemHealth(ctx *gin.Context) {
	// 检查日志管理器状态
	logStats := c.logManager.GetStats()
	monitorStats := c.logMonitor.GetStats()
	
	// 检查磁盘空间（这里应该实现实际的磁盘检查）
	diskStatus := "healthy"
	
	// 检查告警状态
	alertStatus := "healthy"
	if monitorStats.ActiveAlerts > 0 {
		alertStatus = "warning"
	}
	if monitorStats.ActiveAlerts > 10 {
		alertStatus = "critical"
	}
	
	// 计算整体健康状态
	overallStatus := "healthy"
	if alertStatus == "critical" || diskStatus == "critical" {
		overallStatus = "critical"
	} else if alertStatus == "warning" || diskStatus == "warning" {
		overallStatus = "warning"
	}
	
	health := map[string]interface{}{
		"overall_status": overallStatus,
		"components": map[string]interface{}{
			"log_manager": "healthy",
			"log_monitor": alertStatus,
			"disk_space":  diskStatus,
		},
		"metrics": map[string]interface{}{
			"total_logs":      logStats.TotalLogs,
			"active_alerts":   monitorStats.ActiveAlerts,
			"active_rules":    monitorStats.ActiveRules,
		},
		"last_check": time.Now(),
	}
	
	c.Success(ctx, gin.H{
		"health":    health,
		"timestamp": time.Now(),
	}, "获取系统健康状态成功")
}

// 辅助方法

// filterStats 根据条件过滤统计信息
func (c *LogController) filterStats(stats *Services.LogStats, loggerName, level, timeRange string) *Services.LogStats {
	// 这里应该实现实际的过滤逻辑
	// 暂时返回原始统计
	return stats
}

// generateRuleID 生成规则ID
func generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}

// 请求结构体

// CreateLogRuleRequest 创建监控规则请求
type CreateLogRuleRequest struct {
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Logger      string                 `json:"logger" binding:"required"`
	Level       Config.LogLevel        `json:"level"`
	Pattern     string                 `json:"pattern"`
	Keywords    []string               `json:"keywords"`
	Fields      map[string]interface{} `json:"fields"`
	Threshold   int                    `json:"threshold" binding:"required,gt=0"`
	TimeWindow  time.Duration          `json:"time_window"`
	AlertLevel  string                 `json:"alert_level"`
	Message     string                 `json:"message"`
	Actions     []string               `json:"actions"`
}

// UpdateLogRuleRequest 更新监控规则请求
type UpdateLogRuleRequest struct {
	Enabled    *bool          `json:"enabled"`
	Threshold  int            `json:"threshold"`
	TimeWindow time.Duration  `json:"time_window"`
	Message    string         `json:"message"`
}
