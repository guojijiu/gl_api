package Requests

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// Request 基础请求接口
type Request interface {
	Validate(ctx *gin.Context) error
}

// BaseRequest 基础请求结构
type BaseRequest struct {
	// 可以在这里添加通用的请求字段
}

// Validate 基础验证方法
func (r *BaseRequest) Validate(ctx *gin.Context) error {
	// 基础验证逻辑
	return nil
}

// ValidateRequest 验证请求的通用方法
func ValidateRequest(ctx *gin.Context, request Request) error {
	if err := ctx.ShouldBindJSON(request); err != nil {
		return err
	}
	
	return request.Validate(ctx)
}

// HandleValidationError 处理验证错误的通用方法
func HandleValidationError(ctx *gin.Context, err error) {
	ctx.JSON(http.StatusBadRequest, gin.H{
		"success": false,
		"message": "Validation failed",
		"error":   err.Error(),
	})
}

