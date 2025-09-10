package Controllers

import (
	"cloud-platform-api/app/Services"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// QueryOptimizationController 查询优化控制器
type QueryOptimizationController struct {
	Controller
	queryOptService *Services.QueryOptimizationService
}

// NewQueryOptimizationController 创建查询优化控制器
func NewQueryOptimizationController() *QueryOptimizationController {
	return &QueryOptimizationController{}
}

// SetQueryOptimizationService 设置查询优化服务
func (c *QueryOptimizationController) SetQueryOptimizationService(service *Services.QueryOptimizationService) {
	c.queryOptService = service
}

// GetSlowQueries 获取慢查询列表
// @Summary 获取慢查询列表
// @Description 获取系统中记录的慢查询信息
// @Tags 查询优化
// @Accept json
// @Produce json
// @Param limit query int false "限制返回数量" default(50)
// @Param warning_level query string false "警告级别" Enums(WARNING,CRITICAL)
// @Success 200 {object} Response "慢查询列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/slow-queries [get]
func (c *QueryOptimizationController) GetSlowQueries(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	// 获取查询参数
	limitStr := ctx.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.Error(ctx, http.StatusBadRequest, "无效的limit参数")
		return
	}

	warningLevel := ctx.Query("warning_level")

	// 获取慢查询列表
	slowQueries := c.queryOptService.GetSlowQueries(limit, warningLevel)

	c.Success(ctx, gin.H{
		"slow_queries":  slowQueries,
		"total":         len(slowQueries),
		"limit":         limit,
		"warning_level": warningLevel,
	}, "慢查询列表获取成功")
}

// GetQueryStatistics 获取查询统计信息
// @Summary 获取查询统计信息
// @Description 获取数据库查询的统计数据，包括执行次数、平均耗时等
// @Tags 查询优化
// @Accept json
// @Produce json
// @Success 200 {object} Response "查询统计信息"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/query-statistics [get]
func (c *QueryOptimizationController) GetQueryStatistics(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	// 获取查询统计
	queryStats := c.queryOptService.GetQueryStatistics()

	// 转换为数组格式便于前端处理
	statsArray := make([]interface{}, 0, len(queryStats))
	for _, stats := range queryStats {
		statsArray = append(statsArray, stats)
	}

	c.Success(ctx, gin.H{
		"query_statistics": statsArray,
		"total":            len(queryStats),
	}, "查询统计信息获取成功")
}

// GetIndexSuggestions 获取索引建议
// @Summary 获取索引建议
// @Description 获取系统生成的数据库索引优化建议
// @Tags 查询优化
// @Accept json
// @Produce json
// @Param status query string false "建议状态" Enums(PENDING,APPLIED,REJECTED)
// @Success 200 {object} Response "索引建议列表"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/index-suggestions [get]
func (c *QueryOptimizationController) GetIndexSuggestions(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	status := ctx.Query("status")

	// 获取索引建议
	suggestions := c.queryOptService.GetIndexSuggestions(status)

	c.Success(ctx, gin.H{
		"index_suggestions": suggestions,
		"total":             len(suggestions),
		"status":            status,
	}, "索引建议列表获取成功")
}

// ApplyIndexSuggestion 应用索引建议
// @Summary 应用索引建议
// @Description 应用指定的索引建议，在数据库中创建相应的索引
// @Tags 查询优化
// @Accept json
// @Produce json
// @Param suggestion_id path string true "建议ID"
// @Success 200 {object} Response "应用成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/index-suggestions/{suggestion_id}/apply [post]
func (c *QueryOptimizationController) ApplyIndexSuggestion(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	suggestionID := ctx.Param("suggestion_id")
	if suggestionID == "" {
		c.Error(ctx, http.StatusBadRequest, "缺少建议ID")
		return
	}

	// 应用索引建议
	if err := c.queryOptService.ApplyIndexSuggestion(suggestionID); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "应用索引建议失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message":       "索引建议应用成功",
		"suggestion_id": suggestionID,
	}, "索引建议应用成功")
}

// RejectIndexSuggestion 拒绝索引建议
// @Summary 拒绝索引建议
// @Description 拒绝指定的索引建议
// @Tags 查询优化
// @Accept json
// @Produce json
// @Param suggestion_id path string true "建议ID"
// @Success 200 {object} Response "拒绝成功"
// @Failure 400 {object} Response "请求参数错误"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/index-suggestions/{suggestion_id}/reject [post]
func (c *QueryOptimizationController) RejectIndexSuggestion(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	suggestionID := ctx.Param("suggestion_id")
	if suggestionID == "" {
		c.Error(ctx, http.StatusBadRequest, "缺少建议ID")
		return
	}

	// 拒绝索引建议
	if err := c.queryOptService.RejectIndexSuggestion(suggestionID); err != nil {
		c.Error(ctx, http.StatusInternalServerError, "拒绝索引建议失败: "+err.Error())
		return
	}

	c.Success(ctx, gin.H{
		"message":       "索引建议已拒绝",
		"suggestion_id": suggestionID,
	}, "索引建议已拒绝")
}

