# Agent 开放平台 实现计划（面向 Claude Code 执行）

版本：v0.1 draft　　日期：2026-07-01
配套文档：《Agent开放平台_产品规格文档.md》《Agent开放平台_技术规格文档.md》

---

## 一、给 Claude Code 的阅读顺序

1. 先读 `docs/Agent开放平台_产品规格文档.md`、`docs/Agent开放平台_技术规格文档.md`，理解整体设计
2. 再读 `code/backend/README.md`，理解为什么 orchestration-service / marketplace-service 用 DDD 四层，
   gateway-service / billing-service 用简单三层
3. 按本文件的任务顺序执行，每个任务都标了目标文件、依赖关系、验收标准
4. 任何标注 `【需人工决策】` 的任务，先向人确认再动手，不要自行假设产品/架构决策

## 二、环境准备（M0，所有任务的前置条件）

- [ ] Go 1.21+、Node.js 18+、Docker
- [ ] `docker compose -f infra/docker-compose.yml up -d` 起 MySQL / Redis / RabbitMQ
- [ ] 执行 `infra/schema.sql` 建表
- [ ] `shared/proto/orchestration.proto` 跑 `protoc` 生成 Go gRPC 代码，产出放到 `shared/pkg/orchestration/`
      （当前 `orchestration-service/cmd/main.go` 和 `interfaces/grpc_server.go` 里的 gRPC 注册代码是占位，
      生成代码后需要把 `RegisterOrchestrationServiceServer(...)` 这行 TODO 替换成真实调用）

**验收标准**：四个服务能各自 `go run ./cmd` 启动不报错（此时业务逻辑仍是占位返回，属于正常现象）。

---

## 三、M1（MVP）：能力装配 + Agentic Loop + 双通道消费

对应产品规格文档路线图阶段一。目标：跑通"市场发布内置能力 → 装配 Agent → API/Workbench 都能对话"的完整链路。

### 3.1 marketplace-service

| 任务 | 目标文件 | 依赖 | 验收标准 |
|------|---------|------|---------|
| MySQL 仓储替换内存实现 | 新增 `internal/infrastructure/capability_repository_mysql.go`，实现 `domain.CapabilityRepository` | M0 建表完成 | 单元测试覆盖 Save/FindByID/ListPublished，`cmd/main.go` 里替换 `NewInMemoryCapabilityRepository` |
| 内置能力静态注册 | `internal/infrastructure/builtin_capabilities.go`（新建） | 无 | 至少注册 2 个内置 Tool（如网页检索、日历解析），跳过 MCP 校验（`isBuiltin=true`），启动时预写入数据库 |
| HTTP handler 补全 | `internal/interfaces/http_handler.go` | 仓储完成 | 补上 `GET /capabilities/{id}`、`POST /capabilities/{id}/approve`，覆盖产品规格文档 §5.1 发布审核流程 |

### 3.2 orchestration-service

| 任务 | 目标文件 | 依赖 | 验收标准 |
|------|---------|------|---------|
| 接入真实 ChatModel | `internal/infrastructure/eino_runtime.go` | 无 | 用 Eino 的 `ChatModel` 组件封装外部 LLM（Claude API / 火山方舟），跑通一次最简单的单轮对话 |
| 构建 Hub Agent | 同上，`buildToolsConfig` 函数 | ChatModel 接入完成 | 用 `adk.NewChatModelAgent` 构建，`ToolsConfig` 里注册 `agent.Capabilities()` 中 `IsSubagent=false` 的能力；`MaxIterations` 从 `agent.LoopTemplate()` 映射 |
| Hook 绑定 | 同上 + `internal/infrastructure/hooks_registry.go` | 上一步完成 | 用 Eino Callback 的 `OnStart`/`OnEnd` 挂载 `agent.Hooks()` 里启用的 Hook，`hooks_registry.go` 里的 TODO 逐个补齐 |
| MySQL AgentRepository | 新增 `internal/infrastructure/agent_repository_mysql.go` | M0 | 实现 `domain.AgentRepository`，`cmd/main.go` 里替换内存实现 |
| gRPC server 注册 | `cmd/main.go`、`internal/interfaces/grpc_server.go` | M0 proto 生成完成 | 真正调用 `RegisterOrchestrationServiceServer`，跑通 gateway → orchestration 的一次完整调用 |

