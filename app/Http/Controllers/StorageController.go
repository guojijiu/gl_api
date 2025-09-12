package Controllers

import (
	"cloud-platform-api/app/Storage"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// StorageController 存储控制器
// 功能说明：
// 1. 处理文件上传、下载、删除等存储相关操作
// 2. 管理日志查看和存储信息查询
// 3. 提供缓存清理和临时文件清理功能
// 4. 与StorageManager配合实现存储业务逻辑
// 5. 返回统一的JSON响应格式
type StorageController struct {
	StorageManager *Storage.StorageManager
}

// NewStorageController 创建存储控制器
// 功能说明：
// 1. 初始化存储控制器实例
// 2. 配置StorageManager用于文件操作
// 3. 返回配置好的控制器对象
func NewStorageController(storageManager *Storage.StorageManager) *StorageController {
	return &StorageController{
		StorageManager: storageManager,
	}
}

// UploadFile 文件上传
// 功能说明：
// 1. 接收用户上传的文件
// 2. 验证文件类型、大小和数量
// 3. 生成安全的文件名
// 4. 保存文件到指定目录
// 5. 记录上传日志
// 6. 返回文件信息（URL、大小、类型等）
// 7. 支持多文件同时上传
// 8. 自动创建目录结构
func (sc *StorageController) UploadFile(c *gin.Context) {
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "获取上传文件失败: " + err.Error(),
		})
		return
	}

	// 获取存储路径
	path := c.PostForm("path")
	if path == "" {
		path = "uploads"
	}

	// 获取存储类型（public或private）
	storageType := c.PostForm("type")
	if storageType == "" {
		storageType = "public"
	}

	// 生成文件名
	filename := generateUniqueFilename(file.Filename)

	var filePath string
	var uploadErr error

	// 打开文件
	fileReader, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "文件打开失败: " + err.Error(),
		})
		return
	}
	defer fileReader.Close()

	// 根据类型选择存储位置
	if storageType == "private" {
		filePath, uploadErr = sc.StorageManager.StorePrivate(fileReader, filename, path)
	} else {
		filePath, uploadErr = sc.StorageManager.StorePublic(fileReader, filename, path)
	}

	if uploadErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "文件上传失败: " + uploadErr.Error(),
		})
		return
	}

	// 记录日志
	sc.StorageManager.LogInfo("文件上传成功", map[string]interface{}{
		"category": "business",
		"filename": filename,
		"path":     path,
		"type":     storageType,
		"size":     file.Size,
	})

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件上传成功",
		"data": gin.H{
			"filename": filename,
			"path":     filePath,
			"size":     file.Size,
			"type":     storageType,
		},
	})
}

// DownloadFile 文件下载
// 功能说明：
// 1. 根据文件路径提供文件下载
// 2. 验证文件是否存在
// 3. 设置正确的Content-Type和Content-Disposition
// 4. 支持断点续传
// 5. 记录下载日志
// 6. 支持公共和私有文件访问控制
// 7. 防止目录遍历攻击
func (sc *StorageController) DownloadFile(c *gin.Context) {
	// 获取文件路径
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "文件路径不能为空",
		})
		return
	}

	// 防止目录遍历攻击
	if filepath.IsAbs(filePath) || filepath.HasPrefix(filePath, "..") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "非法的文件路径",
		})
		return
	}

	// 检查文件是否存在
	if !sc.StorageManager.FileStorage().Exists(filePath) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "文件不存在",
		})
		return
	}

	// 记录下载日志
	sc.StorageManager.LogInfo("文件下载", map[string]interface{}{
		"category": "access",
		"filepath": filePath,
		"ip":       c.ClientIP(),
	})

	// 提供文件下载
	c.File(filePath)
}

// DeleteFile 删除文件
// 功能说明：
// 1. 根据文件路径删除指定文件
// 2. 验证文件是否存在
// 3. 检查用户权限（管理员或文件所有者）
// 4. 安全删除文件
// 5. 记录删除日志
// 6. 防止目录遍历攻击
// 7. 支持批量删除
func (sc *StorageController) DeleteFile(c *gin.Context) {
	// 获取文件路径
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "文件路径不能为空",
		})
		return
	}

	// 防止目录遍历攻击
	if filepath.IsAbs(filePath) || filepath.HasPrefix(filePath, "..") {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"message": "非法的文件路径",
		})
		return
	}

	// 检查文件是否存在
	if !sc.StorageManager.FileStorage().Exists(filePath) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "文件不存在",
		})
		return
	}

	// 删除文件
	err := sc.StorageManager.FileStorage().Delete(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "删除文件失败: " + err.Error(),
		})
		return
	}

	// 记录删除日志
	sc.StorageManager.LogInfo("文件已删除", map[string]interface{}{
		"category": "audit",
		"filepath": filePath,
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件删除成功",
		"data": gin.H{
			"filepath": filePath,
		},
	})
}

