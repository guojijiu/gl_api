# AI Robot 功能说明文档

本文档面向对 Go 不熟悉的同学，说明 `ai_robot` 从路由入口到业务执行的完整链路、当前能力边界、扩展方式和排障方法。

---

## 1. 路由入口

- 路由文件：`app/Http/Routes/routes.go`
- 路由分组：`/api/v1/ai_robot`
- 已开放接口：
  - `POST /api/v1/ai_robot/chat`：主问答入口
  - `GET /api/v1/ai_robot/capabilities`：查看配置能力矩阵与当前实际开放能力
  - `GET /api/v1/ai_robot/conversations`：分页查看会话列表
  - `DELETE /api/v1/ai_robot/conversations`：按条件批量删除会话
  - `GET /api/v1/ai_robot/conversation`：查看单条会话详情
  - `GET /api/v1/ai_robot/message`：按消息 ID 查看单条消息详情
  - `GET /api/v1/ai_robot/messages`：分页查看某个会话的完整消息历史
  - `GET /api/v1/ai_robot/messages/stats`：查看消息审计统计
  - `DELETE /api/v1/ai_robot/conversation`：删除单条会话

---

## 2. 分层职责（当前结构）

### Controller 层（薄控制层）

- 文件：`app/Http/Controllers/AiRobotController.go`
- 职责：
  - 参数校验（JSON + 业务校验）
  - 读取 Header `Token`
  - 读取接口显式传入的 `user_id`
  - 读取配置
  - 设置超时上下文
  - 调用服务层编排：`flow.Processor.ProcessChat(...)`
  - 对 `capabilities` / `conversation*` 接口做直接输出

Controller 不承载平台分流、意图分流、云平台调用细节。

### Service 层（业务编排）

- 主目录：`app/Services/ai_robot`
- 子模块：
  - `flow/`：主流程编排、平台分流、请求级审计信息采集
  - `platform/`：按平台拆分的独立业务实现，如 `cloud_public/`、`cloud_intranet/`
  - `internal/`：内部复用组件，如 backend 注入、上下文状态、时间工具
  - `policy/`：能力策略判断（例如意图是否允许）
  - `utils/`：工具函数（如 access_code 生成）

---

## 3. chat 请求处理链路

### 步骤 1：Controller 入参处理

`AiRobotController.Chat` 做以下校验：

1. `ShouldBindJSON` 成功
2. `AiRobotChatRequest.Validate()` 通过
3. Header `Token` 非空
4. 只要传入 `user_id`，后端会默认开启上下文持久化
5. `Config.GetAiGatewayConfig()` 不为空

### 步骤 2：进入 flow 编排

入口：`app/Services/ai_robot/flow/AiRobotProcessor.go`  
方法：`ProcessChat(...)`

这里会创建一个 `Deps`（请求级依赖容器），并进入 `dispatchByRegistry(...)`。

### 步骤 3：平台 + 业务类型分流

文件：`app/Services/ai_robot/flow/AiRobotRegistry.go`

- 先做平台和 question_type 能力校验
- 再按 `platform` 分流：
  - `cloud_public`
  - `cloud_intranet`
  - `image_compare`
- 再按 `question_type` 分流：
  - `1`（project）
  - `2`（task）
  - `3`（project_article）

### 步骤 4：意图分流与执行

- 公网云平台：
  - 项目域：`app/Services/ai_robot/platform/cloud_public/project/dispatch.go`
  - 任务域：`app/Services/ai_robot/platform/cloud_public/task/dispatch.go`
  - 项目文章域：`app/Services/ai_robot/platform/cloud_public/project_article/dispatch.go`
- 内网云平台：
  - 当前保留独立入口：`app/Services/ai_robot/platform/cloud_intranet/`
  - 聊天能力尚未正式开放，不与公网云平台做能力复用或耦合
- 图片域：当前仍为占位能力

典型流程是：

