package Controllers

import (
	"cloud-platform-api/app/Models"
	"cloud-platform-api/app/Services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SecurityController 安全防护控制器
type SecurityController struct {
	Controller
	securityService *Services.SecurityService
}

// NewSecurityController 创建安全防护控制器
func NewSecurityController() *SecurityController {
	return &SecurityController{}
}

// SetSecurityService 设置安全防护服务
func (c *SecurityController) SetSecurityService(service *Services.SecurityService) {
	c.securityService = service
}

// GetSecurityEvents 获取安全事件列表
func (c *SecurityController) GetSecurityEvents(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.SecurityEvent{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var events []Models.SecurityEvent
	offset := (page - 1) * limit
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&events)

	c.Success(ctx, gin.H{
		"events":     events,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "安全事件列表获取成功")
}

// GetThreatIntelligence 获取威胁情报列表
func (c *SecurityController) GetThreatIntelligence(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.ThreatIntelligence{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var threats []Models.ThreatIntelligence
	offset := (page - 1) * limit
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&threats)

	c.Success(ctx, gin.H{
		"threats":    threats,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "威胁情报列表获取成功")
}

// GetLoginAttempts 获取登录尝试记录
func (c *SecurityController) GetLoginAttempts(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.LoginAttempt{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var attempts []Models.LoginAttempt
	offset := (page - 1) * limit
	query.Order("attempt_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&attempts)

	c.Success(ctx, gin.H{
		"attempts":   attempts,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "登录尝试记录获取成功")
}

// GetAccountLockouts 获取账户锁定记录
func (c *SecurityController) GetAccountLockouts(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.AccountLockout{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var lockouts []Models.AccountLockout
	offset := (page - 1) * limit
	query.Order("locked_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&lockouts)

	c.Success(ctx, gin.H{
		"lockouts":   lockouts,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "账户锁定记录获取成功")
}

// GetSecurityAlerts 获取安全告警列表
func (c *SecurityController) GetSecurityAlerts(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.SecurityAlert{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var alerts []Models.SecurityAlert
	offset := (page - 1) * limit
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&alerts)

	c.Success(ctx, gin.H{
		"alerts":     alerts,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "安全告警列表获取成功")
}

// GetSecurityReports 获取安全报告列表
func (c *SecurityController) GetSecurityReports(ctx *gin.Context) {
	if c.securityService == nil {
		c.Error(ctx, http.StatusInternalServerError, "安全防护服务未初始化")
		return
	}

	// 获取查询参数
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "20"))

	// 构建查询条件
	query := c.securityService.GetDB().Model(&Models.SecurityReport{})

	// 获取总数
	var total int64
	query.Count(&total)

	// 分页查询
	var reports []Models.SecurityReport
	offset := (page - 1) * limit
	query.Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&reports)

	c.Success(ctx, gin.H{
		"reports":    reports,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"total_pages": (int(total) + limit - 1) / limit,
	}, "安全报告列表获取成功")
}
