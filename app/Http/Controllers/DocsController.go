package Controllers

import (
	"cloud-platform-api/app/Storage"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

// DocsController API文档控制器
type DocsController struct {
	Controller
	storageManager *Storage.StorageManager
	apiVersion     string
	buildTime      string
	gitCommit      string
}

// NewDocsController 创建API文档控制器
func NewDocsController() *DocsController {
	return &DocsController{
		storageManager: Storage.GetStorageManager(),
		apiVersion:     "1.0.0",
		buildTime:      getBuildTime(),
		gitCommit:      getGitCommit(),
	}
}

// APIDoc API文档结构
type APIDoc struct {
	OpenAPI    string                 `json:"openapi"`
	Info       APIInfo                `json:"info"`
	Servers    []APIServer            `json:"servers"`
	Paths      map[string]APIPath     `json:"paths"`
	Components APIComponents          `json:"components"`
	Tags       []APITag               `json:"tags"`
	Extensions map[string]interface{} `json:"-"`
}

// APIInfo API信息
type APIInfo struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	Version        string `json:"version"`
	TermsOfService string `json:"termsOfService,omitempty"`
	Contact        struct {
		Name  string `json:"name,omitempty"`
		URL   string `json:"url,omitempty"`
		Email string `json:"email,omitempty"`
	} `json:"contact,omitempty"`
	License struct {
		Name string `json:"name,omitempty"`
		URL  string `json:"url,omitempty"`
	} `json:"license,omitempty"`
}

// APIServer API服务器信息
type APIServer struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// APIPath API路径
type APIPath struct {
	Get    *APIOperation `json:"get,omitempty"`
	Post   *APIOperation `json:"post,omitempty"`
	Put    *APIOperation `json:"put,omitempty"`
	Delete *APIOperation `json:"delete,omitempty"`
	Patch  *APIOperation `json:"patch,omitempty"`
}

// APIOperation API操作
type APIOperation struct {
	Tags        []string               `json:"tags,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	OperationID string                 `json:"operationId,omitempty"`
	Parameters  []APIParameter         `json:"parameters,omitempty"`
	RequestBody *APIRequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]APIResponse `json:"responses"`
	Security    []map[string][]string  `json:"security,omitempty"`
	Deprecated  bool                   `json:"deprecated,omitempty"`
}

// APIParameter API参数
type APIParameter struct {
	Name        string `json:"name"`
	In          string `json:"in"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Schema      struct {
		Type    string      `json:"type,omitempty"`
		Format  string      `json:"format,omitempty"`
		Enum    []string    `json:"enum,omitempty"`
		Default interface{} `json:"default,omitempty"`
	} `json:"schema,omitempty"`
}

// APIRequestBody API请求体
type APIRequestBody struct {
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
	Content     map[string]struct {
		Schema struct {
			Type       string                 `json:"type,omitempty"`
			Properties map[string]interface{} `json:"properties,omitempty"`
			Required   []string               `json:"required,omitempty"`
		} `json:"schema,omitempty"`
	} `json:"content,omitempty"`
}

// APIResponse API响应
type APIResponse struct {
	Description string `json:"description"`
	Content     map[string]struct {
		Schema struct {
			Type       string                 `json:"type,omitempty"`
			Properties map[string]interface{} `json:"properties,omitempty"`
		} `json:"schema,omitempty"`
	} `json:"content,omitempty"`
}

// APIComponents API组件
type APIComponents struct {
	Schemas         map[string]interface{} `json:"schemas,omitempty"`
	SecuritySchemes map[string]interface{} `json:"securitySchemes,omitempty"`
}

// APITag API标签
type APITag struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}

// GetAPIDocs 获取API文档
// @Summary 获取API文档
// @Description 获取完整的API文档，支持OpenAPI 3.0格式
// @Tags 文档
// @Accept json
// @Produce json
// @Success 200 {object} APIDoc
// @Router /api/v1/docs [get]
func (dc *DocsController) GetAPIDocs(c *gin.Context) {
	// 生成API文档
	apiDoc := dc.generateAPIDoc()

	// 记录访问日志
	dc.storageManager.LogInfo("API文档访问", map[string]interface{}{
		"ip":         c.ClientIP(),
		"user_agent": c.GetHeader("User-Agent"),
		"timestamp":  time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apiDoc,
	})
}

