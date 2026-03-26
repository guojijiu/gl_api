# AI Robot 功能说明文档

本文档面向对 Go 不熟悉的同学，说明 `ai_robot` 从路由入口到业务执行的完整链路、当前能力边界、扩展方式和排障方法。

---

## 1. 路由入口

- 路由文件：`app/Http/Routes/routes.go`
- 路由分组：`/api/v1/ai_robot`
- 已开放接口：
  - `POST /api/v1/ai_robot/chat`：主问答入口
  - `GET /api/v1/ai_robot/capabilities`：查看当前能力矩阵与关键配置

---

## 2. 分层职责（当前结构）

### Controller 层（薄控制层）

- 文件：`app/Http/Controllers/AiRobotController.go`
- 职责：
  - 参数校验（JSON + 业务校验）
  - 读取 Header `Token`
  - 读取配置
  - 设置超时上下文
  - 调用服务层编排：`flow.Processor.ProcessChat(...)`
  - 对 `capabilities` 接口做直接输出

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
4. `Config.GetAiGatewayConfig()` 不为空

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

## 7. 当前确认结果（本次检查）

从路由到服务链路已完整打通，未发现“迁移未完成”问题：

- 路由指向 `Controllers.NewAiRobotController()` 正常
- Controller 已只保留入口职责
- 平台/业务/意图处理均在 `Services/ai_robot/flow`
- 能力策略和工具方法已分离到 `config` / `utils`
- 相关编译与测试通过

---

## 8. 扩展指南

### 8.1 新增平台

1. 在 `AiRobotRequests.go` 增加平台常量与能力声明
2. 在 `AiRobotRegistry.go` 的 `platformRegistry()` 增加平台处理函数
3. 在配置能力矩阵（`ai_gateway` 配置）中加入新平台意图白名单

### 8.2 新增 question_type

1. 在 `AiRobotRequests.go` 增加常量
2. 在 `AiRobotRegistry.go` 的业务注册表加入分支
3. 新增对应处理函数（建议在 `flow` 中独立文件）

### 8.3 新增意图

1. 在 `ai_gateway` 的意图识别逻辑中加入新意图
2. 在对应域的 `*_IntentRegistry()` 注册处理器
3. 同步更新能力矩阵白名单

---

## 9. 排障建议

### 9.1 参数错误

- 常见表现：`code=0`，`showMsg=请求参数错误`
- 检查：
  - `platform` 是否合法
  - `question_type` 是否被该平台支持
  - Header 是否带 `Token`

### 9.2 配置未初始化

- 常见表现：`showMsg=配置未初始化`
- 检查启动配置加载流程和环境变量

### 9.3 云平台调用失败

- 常见表现：`showMsg=云平台返回 HTTP xxx` 或 `请求云平台失败`
- 检查：
  - 平台 base URL 配置
  - token 权限
  - 目标接口可用性

---

## 10. 给后续维护者的建议

- Controller 保持薄：只做校验与入口
- 新业务尽量放 `flow`，工具和策略分别放 `utils`、`config`
- 涉及平台差异时，优先在 registry 层做显式分流，避免混用
- 增加能力前，先同步更新能力矩阵，避免“能识别但不允许执行”

