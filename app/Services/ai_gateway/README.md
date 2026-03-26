# AI Gateway 模块说明

本目录是 AI 网关能力的统一实现，负责把用户自然语言问题转换为云平台接口调用，再将结构化结果整理为可读回复。

## 目录结构

- `service.go`  
  模块入口：`Service` 结构体与构造函数。

- `intents.go`  
  统一定义意图常量（`intent`）和 prompt 枚举构造函数。

- `utils.go`  
  front 风格响应解析工具（如提取 `url`、`file_path`、`id`）。

- `clients.go`  
  面向控制器的轻量客户端封装：`ProjectClient`、`ContractClient`、`TaskClient`。

- `llm.go`  
  大模型相关能力：意图规划、结果总结、OpenAI 兼容调用（百炼）。

- `cloud_client.go`  
  云平台 HTTP 调用基础层（GET/POST JSON，Token 透传）。

- `project_domain.go`  
  项目/合同域逻辑：项目列表、合同列表、下载链接、ID 解析、到期筛选。

- `task_domain.go`  
  任务域逻辑：任务状态、结果下载、模块化任务、UUID/数字 ID 提取。

- `service_test.go`  
  当前模块关键单测（strict mode、多匹配等）。

## 调用链路（简化）

1. 控制器接收请求（`platform`、`question_type`、`question`）。
2. 基于平台/业务分发到对应 handler。
3. handler 调 `Service.Plan*Intent()` 做 intent 路由。
4. 根据 intent 调用项目域/任务域方法请求云平台。
5. 调 `Service.SummarizeWithData()` 生成自然语言回复。

## 能力矩阵关系

平台/业务/intent 的可用性由配置控制（`AI_GATEWAY_CAPABILITY_MATRIX_JSON`），运行时会校验：

- 平台是否支持
- 平台下业务类型是否支持
- 业务类型下 intent 是否支持

建议新增能力时同时更新：

1. 配置能力矩阵
2. `intents.go` 常量
3. 对应域 handler 注册表
4. 最小单测

## 扩展建议（KISS）

- 新增一个业务域时，优先加新文件，不改已有大文件。
- intent 名称保持短小、稳定，避免频繁重命名。
- 所有云平台接口调用统一走 `cloud_client.go`，避免重复 HTTP 处理逻辑。
- JSON 解析统一复用 `utils.go`，不要在 handler 内散落解析代码。

## 开发自检

建议每次改动后执行：

```bash
go test ./app/Services/ai_gateway -v
go test ./app/Http/Requests -v
go build ./...
```