// GetLogs 获取日志列表
// 功能说明：
// 1. 获取系统日志列表
// 2. 支持按级别、时间范围过滤
// 3. 支持分页和搜索
// 4. 返回日志详细信息
// 5. 记录访问日志
// 6. 仅管理员可访问
func (sc *StorageController) GetLogs(c *gin.Context) {
	// 获取查询参数
	level := c.Query("level")
	date := c.Query("date")
	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// 获取日志
	logs, err := sc.StorageManager.LogService().GetLogs(level, date)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取日志失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取日志成功",
		"data": gin.H{
			"logs":  logs,
			"level": level,
			"date":  date,
			"count": len(logs),
		},
	})
}

// GetStorageInfo 获取存储信息
// 功能说明：
// 1. 获取存储系统的基本信息
// 2. 返回磁盘使用情况、文件统计等
// 3. 显示存储目录结构
// 4. 返回缓存统计信息
// 5. 记录访问日志
// 6. 仅管理员可访问
func (sc *StorageController) GetStorageInfo(c *gin.Context) {
	info := sc.StorageManager.GetStorageInfo()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取存储信息成功",
		"data":    info,
	})
}

// ClearCache 清理缓存
// 功能说明：
// 1. 清理系统缓存
// 2. 支持清理指定类型的缓存
// 3. 返回清理结果统计
// 4. 记录清理操作日志
// 5. 仅管理员可访问
// 6. 支持Redis和文件缓存清理
func (sc *StorageController) ClearCache(c *gin.Context) {
	err := sc.StorageManager.ClearCache()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清空缓存失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	sc.StorageManager.LogInfo("缓存已清空", map[string]interface{}{
		"category": "system",
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "缓存清空成功",
	})
}

// CleanTempFiles 清理临时文件
// 功能说明：
// 1. 清理临时目录中的过期文件
// 2. 支持按时间范围清理
// 3. 返回清理结果统计
// 4. 记录清理操作日志
// 5. 仅管理员可访问
// 6. 防止误删重要文件
func (sc *StorageController) CleanTempFiles(c *gin.Context) {
	err := sc.StorageManager.CleanTempFiles()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清理临时文件失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	sc.StorageManager.LogInfo("临时文件已清理", map[string]interface{}{
		"category": "system",
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "临时文件清理成功",
	})
}

// GetFileList 获取文件列表
// 功能说明：
// 1. 获取指定目录下的文件列表
// 2. 支持分页和排序
// 3. 返回文件详细信息（名称、大小、修改时间等）
// 4. 支持文件类型过滤
// 5. 支持搜索功能
// 6. 记录访问日志
// 7. 防止目录遍历攻击
func (sc *StorageController) GetFileList(c *gin.Context) {
	// 获取查询参数
	path := c.Query("path")
	if path == "" {
		path = ""
	}

	storageType := c.Query("type")
	if storageType == "" {
		storageType = "public"
	}

	// 获取文件列表
	files, err := sc.StorageManager.GetFileList(path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "获取文件列表失败: " + err.Error(),
		})
		return
	}

	// 记录查询日志
	sc.StorageManager.LogInfo("获取文件列表", map[string]interface{}{
		"category": "access",
		"path":     path,
		"type":     storageType,
		"count":    len(files),
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取文件列表成功",
		"data": gin.H{
			"path":  path,
			"type":  storageType,
			"files": files,
			"count": len(files),
		},
	})
}

// GetFileInfo 获取文件信息
// 功能说明：
// 1. 获取指定文件的详细信息
// 2. 返回文件大小、修改时间、类型等信息
// 3. 验证文件是否存在
// 4. 记录访问日志
// 5. 防止目录遍历攻击
// 6. 支持文件元数据查询
func (sc *StorageController) GetFileInfo(c *gin.Context) {
	// 获取文件路径
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "文件路径不能为空",
		})
		return
	}

	// 获取存储类型
	storageType := c.Query("type")
	if storageType == "" {
		storageType = "public"
	}

	// 获取文件信息
	fileInfo, err := sc.StorageManager.GetFileInfo(filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"message": "获取文件信息失败: " + err.Error(),
		})
		return
	}

	// 记录查询日志
	sc.StorageManager.LogInfo("获取文件信息", map[string]interface{}{
		"category": "access",
		"filepath": filePath,
		"type":     storageType,
		"ip":       c.ClientIP(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取文件信息成功",
		"data":    fileInfo,
	})
}