1. 大模型意图判定（`PlanProjectIntent` / `PlanTaskIntent`）
2. 判断能力矩阵是否允许该意图
3. 调云平台接口获取原始数据
4. 大模型做摘要/自然语言回答
5. 统一输出 front 格式

---

## 4. 平台边界与能力现状

### 4.1 当前对外开放能力

- `cloud_public`
  - 已开放 `question_type=1`（项目）
  - 已开放 `question_type=2`（任务）
  - 已开放 `question_type=3`（项目文章）
- `cloud_intranet`
  - 平台入口与后端注入链路已保留
  - 当前聊天能力未正式开放
  - 原则上按独立平台维护，不与 `cloud_public` 互相耦合
- `image_compare`
  - 当前仍为占位能力

### 4.2 `Capabilities` 接口怎么看

- 接口：`GET /api/v1/ai_robot/capabilities`
- 返回里有两层能力信息：
  - `configured_capability_matrix`：配置层声明，便于看环境变量/能力矩阵是否生效
  - `exposed_chat_capability_matrix`：当前接口层实际对外开放的聊天能力
- 推荐排障顺序：
  1. 先看 `exposed_chat_capability_matrix`，确认这个平台/类型当前是否真的开放
  2. 再看 `configured_capability_matrix`，确认配置是否与代码预期一致
- 这样设计是为了避免出现“配置里声明可用，但业务实现仍未开放”的误判

---

## 5. 请求体与枚举

文件：`app/Http/Requests/AiRobotRequests.go`

请求结构：

- `question`：用户自然语言问题
- `platform`：平台
  - `cloud_public`
  - `cloud_intranet`
  - `image_compare`
- `question_type`：业务类型
  - `1`：项目域
  - `2`：任务域
  - `3`：项目文章域
- `conversation_id`：会话标识；不传时后端自动生成
- `message_id`：消息标识；不传时后端自动生成
- `enable_context`：是否启用上下文持久化；传入 `user_id` 时后端默认会开启
- `user_id`：当前对话所属用户标识，由接口显式传入，支持字符串

### 5.1 chat 调用示例

#### 推荐：开启上下文

```bash
curl -X POST "http://localhost:8080/api/v1/ai_robot/chat" \
  -H "Content-Type: application/json" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN" \
  -d '{
    "question": "帮我查一下这个项目的结题报告",
    "platform": "cloud_public",
    "question_type": 1,
    "user_id": "user-1",
    "conversation_id": "conv-001",
    "message_id": "msg-001",
    "enable_context": true
  }'
```

#### 简化：不开启上下文

```bash
curl -X POST "http://localhost:8080/api/v1/ai_robot/chat" \
  -H "Content-Type: application/json" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN" \
  -d '{
    "question": "帮我查一下任务状态",
    "platform": "cloud_public",
    "question_type": 2,
    "user_id": "external-user-abc",
    "enable_context": false
  }'
```

#### 参数建议

- `user_id`：建议传业务侧稳定用户标识；字符串和数字字符串都支持
- `conversation_id`：建议同一轮连续对话保持不变；首次不传可由后端生成，后续复用返回值
- `message_id`：建议每条消息唯一；不传时后端会自动生成并在响应中返回
- 只要传 `user_id`，响应里就建议取回 `conversation_id` / `message_id` 供下一轮继续使用
- 若 `platform=cloud_intranet`，当前应视为“平台预留”而不是“聊天能力已开放”

---

## 6. 返回格式约定

`chat`、`capabilities`、会话查询接口都走统一前端格式：

- HTTP 状态固定 `200`
- 业务状态看 `code` 字段
  - `1`：成功
  - `0`：失败
- 关键字段：
  - `showMsg`
  - `debugMsg`
  - `content`
- `Capabilities` 的 `content` 内重点字段：
  - `configured_capability_matrix`
  - `exposed_chat_capability_matrix`

---

## 7. 会话管理接口

### 7.1 接口速查表

