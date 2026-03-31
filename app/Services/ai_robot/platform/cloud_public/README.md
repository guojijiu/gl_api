# 公网云（及内网复用）平台代码

- **`client/`**（`cloudclient`）：云平台 HTTP 与通用 JSON 解析。
- **`domain/`**：按业务拆分子包（内网 `cloud_intranet` 同样 import 这些路径）：
  - **`intent`**：意图常量、`IntentPlan`、`Plan*Intent`
  - **`summary`**：`SummarizeWithData` / `SummarizeUnsupported`
  - **`project`**：项目 ID 解析、到期列表过滤
  - **`contract`**：合同 ID 解析
  - **`parse`**：从用户问题抽 task uuid、数字 id
- **`project.go` / `task.go` / `project_article.go` / `registry.go`**：`cloudpublic` 包内按领域的 HTTP 入口。