// GetPerformanceReport 获取性能报告
// @Summary 获取性能报告
// @Description 获取系统性能监控报告，包括响应时间、吞吐量、错误率等指标
// @Tags 查询优化
// @Accept json
// @Produce json
// @Success 200 {object} Response "性能报告"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/performance-report [get]
func (c *QueryOptimizationController) GetPerformanceReport(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	// 获取性能报告 - 使用默认时间范围（最近24小时）
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	performanceReport := c.queryOptService.GetPerformanceReport(startTime, now)

	c.Success(ctx, gin.H{
		"performance_report": performanceReport,
	}, "性能报告获取成功")
}

// GenerateOptimizationReport 生成优化报告
// @Summary 生成优化报告
// @Description 生成完整的数据库优化报告文件
// @Tags 查询优化
// @Accept json
// @Produce json
// @Success 200 {object} Response "报告生成成功"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/generate-report [post]
func (c *QueryOptimizationController) GenerateOptimizationReport(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	// 生成优化报告
	report := c.queryOptService.GenerateOptimizationReport()

	c.Success(ctx, gin.H{
		"message": "优化报告生成成功",
		"report":  report,
	}, "优化报告生成成功")
}

// GetOptimizationSummary 获取优化摘要
// @Summary 获取优化摘要
// @Description 获取查询优化的整体摘要信息
// @Tags 查询优化
// @Accept json
// @Produce json
// @Success 200 {object} Response "优化摘要"
// @Failure 500 {object} Response "服务器错误"
// @Router /api/v1/query-optimization/summary [get]
func (c *QueryOptimizationController) GetOptimizationSummary(ctx *gin.Context) {
	if c.queryOptService == nil {
		c.Error(ctx, http.StatusInternalServerError, "查询优化服务未初始化")
		return
	}

	// 获取各项数据
	slowQueries := c.queryOptService.GetSlowQueries(0, "")
	queryStats := c.queryOptService.GetQueryStatistics()
	indexSuggestions := c.queryOptService.GetIndexSuggestions("")
	now := time.Now()
	startTime := now.Add(-24 * time.Hour)
	performanceReport := c.queryOptService.GetPerformanceReport(startTime, now)

	// 统计各类数据
	criticalSlowQueries := 0
	warningSlowQueries := 0
	for _, query := range slowQueries {
		if query.WarningLevel == "CRITICAL" {
			criticalSlowQueries++
		} else if query.WarningLevel == "WARNING" {
			warningSlowQueries++
		}
	}

	pendingSuggestions := 0
	appliedSuggestions := 0
	rejectedSuggestions := 0
	for _, suggestion := range indexSuggestions {
		switch suggestion.Status {
		case "PENDING":
			pendingSuggestions++
		case "APPLIED":
			appliedSuggestions++
		case "REJECTED":
			rejectedSuggestions++
		}
	}

	// 构建摘要数据
	summary := gin.H{
		"slow_queries": gin.H{
			"total":    len(slowQueries),
			"critical": criticalSlowQueries,
			"warning":  warningSlowQueries,
		},
		"query_statistics": gin.H{
			"total_unique_queries": len(queryStats),
		},
		"index_suggestions": gin.H{
			"total":    len(indexSuggestions),
			"pending":  pendingSuggestions,
			"applied":  appliedSuggestions,
			"rejected": rejectedSuggestions,
		},
		"performance":   performanceReport,
		"health_status": c.calculateHealthStatus(performanceReport),
	}

	c.Success(ctx, gin.H{
		"summary": summary,
	}, "优化摘要获取成功")
}

// calculateHealthStatus 计算系统健康状态
func (c *QueryOptimizationController) calculateHealthStatus(performanceReport map[string]interface{}) string {
	thresholdsMet, ok := performanceReport["thresholds_met"].(map[string]bool)
	if !ok {
		return "UNKNOWN"
	}

	allMet := true
	criticalFailed := false

	for metric, met := range thresholdsMet {
		if !met {
			allMet = false
			// 如果错误率或P99响应时间超标，视为严重问题
			if metric == "error_rate" || metric == "p99_response_time" {
				criticalFailed = true
			}
		}
	}

	if criticalFailed {
		return "CRITICAL"
	} else if !allMet {
		return "WARNING"
	} else {
		return "HEALTHY"
	}
}