| 方法 | 路径 | 必填参数 | 说明 |
| ---- | ---- | ---- | ---- |
| `POST` | `/api/v1/ai_robot/chat` | `question` `platform` `question_type` `user_id`(上下文场景必填) | 主问答入口，Header 需带 `Token` |
| `GET` | `/api/v1/ai_robot/capabilities` | 无 | 查看配置能力矩阵、实际开放能力和关键配置 |
| `GET` | `/api/v1/ai_robot/conversations` | `user_id` | 分页查看会话列表，支持多条件筛选 |
| `GET` | `/api/v1/ai_robot/conversation` | `user_id` `platform` `conversation_id` | 查看单条会话详情 |
| `GET` | `/api/v1/ai_robot/message` | `user_id` `platform` `conversation_id` `message_id` | 查看单条消息详情 |
| `GET` | `/api/v1/ai_robot/messages` | `user_id` `platform` `conversation_id` | 分页查看会话消息历史 |
| `DELETE` | `/api/v1/ai_robot/conversation` | `user_id` `platform` `conversation_id` | 删除单条会话 |
| `DELETE` | `/api/v1/ai_robot/conversations` | `user_id` + 至少一个筛选条件 | 按条件批量删除会话 |

### 7.2 权限规则

- `ai_robot` 接口本身不参与登录权限验证
- 用户来源由接口参数中的 `user_id` 显式传入
- 会话列表/详情/删除都会按传入的 `user_id` 进行范围定位
- 会话隔离键为：`user_id + platform + conversation_id`
- `user_id` 按字符串处理，不要求必须是整型

### 7.3 获取能力矩阵

- 接口：`GET /api/v1/ai_robot/capabilities`
- 用途：查看配置能力矩阵、当前实际开放能力、模型和平台地址，便于联调/排障
- 返回重点：
  - `configured_capability_matrix`：配置层能力矩阵
  - `exposed_chat_capability_matrix`：代码层真实开放能力
  - `cloud_public_base` / `cloud_intranet_base`：平台地址
- 注意：
  - `cloud_intranet_base` 有值，不代表内网聊天能力已经开放
  - 平台地址存在仅表示该平台具备独立接入配置

### 7.4 会话列表

- 接口：`GET /api/v1/ai_robot/conversations`
- 支持参数：
  - `user_id`（必填）
  - `platform`
  - `conversation_id`
  - `question_type`
  - `start_time`
  - `end_time`
  - `page`
  - `limit`
- 时间格式支持：
  - `RFC3339`
  - `2006-01-02 15:04:05`
  - `2006-01-02`
- 返回特点：
  - 按 `updated_at` 倒序
  - 返回精简列表字段
  - 每条列表项额外包含 `latest_message`，内含最近一条消息的 `message_id`、`question`、`answer`、`created_at`
  - 详情仍走单条接口

### 7.5 单条会话详情

- 接口：`GET /api/v1/ai_robot/conversation`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`
- 可选参数：
  - `message_id`
- 返回特点：
  - 保留原始会话字段，便于排障
  - `latest_message` 优先读取独立消息集合中的最新一条，因此会尽量带全 `result_kind/result_count/result_brief/response_summary`
  - 额外返回 `messages` 明细数组，适合前端直接展示
  - `messages` 内每项包含：`message_id`、`question`、`resolved_question`、`normalized_question`、`hit_context`、`context_source`、`clarification_needed`、`clarification_reason`、`answer`、`intent`、`success`、`show_msg`、`debug_msg`、`stage`、`error_type`、`cloud_api`、`cloud_status_code`、`llm_model`、`request_payload_summary`、`result_kind`、`result_count`、`result_brief`、`response_summary`、`response_size`、`cloud_called`、`duration_ms`、`llm_duration_ms`、`cloud_duration_ms`、`created_at`
  - 额外返回 `current_message`；传了 `message_id` 时返回对应那一轮，未传时默认返回最后一轮

### 7.6 删除单条会话

- 接口：`DELETE /api/v1/ai_robot/conversation`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`
- 返回特点：
  - 先删除会话主记录，再删除该会话下的消息记录
  - 成功时会额外返回 `deleted_message_count`
  - 若会话删除成功、但消息清理失败，接口仍返回成功态，但会明确提示“消息清理失败”
  - 部分成功时，`content.data` 内会带 `message_cleanup_failed=true` 和 `message_cleanup_error`

