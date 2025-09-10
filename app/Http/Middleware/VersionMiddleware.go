package Middleware

import (
	"cloud-platform-api/app/Storage"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"strings"
)

// VersionMiddleware API版本控制中间件
type VersionMiddleware struct {
	BaseMiddleware
	storageManager *Storage.StorageManager
	supportedVersions map[string]bool
	defaultVersion    string
}

// NewVersionMiddleware 创建API版本控制中间件
// 功能说明：
// 1. 初始化API版本控制中间件
// 2. 支持多个API版本的管理
// 3. 提供版本兼容性检查
// 4. 记录版本使用日志
func NewVersionMiddleware(storageManager *Storage.StorageManager) *VersionMiddleware {
	return &VersionMiddleware{
		storageManager: storageManager,
		supportedVersions: map[string]bool{
			"v1": true,
			"v2": true, // 未来版本
		},
		defaultVersion: "v1",
	}
}

// Handle 处理API版本控制
// 功能说明：
// 1. 从请求头或URL路径中提取API版本
// 2. 验证版本是否支持
// 3. 设置版本信息到上下文
// 4. 处理版本兼容性
func (m *VersionMiddleware) Handle() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从多个位置获取版本信息
		version := m.extractVersion(c)
		
		// 验证版本是否支持
		if !m.isVersionSupported(version) {
			m.logVersionError(c, version, "不支持的API版本")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "不支持的API版本",
				"error":   "支持的版本: " + strings.Join(m.getSupportedVersions(), ", "),
				"current_version": version,
			})
			c.Abort()
			return
		}

		// 检查版本是否已弃用
		if m.isVersionDeprecated(version) {
			m.logVersionWarning(c, version, "API版本已弃用")
			c.Header("X-API-Version-Deprecated", "true")
			c.Header("X-API-Version-Sunset", m.getVersionSunsetDate(version))
		}

		// 设置版本信息到上下文
		c.Set("api_version", version)
		c.Set("api_version_major", m.getVersionMajor(version))
		
		// 记录版本使用日志
		m.logVersionUsage(c, version)
		
		c.Next()
	}
}

// RequireVersion 要求特定版本
// 功能说明：
// 1. 检查请求是否使用指定版本
// 2. 支持版本范围检查
// 3. 提供版本升级建议
func (m *VersionMiddleware) RequireVersion(requiredVersion string) gin.HandlerFunc {
	return func(c *gin.Context) {
		currentVersion := c.GetString("api_version")
		if currentVersion == "" {
			currentVersion = m.defaultVersion
		}

		if currentVersion != requiredVersion {
			m.logVersionError(c, currentVersion, "版本不匹配")
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"message": "API版本不匹配",
				"error":   "需要版本: " + requiredVersion + ", 当前版本: " + currentVersion,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// extractVersion 提取版本信息
// 功能说明：
// 1. 从多个位置提取API版本
// 2. 支持URL路径、请求头、查询参数
// 3. 提供版本提取的优先级
func (m *VersionMiddleware) extractVersion(c *gin.Context) string {
	// 1. 从URL路径提取 (例如: /api/v1/users)
	path := c.Request.URL.Path
	if strings.Contains(path, "/api/v") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if strings.HasPrefix(part, "v") && i+1 < len(parts) {
				return part
			}
		}
	}

	// 2. 从请求头提取
	if version := c.GetHeader("X-API-Version"); version != "" {
		return version
	}

	// 3. 从Accept头提取
	if accept := c.GetHeader("Accept"); accept != "" {
		if strings.Contains(accept, "version=") {
			parts := strings.Split(accept, ";")
			for _, part := range parts {
				if strings.Contains(part, "version=") {
					version := strings.TrimSpace(strings.Split(part, "=")[1])
					return strings.Trim(version, "\"")
				}
			}
		}
	}

	// 4. 从查询参数提取
	if version := c.Query("version"); version != "" {
		return version
	}

	// 5. 返回默认版本
	return m.defaultVersion
}

// isVersionSupported 检查版本是否支持
func (m *VersionMiddleware) isVersionSupported(version string) bool {
	return m.supportedVersions[version]
}

// isVersionDeprecated 检查版本是否已弃用
func (m *VersionMiddleware) isVersionDeprecated(version string) bool {
	// 这里可以配置已弃用的版本
	deprecatedVersions := map[string]bool{
		// "v1": true, // 示例：v1版本已弃用
	}
	return deprecatedVersions[version]
}

// getVersionSunsetDate 获取版本弃用日期
func (m *VersionMiddleware) getVersionSunsetDate(version string) string {
	// 这里可以配置版本弃用日期
	sunsetDates := map[string]string{
		// "v1": "2024-12-31", // 示例
	}
	return sunsetDates[version]
}

// getVersionMajor 获取版本主版本号
func (m *VersionMiddleware) getVersionMajor(version string) int {
	if strings.HasPrefix(version, "v") {
		if major, err := strconv.Atoi(version[1:]); err == nil {
			return major
		}
	}
	return 1
}

// getSupportedVersions 获取支持的版本列表
func (m *VersionMiddleware) getSupportedVersions() []string {
	var versions []string
	for version := range m.supportedVersions {
		versions = append(versions, version)
	}
	return versions
}

// logVersionUsage 记录版本使用日志
func (m *VersionMiddleware) logVersionUsage(c *gin.Context, version string) {
	m.storageManager.LogInfo("API版本使用", map[string]interface{}{
		"version":   version,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
		"user_agent": c.Request.UserAgent(),
	})
}

// logVersionError 记录版本错误日志
func (m *VersionMiddleware) logVersionError(c *gin.Context, version, reason string) {
	m.storageManager.LogError("API版本错误", map[string]interface{}{
		"version":   version,
		"reason":    reason,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
	})
}

// logVersionWarning 记录版本警告日志
func (m *VersionMiddleware) logVersionWarning(c *gin.Context, version, reason string) {
	m.storageManager.LogWarning("API版本警告", map[string]interface{}{
		"version":   version,
		"reason":    reason,
		"path":      c.Request.URL.Path,
		"method":    c.Request.Method,
		"client_ip": c.ClientIP(),
	})
}

// SetSupportedVersions 设置支持的版本
func (m *VersionMiddleware) SetSupportedVersions(versions []string) {
	m.supportedVersions = make(map[string]bool)
	for _, version := range versions {
		m.supportedVersions[version] = true
	}
}

// SetDefaultVersion 设置默认版本
func (m *VersionMiddleware) SetDefaultVersion(version string) {
	m.defaultVersion = version
}
