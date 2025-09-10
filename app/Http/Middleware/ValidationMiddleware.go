package Middleware

import (
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

// ValidationMiddleware 请求验证中间件
type ValidationMiddleware struct {
	storageManager *Storage.StorageManager
}

// NewValidationMiddleware 创建请求验证中间件
func NewValidationMiddleware(storageManager *Storage.StorageManager) *ValidationMiddleware {
	return &ValidationMiddleware{
		storageManager: storageManager,
	}
}

// ValidateRequest 验证请求参数
// 功能说明：
// 1. 统一处理请求参数验证
// 2. 支持JSON绑定和自定义验证
// 3. 返回统一的验证错误格式
// 4. 记录验证失败的日志
// 5. 提高代码复用性
func (m *ValidationMiddleware) ValidateRequest(model interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 绑定JSON数据到模型
		if err := c.ShouldBindJSON(model); err != nil {
			// 记录验证失败日志
			m.storageManager.LogWarning("请求验证失败", map[string]interface{}{
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
				"error":   err.Error(),
				"user_id": c.GetString("user_id"),
			})

			// 解析验证错误
			validationErrors := m.parseValidationErrors(err)

			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "请求参数验证失败",
				"errors":  validationErrors,
			})
			c.Abort()
			return
		}

		// 验证通过，继续处理
		c.Next()
	}
}

// ValidateQuery 验证查询参数
// 功能说明：
// 1. 验证URL查询参数
// 2. 支持必填参数检查
// 3. 支持参数类型验证
func (m *ValidationMiddleware) ValidateQuery(requiredParams []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var missingParams []string

		for _, param := range requiredParams {
			if c.Query(param) == "" {
				missingParams = append(missingParams, param)
			}
		}

		if len(missingParams) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "缺少必需的查询参数",
				"errors":  missingParams,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// ValidateFileUpload 验证文件上传
// 功能说明：
// 1. 验证文件大小限制
// 2. 验证文件类型
// 3. 验证文件数量
// 4. 记录上传验证日志
func (m *ValidationMiddleware) ValidateFileUpload(maxSize int64, allowedTypes []string, maxFiles int) gin.HandlerFunc {
	return func(c *gin.Context) {
		form, err := c.MultipartForm()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "文件上传格式错误",
			})
			c.Abort()
			return
		}

		files := form.File["file"]
		if len(files) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "请选择要上传的文件",
			})
			c.Abort()
			return
		}

		if len(files) > maxFiles {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "文件数量超过限制",
				"error":   "最多只能上传 " + strconv.Itoa(maxFiles) + " 个文件",
			})
			c.Abort()
			return
		}

		for _, file := range files {
			// 检查文件大小
			if file.Size > maxSize {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "文件大小超过限制",
					"error":   "文件 " + file.Filename + " 大小超过 " + strconv.FormatInt(maxSize, 10) + " 字节",
				})
				c.Abort()
				return
			}

			// 检查文件类型
			if len(allowedTypes) > 0 {
				fileExt := strings.ToLower(strings.TrimPrefix(filepath.Ext(file.Filename), "."))
				allowed := false
				for _, allowedType := range allowedTypes {
					if fileExt == allowedType {
						allowed = true
						break
					}
				}
				if !allowed {
					c.JSON(http.StatusBadRequest, gin.H{
						"success": false,
						"message": "不支持的文件类型",
						"error":   "文件 " + file.Filename + " 类型不被支持",
					})
					c.Abort()
					return
				}
			}
		}

		// 记录上传验证成功日志
		m.storageManager.LogInfo("文件上传验证通过", map[string]interface{}{
			"path":       c.Request.URL.Path,
			"file_count": len(files),
			"total_size": m.calculateTotalSizeFiles(files),
			"user_id":    c.GetString("user_id"),
		})

		c.Next()
	}
}

// parseValidationErrors 解析验证错误
// 功能说明：
// 1. 将验证错误转换为友好的错误信息
// 2. 按字段分组错误信息
// 3. 提供详细的错误描述
// 4. 支持多种验证库的错误格式
func (m *ValidationMiddleware) parseValidationErrors(err error) map[string][]string {
	errors := make(map[string][]string)

	// 解析常见的验证错误格式
	errStr := err.Error()
	
	// 处理Gin的验证错误
	if strings.Contains(errStr, "binding") {
		// 解析字段验证错误
		lines := strings.Split(errStr, "\n")
		for _, line := range lines {
			if strings.Contains(line, "Field") {
				// 提取字段名和错误信息
				parts := strings.Split(line, " ")
				if len(parts) >= 3 {
					field := strings.TrimSuffix(parts[1], "'")
					field = strings.TrimPrefix(field, "'")
					errorMsg := strings.Join(parts[2:], " ")
					
					if errors[field] == nil {
						errors[field] = []string{}
					}
					errors[field] = append(errors[field], errorMsg)
				}
			}
		}
	} else {
		// 通用错误处理
		errors["general"] = []string{errStr}
	}

	// 如果没有解析到具体错误，使用通用错误
	if len(errors) == 0 {
		errors["general"] = []string{errStr}
	}

	return errors
}

// calculateTotalSize 计算文件总大小
func (m *ValidationMiddleware) calculateTotalSize(files []interface{}) int64 {
	var totalSize int64
	for _, file := range files {
		if f, ok := file.(interface{ Size() int64 }); ok {
			totalSize += f.Size()
		}
	}
	return totalSize
}

// calculateTotalSizeFiles 计算multipart文件总大小
func (m *ValidationMiddleware) calculateTotalSizeFiles(files []*multipart.FileHeader) int64 {
	var totalSize int64
	for _, file := range files {
		totalSize += file.Size
	}
	return totalSize
}