**`【需人工决策】`**：Eino ChatModel 具体接哪个模型服务（Claude API 还是火山方舟豆包），需要确认 API Key 获取渠道，这会影响 `eino_runtime.go` 里的具体初始化代码。

### 3.3 gateway-service

| 任务 | 目标文件 | 依赖 | 验收标准 |
|------|---------|------|---------|
| API Key 鉴权中间件 | `internal/handler/middleware.go` | `api_key` 表已建 | 解析 `Authorization: Bearer sk-xxx`，哈希后查表，校验 scope，未授权返回 401（技术规格文档 §6.4） |
| Redis 令牌桶限流 | 同上 | Redis 起好 | Lua 脚本实现，按 apikey/agent/tenant/token 四个维度，超限返回 429 + `X-RateLimit-*` Header（技术规格文档 §6.3） |
| 转发到 orchestration-service | `internal/handler/agent.go` | orchestration gRPC 可用 | `SendMsg`/`StreamMsg` 真正建立 gRPC 调用并把结果转成 HTTP/SSE 响应 |
| OpenAPI 自动生成 | 同上 `OpenAPISpec` | 无 | 按产品规格文档 §6.3 的骨架 yaml，用 Agent 的能力 schema 动态填充 |

### 3.4 billing-service

| 任务 | 目标文件 | 依赖 | 验收标准 |
|------|---------|------|---------|
| 消息队列 consumer | `cmd/main.go` | RabbitMQ 起好 | 消费用量事件，写入 `usage_record` 表 |
| 用量上报对接 | orchestration-service `internal/infrastructure/usage_reporter_async.go` | 上一步完成 | 每次调用结束后异步发消息，字段覆盖 AgentID/TraceID/Depth/Tokens |

### 3.5 frontend

| 任务 | 目标文件 | 依赖 | 验收标准 |
|------|---------|------|---------|
| 接入 TanStack Query | `src/api/client.ts`、四个 `pages/*.tsx` | 无 | 按技术规格文档 §9.2 的状态分层方案，服务端数据（能力列表/Agent配置/用量）改用 `useQuery`/`useMutation`，替换裸 `fetch` |
| 市场页接入真实数据 | `src/pages/MarketplacePage.tsx`、`src/api/client.ts` | marketplace-service 可用 | 替换 `MOCK_CAPABILITIES`，调用 `api.listCapabilities()` |
| 构建器页提交发布 | `src/pages/AgentBuilderPage.tsx` | orchestration-service 配置接口可用 | "发布 Agent" 按钮调用真实接口，而不是 `disabled` 占位 |
| Workbench 接入 SSE | `src/pages/WorkbenchPage.tsx`、`src/api/client.ts` | gateway streamMsg 可用 | 用 `EventSource` 消费流式响应，替换掉现在的本地 state 假回复，参考技术规格文档 §9.3 |

**M1 整体验收标准**：在 Workbench 里对一个只挂载内置 Tool 的 Agent 发一条消息，能收到真实 LLM 生成的流式回复；同一个 Agent 通过 `curl` 调用开放 API 也能拿到一致的结果。

---

## 四、M2：Subagent-as-Tool、MCP 第三方接入、计费分成

对应产品规格文档路线图阶段二，建议在 M1 验收通过后再启动。