### 7.7 单条消息详情

- 接口：`GET /api/v1/ai_robot/message`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`
  - `message_id`
- 返回特点：
  - 直接返回单条 `message`
  - `message` 内包含成功/失败结果元信息，以及 `stage`、`error_type`、`cloud_api`、`cloud_status_code`、`llm_model`、`request_payload_summary`、`result_kind`、`result_count`、`result_brief`、`response_summary`、`response_size`，同时保留 `question/resolved_question/normalized_question/hit_context/context_source/clarification_needed/clarification_reason`，适合排障
  - `request_payload_summary` 会按 `question_type` 输出结构化过滤条件摘要
  - 适合消息详情页或按消息回查

### 7.8 会话消息列表

- 接口：`GET /api/v1/ai_robot/messages`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`
- 可选参数：
  - `page`
  - `limit`
- 返回特点：
  - 优先从独立消息集合读取完整历史
  - 每条消息都带 `success`、`show_msg`、`debug_msg`、`stage`、`error_type`、`cloud_api`、`cloud_status_code`、`llm_model`、`request_payload_summary`、`result_kind`、`result_count`、`result_brief`、`response_summary`、`response_size`、`cloud_called`、`duration_ms`、`llm_duration_ms`、`cloud_duration_ms`
  - `normalized_question` 表示系统最终实际处理的标准问题，优先取 `resolved_question`，为空时回退原始 `question`
  - `hit_context` 表示这次是否命中过上下文增强
  - `context_source` 当前细化记录为 `none / recent_turn / summary`
  - `clarification_needed` 表示这轮是否更适合先做客服式澄清
  - `clarification_reason` 当前会标记为 `empty_question / question_too_short / follow_up_without_context / missing_core_identifier`
  - `request_payload_summary` 对 `project/task/project_article` 会尽量拆成结构化字段，方便排障和筛选
  - `result_kind` 目前会归类为 `list / detail / download / empty / error`
  - `result_count` 目前会统计 `links/details/contract_projects/content.data` 等结果数量
  - `result_brief` 是给列表直接展示的一行短描述
  - `response_summary` 会提炼 `answer`、`intent`、`uuids`、`links`、`raw_cloud_json` 的尺寸/类型概览
  - 消息集合已额外建立 `user_id + platform + updated_at`、`user_id + platform + question_type + updated_at`、`user_id + platform + result_kind + updated_at` 复合索引，便于后台筛选与统计
  - 适合长会话分页展示

### 7.9 消息审计统计

- 接口：`GET /api/v1/ai_robot/messages/stats`
- 必填参数：
  - `user_id`
- 可选参数：
  - `platform`
  - `question_type`
  - `start_time`
  - `end_time`
- 返回特点：
  - 返回 `total`、`success_count`、`failed_count`、`clarification_count`
  - 返回 `result_kind_counts` 分布，便于观察 `list/detail/download/empty/error`
  - 返回 `question_type_counts` 分布，便于看项目/任务/文章问题占比
  - 返回 `error_type_counts` 分布，便于快速定位失败类别
  - 返回 `clarification_reason_counts` 分布，便于分析最常见的澄清原因
  - 返回 `daily_trend`，按天提供 `total / failed_count / empty_count / download_count / success_without_download_count / clarification_count`
  - 适合后台审计看板、效果复盘和排障统计

### 7.10 批量删除会话