// GetSwaggerUI 获取Swagger UI
// @Summary 获取Swagger UI
// @Description 获取Swagger UI界面
// @Tags 文档
// @Accept html
// @Produce html
// @Success 200 {string} string
// @Router /api/v1/docs/ui [get]
func (dc *DocsController) GetSwaggerUI(c *gin.Context) {
	// 生成Swagger UI HTML
	html := dc.generateSwaggerUI()

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

// ExportAPIDocs 导出API文档
// @Summary 导出API文档
// @Description 导出API文档为JSON文件
// @Tags 文档
// @Accept json
// @Produce application/json
// @Success 200 {file} file
// @Router /api/v1/docs/export [get]
func (dc *DocsController) ExportAPIDocs(c *gin.Context) {
	// 生成API文档
	apiDoc := dc.generateAPIDoc()

	// 序列化为JSON
	jsonData, err := json.MarshalIndent(apiDoc, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "导出API文档失败",
			"error":   err.Error(),
		})
		return
	}

	// 设置响应头
	filename := fmt.Sprintf("api-docs-%s.json", time.Now().Format("20060102-150405"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/json")

	c.Data(http.StatusOK, "application/json", jsonData)
}

// generateAPIDoc 生成API文档
func (dc *DocsController) generateAPIDoc() *APIDoc {
	return &APIDoc{
		OpenAPI: "3.0.0",
		Info: APIInfo{
			Title:       "云平台API",
			Description: "云平台API文档，提供完整的API接口说明",
			Version:     dc.apiVersion,
			Contact: struct {
				Name  string `json:"name,omitempty"`
				URL   string `json:"url,omitempty"`
				Email string `json:"email,omitempty"`
			}{
				Name:  "云平台团队",
				Email: "support@cloudplatform.com",
			},
		},
		Servers: []APIServer{
			{
				URL:         "http://localhost:8080",
				Description: "开发环境",
			},
			{
				URL:         "https://api.cloudplatform.com",
				Description: "生产环境",
			},
		},
		Paths: dc.generatePaths(),
		Components: APIComponents{
			Schemas: dc.generateSchemas(),
			SecuritySchemes: map[string]interface{}{
				"BearerAuth": map[string]interface{}{
					"type":         "http",
					"scheme":       "bearer",
					"bearerFormat": "JWT",
				},
				"ApiKeyAuth": map[string]interface{}{
					"type": "apiKey",
					"in":   "header",
					"name": "X-API-Key",
				},
			},
		},
		Tags: dc.generateTags(),
	}
}

// generatePaths 生成API路径
func (dc *DocsController) generatePaths() map[string]APIPath {
	paths := make(map[string]APIPath)

	// 健康检查
	paths["/api/v1/health"] = APIPath{
		Get: &APIOperation{
			Tags:        []string{"健康检查"},
			Summary:     "健康检查",
			Description: "检查应用健康状态",
			OperationID: "healthCheck",
			Responses: map[string]APIResponse{
				"200": {
					Description: "健康状态正常",
					Content: map[string]struct {
						Schema struct {
							Type       string                 `json:"type,omitempty"`
							Properties map[string]interface{} `json:"properties,omitempty"`
						} `json:"schema,omitempty"`
					}{
						"application/json": {
							Schema: struct {
								Type       string                 `json:"type,omitempty"`
								Properties map[string]interface{} `json:"properties,omitempty"`
							}{
								Type: "object",
								Properties: map[string]interface{}{
									"success": map[string]interface{}{
										"type": "boolean",
									},
									"data": map[string]interface{}{
										"type": "object",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// 用户认证
	paths["/api/v1/auth/login"] = APIPath{
		Post: &APIOperation{
			Tags:        []string{"认证"},
			Summary:     "用户登录",
			Description: "用户登录获取访问令牌",
			OperationID: "userLogin",
			RequestBody: &APIRequestBody{
				Description: "登录信息",
				Required:    true,
				Content: map[string]struct {
					Schema struct {
						Type       string                 `json:"type,omitempty"`
						Properties map[string]interface{} `json:"properties,omitempty"`
						Required   []string               `json:"required,omitempty"`
					} `json:"schema,omitempty"`
				}{
					"application/json": {
						Schema: struct {
							Type       string                 `json:"type,omitempty"`
							Properties map[string]interface{} `json:"properties,omitempty"`
							Required   []string               `json:"required,omitempty"`
						}{
							Type: "object",
							Properties: map[string]interface{}{
								"email": map[string]interface{}{
									"type":    "string",
									"format":  "email",
									"example": "user@example.com",
								},
								"password": map[string]interface{}{
									"type":    "string",
									"format":  "password",
									"example": "password123",
								},
							},
							Required: []string{"email", "password"},
						},
					},
				},
			},
			Responses: map[string]APIResponse{
				"200": {
					Description: "登录成功",
				},
				"401": {
					Description: "认证失败",
				},
			},
		},
	}

	// 用户管理
	paths["/api/v1/users"] = APIPath{
		Get: &APIOperation{
			Tags:        []string{"用户管理"},
			Summary:     "获取用户列表",
			Description: "获取用户列表，支持分页和搜索",
			OperationID: "getUsers",
			Security: []map[string][]string{
				{"BearerAuth": {}},
			},
			Parameters: []APIParameter{
				{
					Name:        "page",
					In:          "query",
					Description: "页码",
					Schema: struct {
						Type    string      `json:"type,omitempty"`
						Format  string      `json:"format,omitempty"`
						Enum    []string    `json:"enum,omitempty"`
						Default interface{} `json:"default,omitempty"`
					}{
						Type:    "integer",
						Default: 1,
					},
				},
				{
					Name:        "limit",
					In:          "query",
					Description: "每页数量",
					Schema: struct {
						Type    string      `json:"type,omitempty"`
						Format  string      `json:"format,omitempty"`
						Enum    []string    `json:"enum,omitempty"`
						Default interface{} `json:"default,omitempty"`
					}{
						Type:    "integer",
						Default: 10,
					},
				},
			},
			Responses: map[string]APIResponse{
				"200": {
					Description: "获取成功",
				},
				"401": {
					Description: "未授权",
				},
			},
		},
		Post: &APIOperation{
			Tags:        []string{"用户管理"},
			Summary:     "创建用户",
			Description: "创建新用户",
			OperationID: "createUser",
			Security: []map[string][]string{
				{"BearerAuth": {}},
			},
			RequestBody: &APIRequestBody{
				Description: "用户信息",
				Required:    true,
				Content: map[string]struct {
					Schema struct {
						Type       string                 `json:"type,omitempty"`
						Properties map[string]interface{} `json:"properties,omitempty"`
						Required   []string               `json:"required,omitempty"`
					} `json:"schema,omitempty"`
				}{
					"application/json": {
						Schema: struct {
							Type       string                 `json:"type,omitempty"`
							Properties map[string]interface{} `json:"properties,omitempty"`
							Required   []string               `json:"required,omitempty"`
						}{
							Type: "object",
							Properties: map[string]interface{}{
								"name": map[string]interface{}{
									"type":    "string",
									"example": "张三",
								},
								"email": map[string]interface{}{
									"type":    "string",
									"format":  "email",
									"example": "user@example.com",
								},
								"password": map[string]interface{}{
									"type":    "string",
									"format":  "password",
									"example": "password123",
								},
							},
							Required: []string{"name", "email", "password"},
						},
					},
				},
			},
			Responses: map[string]APIResponse{
				"201": {
					Description: "创建成功",
				},
				"400": {
					Description: "请求参数错误",
				},
				"401": {
					Description: "未授权",
				},
			},
		},
	}

	return paths
}

// generateSchemas 生成数据模型
func (dc *DocsController) generateSchemas() map[string]interface{} {
	schemas := make(map[string]interface{})

	// 用户模型
	schemas["User"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"id": map[string]interface{}{
				"type":   "integer",
				"format": "int64",
			},
			"name": map[string]interface{}{
				"type": "string",
			},
			"email": map[string]interface{}{
				"type":   "string",
				"format": "email",
			},
			"created_at": map[string]interface{}{
				"type":   "string",
				"format": "date-time",
			},
			"updated_at": map[string]interface{}{
				"type":   "string",
				"format": "date-time",
			},
		},
	}

	// 错误响应模型
	schemas["ErrorResponse"] = map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"success": map[string]interface{}{
				"type": "boolean",
			},
			"message": map[string]interface{}{
				"type": "string",
			},
			"error": map[string]interface{}{
				"type": "string",
			},
		},
	}

	return schemas
}

// generateTags 生成标签
func (dc *DocsController) generateTags() []APITag {
	return []APITag{
		{
			Name:        "健康检查",
			Description: "系统健康状态检查",
		},
		{
			Name:        "认证",
			Description: "用户认证和授权",
		},
		{
			Name:        "用户管理",
			Description: "用户信息管理",
		},
		{
			Name:        "文档",
			Description: "API文档相关",
		},
	}
}

// generateSwaggerUI 生成Swagger UI
func (dc *DocsController) generateSwaggerUI() string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <title>云平台API文档</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui.css" />
    <style>
        html {
            box-sizing: border-box;
            overflow: -moz-scrollbars-vertical;
            overflow-y: scroll;
        }
        *, *:before, *:after {
            box-sizing: inherit;
        }
        body {
            margin:0;
            background: #fafafa;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@4.15.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: '/api/v1/docs',
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout",
                validatorUrl: null,
                docExpansion: "list",
                defaultModelsExpandDepth: 3,
                defaultModelExpandDepth: 3,
                displayRequestDuration: true,
                tryItOutEnabled: true,
                supportedSubmitMethods: ['get', 'post', 'put', 'delete', 'patch'],
                onComplete: function() {
                    console.log('Swagger UI loaded');
                }
            });
        };
    </script>
</body>
</html>`
}

// getBuildTime 获取构建时间
func getBuildTime() string {
	if buildTime := os.Getenv("BUILD_TIME"); buildTime != "" {
		return buildTime
	}
	return time.Now().Format("2006-01-02 15:04:05")
}

// getGitCommit 获取Git提交哈希
func getGitCommit() string {
	if gitCommit := os.Getenv("GIT_COMMIT"); gitCommit != "" {
		return gitCommit
	}
	return "unknown"
}
