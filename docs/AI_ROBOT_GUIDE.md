# AI Robot 功能说明文档

本文档面向对 Go 不熟悉的同学，说明 `ai_robot` 从路由入口到业务执行的完整链路、当前能力边界、扩展方式和排障方法。

---

## 1. 路由入口

- 路由文件：`app/Http/Routes/routes.go`
- 路由分组：`/api/v1/ai_robot`
- 已开放接口：
  - `POST /api/v1/ai_robot/chat`：主问答入口
  - `GET /api/v1/ai_robot/capabilities`：查看当前能力矩阵与关键配置
  - `GET /api/v1/ai_robot/conversations`：分页查看会话列表
  - `DELETE /api/v1/ai_robot/conversations`：按条件批量删除会话
  - `GET /api/v1/ai_robot/conversation`：查看单条会话详情
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
  - `flow/`：主流程编排、平台分流、项目域/任务域处理
  - `config/`：能力策略判断（例如意图是否允许）
  - `utils/`：工具函数（如 access_code 生成）
  - `handler/`：通用路由器组件（保留扩展用）

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

这里会创建一个 `aiRobotDeps`（请求级依赖容器），并进入 `dispatchByRegistry(...)`。

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

### 步骤 4：意图分流与执行

- 项目域：`AiRobotProjectHandler.go`
- 任务域：`AiRobotTaskHandler.go`
- 图片域：`AiRobotImageHandler.go`（当前占位）

典型流程是：

1. 大模型意图判定（`PlanProjectIntent` / `PlanTaskIntent`）
2. 判断能力矩阵是否允许该意图
3. 调云平台接口获取原始数据
4. 大模型做摘要/自然语言回答
5. 统一输出 front 格式

---

## 4. 懒初始化说明（已实现）

文件：`app/Services/ai_robot/flow/AiRobotTypes.go`

`aiRobotDeps` 提供了以下懒加载方法：

- `project()`
- `contract()`
- `task()`

含义：只有当分支真的需要时，才创建对应 client，避免每个请求都无条件创建 3 个 client。

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

---

## 6. 返回格式约定

`chat` 与 `capabilities` 都走统一前端格式：

- HTTP 状态固定 `200`
- 业务状态看 `code` 字段
  - `1`：成功
  - `0`：失败
- 关键字段：
  - `showMsg`
  - `debugMsg`
  - `content`

---

## 7. 会话管理接口

### 7.1 接口速查表

| 方法 | 路径 | 必填参数 | 说明 |
| ---- | ---- | ---- | ---- |
| `POST` | `/api/v1/ai_robot/chat` | `question` `platform` `question_type` `user_id`(上下文场景必填) | 主问答入口，Header 需带 `Token` |
| `GET` | `/api/v1/ai_robot/capabilities` | 无 | 查看当前能力矩阵和关键配置 |
| `GET` | `/api/v1/ai_robot/conversations` | `user_id` | 分页查看会话列表，支持多条件筛选 |
| `GET` | `/api/v1/ai_robot/conversation` | `user_id` `platform` `conversation_id` | 查看单条会话详情 |
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
- 用途：查看当前能力矩阵、模型和平台地址，便于联调/排障

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
  - 详情仍走单条接口

### 7.5 单条会话详情

- 接口：`GET /api/v1/ai_robot/conversation`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`

### 7.6 删除单条会话

- 接口：`DELETE /api/v1/ai_robot/conversation`
- 必填参数：
  - `user_id`
  - `platform`
  - `conversation_id`

### 7.7 批量删除会话

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

### 7.8 调用示例

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/conversations?user_id=user-1&platform=cloud_public&page=1&limit=20" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X GET "http://localhost:8080/api/v1/ai_robot/conversation?user_id=user-1&platform=cloud_public&conversation_id=conv-001" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

```bash
curl -X DELETE "http://localhost:8080/api/v1/ai_robot/conversations?user_id=user-1&platform=cloud_public&start_time=2026-04-01&end_time=2026-04-07" \
  -H "Token: YOUR_CLOUD_PLATFORM_TOKEN"
```

---

## 8. 当前确认结果（本次检查）

从路由到服务链路已完整打通，未发现“迁移未完成”问题：

- 路由指向 `Controllers.NewAiRobotController()` 正常
- Controller 已只保留入口职责
- 平台/业务/意图处理均在 `Services/ai_robot/flow`
- 能力策略和工具方法已分离到 `config` / `utils`
- 相关编译与测试通过

---

## 9. 扩展指南

### 9.1 新增平台

1. 在 `AiRobotRequests.go` 增加平台常量与能力声明
2. 在 `AiRobotRegistry.go` 的 `platformRegistry()` 增加平台处理函数
3. 在配置能力矩阵（`ai_gateway` 配置）中加入新平台意图白名单

### 9.2 新增 question_type

1. 在 `AiRobotRequests.go` 增加常量
2. 在 `AiRobotRegistry.go` 的业务注册表加入分支
3. 新增对应处理函数（建议在 `flow` 中独立文件）

### 9.3 新增意图

1. 在 `ai_gateway` 的意图识别逻辑中加入新意图
2. 在对应域的 `*_IntentRegistry()` 注册处理器
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
- 新业务尽量放 `flow`，工具和策略分别放 `utils`、`config`
- 涉及平台差异时，优先在 registry 层做显式分流，避免混用
- 增加能力前，先同步更新能力矩阵，避免“能识别但不允许执行”
- 会话相关能力新增时，优先考虑是否需要绑定 `user_id` 和管理员权限边界