- 接口：`DELETE /api/v1/ai_robot/conversations`
- 支持参数：
  - `user_id`（必填）
  - `platform`
  - `question_type`
  - `start_time`
  - `end_time`
- 安全限制：
  - 至少需要一个筛选条件
  - 避免误删全量会话
- 返回特点：
  - 返回 `deleted_count` 表示删除的会话数
  - 成功时额外返回 `deleted_message_count` 表示清理掉的消息数
  - 若会话批量删除成功、但消息批量清理失败，接口会显式返回“部分成功”语义
  - 部分成功时，`content.data` 内会带 `message_cleanup_failed=true` 和 `message_cleanup_error`

### 7.10 调用示例

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/conversations?user_id=user-1&platform=cloud_public&page=1&limit=20" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/conversation?user_id=user-1&platform=cloud_public&conversation_id=conv-001" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/message?user_id=user-1&platform=cloud_public&conversation_id=conv-001&message_id=msg-003" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/messages?user_id=user-1&platform=cloud_public&conversation_id=conv-001&page=1&limit=20" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X DELETE "http://localhost:8080/api/v1/ai_robot/conversations?user_id=user-1&platform=cloud_public&start_time=2026-04-01&end_time=2026-04-07" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

---

## 8. 当前确认结果（本次检查）

本次检查后的结论：

- 路由指向 `Controllers.NewAiRobotController()` 正常
- Controller 基本保持薄入口职责
- 公网云平台的项目/任务/项目文章能力链路完整
- 内网云平台当前仅保留独立入口，聊天能力未正式开放，这是符合设计预期的，不需要与公网耦合
- 已修正一个对外误导点：避免把 `cloud_intranet` 展示成“已开放聊天能力”
- 本次主要做静态检查与注释/文档修正，未补跑完整自动化测试

---

## 9. 扩展指南

### 9.1 新增平台

1. 在 `AiRobotRequests.go` 增加平台常量与能力声明
2. 在 `internal/backends/registry.go` 注册平台 Backend
3. 在 `AiRobotRegistry.go` 增加平台分流
4. 在 `platform/<平台>/` 下实现独立分发与业务逻辑
5. 根据是否正式开放，决定是否写入 `exposed_chat_capability_matrix`
6. 在配置能力矩阵（`ai_gateway` 配置）中加入新平台意图白名单

### 9.2 新增 question_type

1. 在 `AiRobotRequests.go` 增加常量
2. 在对应平台的 `Dispatch...` 中加入分支
3. 在 `platform/<平台>/<领域>/` 新增对应处理逻辑
4. 同步更新能力矩阵和 `Capabilities` 展示

### 9.3 新增意图

1. 在 `ai_gateway` 的意图识别逻辑中加入新意图
2. 在对应平台领域的 `dispatch.go` 中加入处理分支
3. 同步更新能力矩阵白名单

---

## 10. 排障建议

### 10.1 参数错误

- 常见表现：`code=0`，`showMsg=请求参数错误`
- 检查：
  - `platform` 是否合法
  - `question_type` 是否被该平台支持
  - Header 是否带 `Token`

### 10.2 配置未初始化

- 常见表现：`showMsg=配置未初始化`
- 检查启动配置加载流程和环境变量

### 10.3 云平台调用失败

- 常见表现：`showMsg=云平台返回 HTTP xxx` 或 `请求云平台失败`
- 检查：
  - 平台 base URL 配置
  - token 权限
  - 目标接口可用性

---

## 11. 给后续维护者的建议

- Controller 保持薄：只做校验与入口
- 平台业务尽量放 `platform/<平台>/`，不要把内网/公网逻辑揉在一起
- 涉及平台差异时，优先在 registry 层做显式分流，避免混用
- 增加能力前，先同步更新能力矩阵和 `Capabilities` 展示，避免“配置可见但实际不可用”
- 会话相关能力新增时，优先考虑是否需要绑定 `user_id` 和管理员权限边界