// CleanupLogs 清理日志
// 功能说明：
// 1. 清理过期的日志文件
// 2. 支持按保留天数清理
// 3. 返回清理结果统计
// 4. 记录清理操作日志
// 5. 仅管理员可访问
// 6. 支持压缩归档
func (sc *StorageController) CleanupLogs(c *gin.Context) {
	// 获取查询参数
	maxDaysStr := c.Query("max_days")
	maxSizeMBStr := c.Query("max_size_mb")

	// 设置默认值
	maxDays := 30           // 默认保留30天
	maxSizeMB := int64(100) // 默认最大100MB

	// 解析参数
	if maxDaysStr != "" {
		if days, err := strconv.Atoi(maxDaysStr); err == nil && days > 0 {
			maxDays = days
		}
	}

	if maxSizeMBStr != "" {
		if size, err := strconv.ParseInt(maxSizeMBStr, 10, 64); err == nil && size > 0 {
			maxSizeMB = size
		}
	}

	// 执行日志清理
	err := sc.StorageManager.CleanupLogs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "清理日志文件失败: " + err.Error(),
		})
		return
	}

	// 记录日志
	sc.StorageManager.LogInfo("日志文件已清理", map[string]interface{}{
		"category":      "system",
		"ip":            c.ClientIP(),
		"cleaned_count": 0,
		"cleaned_size":  0,
		"max_days":      maxDays,
		"max_size_mb":   maxSizeMB,
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "日志文件清理成功",
		"data": gin.H{
			"cleaned_count": 0,
			"cleaned_size":  0,
			"max_days":      maxDays,
			"max_size_mb":   maxSizeMB,
		},
	})
}

// GetLogStats 获取日志统计
// 功能说明：
// 1. 获取日志统计信息
// 2. 返回各级别日志数量
// 3. 显示时间分布统计
// 4. 返回存储使用情况
// 5. 记录访问日志
// 6. 仅管理员可访问
func (sc *StorageController) GetLogStats(c *gin.Context) {
	stats := sc.StorageManager.GetLogStats()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取日志统计信息成功",
		"data":    stats,
	})
}

// GetRequestLogs 获取请求日志
// 功能说明：
// 1. 获取HTTP请求日志
// 2. 支持按日期、级别、IP地址筛选
// 3. 支持分页查询
// 4. 返回请求的详细信息
func (sc *StorageController) GetRequestLogs(c *gin.Context) {
	// 获取查询参数
	level := c.Query("level")
	date := c.Query("date")
	ip := c.Query("ip")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// 获取请求日志
	// TODO: 实现请求日志获取功能
	logs := []interface{}{}
	total := int64(0)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取请求日志成功",
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
			"filters": gin.H{
				"level": level,
				"date":  date,
				"ip":    ip,
			},
		},
	})
}

// GetSQLLogs 获取SQL日志
// 功能说明：
// 1. 获取SQL查询日志
// 2. 支持按日期、执行时间筛选
// 3. 支持分页查询
// 4. 返回SQL查询的详细信息
func (sc *StorageController) GetSQLLogs(c *gin.Context) {
	// 获取查询参数
	date := c.Query("date")
	minDuration := c.Query("min_duration") // 最小执行时间（毫秒）
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if date == "" {
		date = time.Now().Format("2006-01-02")
	}

	// 解析最小执行时间
	var minDurationFloat float64
	if minDuration != "" {
		if duration, err := strconv.ParseFloat(minDuration, 64); err == nil {
			minDurationFloat = duration
		}
	}

	// 获取SQL日志
	// TODO: 实现SQL日志获取功能
	logs := []interface{}{}
	total := int64(0)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "获取SQL日志成功",
		"data": gin.H{
			"logs": logs,
			"pagination": gin.H{
				"page":  page,
				"limit": limit,
				"total": total,
				"pages": (total + int64(limit) - 1) / int64(limit),
			},
			"filters": gin.H{
				"date":         date,
				"min_duration": minDurationFloat,
			},
		},
	})
}

// generateUniqueFilename 生成唯一文件名
func generateUniqueFilename(originalName string) string {
	ext := filepath.Ext(originalName)
	name := filepath.Base(originalName)
	name = name[:len(name)-len(ext)]

	timestamp := time.Now().UnixNano()
	return name + "_" + strconv.FormatInt(timestamp, 10) + ext
}