| 任务 | 目标文件 | 验收标准 |
|------|---------|---------|
| Subagent 包装为 Tool | orchestration `eino_runtime.go` 的 `buildToolsConfig` | `MountedCapability.IsSubagent=true` 时包装成 `BaseTool`，内部递归调用本服务，`Depth+1` 后经过 `agent.ValidateInvokeDepth` 校验 |
| 上下文精简传递 | 同上 | Subagent 只收到精简任务描述，不透传 Hub 完整对话历史（技术规格文档 §6.1） |
| MCP 健康检查 | marketplace `internal/infrastructure/mcp_registry_adapter.go` | `Register` 时对 endpoint 做连通性检查，失败拒绝注册 |
| 第三方能力提交流程 | marketplace `application/publish_usecase.go` + 前端 | 非内置能力必须提供 MCP endpoint 才能 `SubmitForReview`（规则已在聚合根里，补前端表单和审核后台） |
| 分成结算 | billing-service | 按 `usage_record.capability_id` 聚合，定时任务生成结算单 |
| 订阅关系 | marketplace 新增 `Subscription` 相关领域逻辑 | Hub 挂载他人能力前需要建立 `subscription` 记录（`【需人工决策】`：是否强制要求订阅，见产品规格文档 §5.2） |

---

## 五、M3：确定性 Pipeline、多租户隔离、容量扩展

远期规划，暂不展开详细任务，等 M1/M2 上线并有真实企业客户需求后再细化。方向参考技术规格文档第九章演进路径。

---

## 六、依赖关系图（先后顺序）

```
M0 建表 + proto 生成
  │
  ├─→ marketplace-service MySQL 仓储 ─→ 内置能力注册 ─→ HTTP handler 补全
  │                                                          │
  ├─→ orchestration-service ChatModel 接入 ─→ Hub Agent 构建 ─→ Hook 绑定 ─→ MySQL 仓储 ─→ gRPC 注册
  │                                                                                          │
  ├─→ gateway-service 鉴权中间件 ─→ 限流中间件 ─→ 转发 orchestration（依赖上面 gRPC 注册完成）
  │                                                                                          │
  └─→ frontend 接入三个后端服务（依赖以上全部完成）───────────────────────────────────────────┘
```

billing-service 的用量消费链路可以和上面并行开发，只在"验收 M1 整体链路"时需要接上。

---

## 七、给 Claude Code 的执行建议

- 每个任务对应一个独立的改动范围，建议一次只做一个任务，改完跑对应服务的 `go build ./...` 或前端的 `npm run build` 再继续下一个。
- domain 层的聚合根方法（`Agent`、`Capability`）已经包含核心业务规则，新增功能优先看能不能在聚合根里加方法，而不是在 application/infrastructure 层散写校验逻辑。
- 涉及外部技术细节（Eino ADK 具体 API、MCP 协议字段）如果记忆库里的用法和实际安装的包版本对不上，以实际 `go doc` 查到的签名为准，不要凭记忆硬编。
- 遇到本文件里标 `【需人工决策】` 的地方，先在对话里向人提出，不要自行选择一种方案往下写。
- 每完成一个 M1 任务，对照第三章表格里的"验收标准"自查一遍再继续。

---

## 八、待人工决策事项汇总（开发前建议先拍板）

以下事项在产品规格文档、技术规格文档中已标注为 `【待确认】`，汇总在这里方便一次性决策：

- Marketplace 是否对外部第三方开发者完全开放发布权限
- Subagent 挂载是否需要显式订阅/授权（影响 M2 订阅关系任务的实现方式）
- 被挂载能力版本更新后的默认跟随策略（锁定 vs 自动跟随）
- 是否支持对外暴露 Agent 内部单个 Skill，而非只暴露整体对话接口
- 是否所有第三方能力强制走 MCP 协议接入，还是允许平台内置能力走进程内调用（当前骨架代码已按"内置免 MCP、第三方强制 MCP"实现，如有异议需在 M2 前调整）
- 沙箱隔离粒度：按会话隔离还是按租户隔离
- Eino ChatModel 具体对接哪个 LLM 服务
