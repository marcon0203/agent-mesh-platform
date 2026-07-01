# Agent 开放平台产品规格文档（Spec）

版本：v0.1 draft　　日期：2026-07-01

---

## 一、产品概览

### 1.1 产品定位

一句话定义：一个 **Agent Marketplace / 开放平台**，让用户把 Tool、Skill 拆到最小粒度后自由挑选组装，快速构建出可用的 Agent，并支持通过开放 API 对外集成，或在 Workbench 内直接对话使用。

### 1.2 目标用户

- **能力提供方（开发者）**：发布 Tool / Skill / Agent 到市场，赚取调用分成。
- **组装方（构建者）**：从市场挑选能力，组装出满足自己业务需求的 Agent，不需要从零写 Prompt 和工具集成。
- **消费方（集成方 / 终端用户）**：通过开放 API 把已发布的 Agent 集成进自己的业务系统，或直接在 Workbench 对话使用。

### 1.3 核心挑战

- 挑战 1：能力粒度拆得太细，容易导致组装时选择爆炸，用户不知道该选哪些 Tool/Skill。
- 挑战 2：多 Agent 协同（Subagent 调度）如果做成独立的复杂编排系统，会大幅推高构建门槛，偏离"低门槛组装"的产品初衷。
- 挑战 3：作为开放平台，需要同时保证 API 集成通道和 Workbench 对话通道行为一致，且用量、计费、权限要跨通道统一。
- 挑战 4：市场生态问题——能力发布方的收益分成、版本兼容、审核机制，直接决定生态是否能持续运转。

### 1.4 设计原则

- 默认体验优先"配置"而非"编排"：创建 Agent 应该像勾选清单一样简单，执行顺序交给 Agentic Loop 自主决定，而不是让用户手工画流程图。
- 多 Agent 协同复用现有机制：Subagent 优先包装成标准 Tool 接口调用，避免引入第二套编排系统；只有强顺序/强合规场景才启用独立的确定性 Pipeline。
- 双通道（API / Workbench）必须共享同一套执行引擎，禁止出现"同一 Agent 在两个入口表现不一致"。
- 能力（Tool / Skill / Agent）三者都是市场里可独立发布、版本化、计费的资产。

### 1.5 假设与约束

- 【假设】初期目标场景以企业内部业务系统集成和中小开发者自建 Agent 为主，尚不考虑超大规模并发（QPS 千级以下）。
- 【假设】底层依赖某个 LLM 服务（如 Claude API）提供推理能力，平台自身不训练模型。
- 【待确认】Marketplace 是否对外部第三方开发者完全开放发布权限，还是初期仅平台方/白名单机构可发布。

---

## 二、整体架构

平台按五层设计，参考现有 Agent Harness 分层架构（接入层 / Agent 引擎 / 中间件 / 工具系统 / 基础设施），并在此基础上增加"能力市场"和"双通道消费"两个开放平台特有的结构。

```
┌───────────────────────────────────────────────────────────┐
│                     能力市场层 Marketplace                   │
│     Tools（原子工具）　　Skills（复合技能）　　Agent（可挂载子能力）│
│              ↓ 动态发现 / 注册中心（ToolSearchIndex + MCP）    │
└───────────────────────────┬───────────────────────────────┘
                             │
┌───────────────────────────▼───────────────────────────────┐
│                  编排引擎层 Orchestration Engine              │
│   能力装配（勾选配置） → Agentic Loop（LLM自主） → 多Agent协同  │
└───────────────────────────┬───────────────────────────────┘
                             │
┌───────────────────────────▼───────────────────────────────┐
│               中间件 Hooks 层（可插拔，前处理/后处理）          │
└───────────────────────────┬───────────────────────────────┘
                             │
┌───────────────────────────▼───────────────────────────────┐
│                    已发布 Agent 实例                          │
└─────────────┬─────────────────────────────┬─────────────────┘
              │                             │
     ┌────────▼────────┐          ┌─────────▼─────────┐
     │   开放 API 通道    │          │  Workbench 对话通道  │
     │ sendMsg/streamMsg │          │  调试模式/会话历史   │
     └────────┬────────┘          └─────────┬─────────┘
              │                             │
┌─────────────▼─────────────────────────────▼─────────────────┐
│                     基础设施层 Infrastructure                  │
│  记忆/会话  权限/沙箱  计费/用量  Checkpoint  可观测性  VFS/文件  │
└───────────────────────────────────────────────────────────┘
```

各层职责说明：

| 层级 | 组件 | 核心职责 |
|------|------|---------|
| 能力市场层 | Tools / Skills / Agent-as-capability | 原子能力的发布、发现、版本管理 |
| 编排引擎层 | Agentic Loop / Hub 调度 / Pipeline 配置 | 决定 Agent 运行时的执行逻辑 |
| 中间件层 | 前处理 / 后处理 Hooks | 会话初始化、记忆加载、权限校验、消费上报等横切逻辑 |
| 消费通道层 | 开放 API / Workbench | 对外暴露已发布 Agent 的两种使用形态 |
| 基础设施层 | 记忆、权限、计费、Checkpoint、可观测性 | 支撑全链路，跨通道统一 |

---

## 三、模块划分与职责

```
[能力市场 marketplace-service]
  对外接口：发布 Tool/Skill/Agent、检索能力、能力详情、版本列表
  依赖：注册中心（ToolSearchIndex）、审核服务
  关键设计：Tool/Skill/Agent 三种资产统一走发布-审核-上架流程，Agent 可作为一种特殊"能力"被其他 Agent 挂载
  数据所有权：capability、capability_version、publisher

[编排引擎 orchestration-engine]
  对外接口：创建/更新 Agent 配置、运行时执行（Agentic Loop）、Subagent 调度
  依赖：能力市场（拉取已挂载能力）、中间件 Hooks、LLM 推理服务
  关键设计：默认走 Agentic Loop（LLM 自主决定调用顺序），Subagent 包装成标准 Tool 接口；仅在需要确定性流程时启用 Pipeline 配置（DAG）
  数据所有权：agent_config、agent_run、run_trace

[中间件 Hooks middleware-hooks]
  对外接口：Hook 注册、Hook 启用/禁用配置
  依赖：无（被编排引擎在 Loop 各阶段调用）
  关键设计：前处理（会话初始化、记忆加载、上下文组装、提示词构建、工具发现、权限校验）与后处理（会话清理、消费上报、沙箱清理、状态回传）两类，均可插拔
  数据所有权：hook_definition、hook_binding

[消费通道 channel-gateway]
  对外接口：开放 API（sendMsg/streamMsg）、Workbench 对话接口
  依赖：编排引擎（同一执行入口）、鉴权服务
  关键设计：两条通道共享同一套执行引擎和能力配置，仅入口鉴权方式和会话管理方式不同；每个发布的 Agent 自动生成 API 文档
  数据所有权：api_key、workbench_session

[基础设施 infra]
  对外接口：记忆读写、权限校验、计费上报、Checkpoint 保存/恢复、Trace 查询
  依赖：无（被所有上层模块调用）
  关键设计：计费/用量按 Agent 维度统一，不区分调用来自 API 还是 Workbench；调用链路 Trace 支持多层 Subagent 嵌套追踪
  数据所有权：memory、permission、usage_record、checkpoint、trace
```

模块依赖关系图（文字版）：

```
marketplace-service ←── orchestration-engine ──→ middleware-hooks
                              │
                              ▼
                      channel-gateway
                              │
                              ▼
                            infra
```

---

## 四、核心机制方案

### 4.1 三层执行模型（避免"过度编排"）

创建 Agent 时不要求用户理解"编排"，而是分三层递进，绝大多数用户只需要用到第一层：

| 层级 | 用户操作 | 运行时行为 | 适用场景 |
|------|---------|-----------|---------|
| 能力装配 | 勾选 Tool/Skill/Hook | 静态绑定，无执行顺序 | 所有 Agent 创建的基础步骤 |
| Agentic Loop | 无需额外操作 | LLM 自主决定调用顺序，循环执行直到完成 | 默认运行时机制，覆盖绝大多数场景 |
| 多 Agent 协同 | 挂载其他 Agent 作为子能力（可选） | Hub 的 LLM 动态决定是否调用 Subagent | 任务需要拆分给多个专业 Agent 时 |

### 4.2 多 Agent 协同：Subagent-as-Tool（默认机制）

- Subagent 不需要单独的编排系统，直接包装成标准 Tool 接口，供 Hub Agent 的 `tool_call` 环节调用，与调用普通工具走同一套机制。
- 上下文传递采用"精简任务描述 + 必要上下文片段"，禁止把 Hub 完整对话历史转发给 Subagent；Subagent 执行完毕只回传结果摘要，不暴露内部执行细节，防止上下文膨胀。
- 递归调用需设最大深度限制（建议默认 3 层），超过强制拦截，防止无限递归导致 Token 消耗失控。

### 4.3 确定性 Pipeline（进阶机制，可选）

- 仅在客户明确要求强顺序、强分支、需人工审批介入等确定性场景下启用。
- 本质是一个轻量 DAG 配置：节点为 Agent 或 Tool，边表示执行顺序与条件分支。
- 建议作为独立的进阶功能模块后置开发，不影响 MVP 阶段的默认体验。

### 4.4 能力检索与"选择爆炸"应对

- 市场能力量级上升后，创建 Agent 时的能力选择页面需支持：分类目录浏览、语义搜索（按任务描述检索匹配的 Tool/Skill）、智能装配推荐（根据用户输入的 Agent 用途，自动推荐一组能力组合供用户勾选确认，而非从零挑选）。

---

## 五、Marketplace 特有机制

### 5.1 发布与审核

```
能力提交（Tool/Skill/Agent）→ 自动化校验（schema 合法性、安全扫描）
  → 人工/规则审核 → 上架 → 版本迭代（沿用同一发布流程）
```

### 5.2 权限与订阅

- 【待确认】Hub Agent 是否可无限制挂载市场任意 Agent 作为 Subagent，还是需要显式订阅/授权后才可调用。建议默认要求订阅关系，避免能力提供方内容被无授权大量调用。

### 5.3 计费与分成

- 计费按 Agent 维度统一核算，覆盖 API 调用和 Workbench 对话两个入口。
- 涉及调用他人发布的 Tool/Skill/Agent 时，需设计分成机制（按调用次数或 Token 消耗比例分成），这是能力提供方持续发布优质内容的核心激励。

### 5.4 版本管理

- 被挂载的能力（Tool/Skill/Agent）版本更新后，默认策略【待确认】：自动跟随最新版本，还是锁定挂载时版本、由用户手工升级。建议默认锁定版本 + 提供"有新版本可用"提醒，避免 Hub Agent 行为因依赖方更新而意外突变。

---

## 六、API 开放规范

### 6.1 鉴权（Authentication）

开放平台面向的是"第三方系统集成"，鉴权设计要和企业内部 SSO 区分开：

| 场景 | 鉴权方式 | 说明 |
|------|---------|------|
| 开放 API（第三方集成） | API Key | 开发者在控制台为每个应用生成 Key，Header 携带 `Authorization: Bearer sk-xxx` |
| 开放 API（企业级/需用户身份透传） | OAuth2（Client Credentials / Authorization Code） | 需要代表终端用户调用、或要做精细审计的场景 |
| Workbench 对话 | 平台账号会话（Session Cookie / SSO） | 走平台自身登录态，不暴露 API Key |
| Agent 间调用（Subagent-as-Tool） | 内部服务凭证（不经过公网） | Hub 调 Subagent 走内部信任链，不复用外部 API Key 体系 |

**API Key 设计要点：**
- 一个开发者账号下可创建多个 API Key，分别绑定不同的可访问 Agent 范围（scope），避免一把 Key 拥有全部权限。
- Key 支持吊销、轮换（rotate），旧 Key 设置宽限期后失效，避免线上系统因轮换中断。
- Key 只在创建时完整展示一次，之后仅展示前后若干位，降低泄露风险。

**Scope 设计：**
```
scope 示例：
  agent:{agent_id}:invoke     — 允许调用指定 Agent
  agent:{agent_id}:read       — 允许查询该 Agent 的运行记录/用量
  marketplace:capability:read — 允许检索市场能力（不含调用）
```

### 6.2 限流（Rate Limiting）

限流需要分层设计，防止单一租户打满平台资源，也防止单个 Agent 被滥用：

| 限流维度 | 说明 | 默认策略（示意） |
|---------|------|-----------------|
| API Key 级 | 单个 Key 的总调用频率 | 令牌桶，60 次/分钟（可按套餐调整） |
| Agent 级 | 单个 Agent 承受的总并发/QPS，保护发布方不被下游打爆 | Agent 发布时可自主设置上限 |
| 租户/账号级 | 防止单租户占满平台共享资源 | 按套餐分级（免费/付费/企业） |
| Token 消耗级 | 按 Token 消耗限流，而非仅按请求数（Agent 场景下单次调用成本差异很大） | 按分钟/小时/天设置 Token 配额 |
| Subagent 递归级 | 防止一次外部请求触发深层递归调用打满下游 | 结合第四章的最大递归深度限制一并生效 |

**限流响应规范：**
```
HTTP 429 Too Many Requests
Headers:
  X-RateLimit-Limit: 60
  X-RateLimit-Remaining: 0
  X-RateLimit-Reset: 1719820800
  Retry-After: 12

Body:
{
  "code": 40002,
  "message": "超出调用配额，请 12 秒后重试"
}
```

限流算法建议采用令牌桶（Token Bucket），允许一定程度的突发流量，同时长期速率受限；Token 消耗类限流建议采用滑动窗口，避免固定窗口边界处的突刺问题。

### 6.3 OpenAPI 规范与自动生成文档

每个 Agent 发布后，平台根据其声明的输入输出 schema **自动生成一份 OpenAPI 3.0 文档**，开发者无需手写。核心设计：

- Agent 的输入输出 schema 在创建/编辑阶段就以 JSON Schema 形式定义（对应能力市场里 Tool/Skill 的 schema 规范，Agent 本身也遵循同一套 schema 体系）。
- 平台服务根据该 schema 渲染出 `/api/v1/agents/{agent_id}/openapi.json`，供开发者直接导入 Postman / 生成 SDK。
- 所有 Agent 共用统一的基础 OpenAPI 骨架（鉴权方式、错误码结构、通用 Header），仅 `sendMsg` 的 request/response body 按各 Agent 的 schema 动态填充。

**OpenAPI 骨架示意（节选）：**
```yaml
openapi: 3.0.3
info:
  title: "{agent_name} API"
  version: "{agent_version}"
servers:
  - url: https://api.platform.com/v1
security:
  - ApiKeyAuth: []
paths:
  /agents/{agent_id}/sendMsg:
    post:
      summary: 向该 Agent 发送一条消息（同步）
      parameters:
        - name: agent_id
          in: path
          required: true
          schema: { type: string }
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                message: { type: string, description: "用户输入" }
                session_id: { type: string }
                context: { type: object }
              required: [message]
      responses:
        "200":
          description: 成功
          content:
            application/json:
              schema:
                type: object
                properties:
                  code: { type: integer }
                  data:
                    type: object
                    properties:
                      reply: { type: string }
                      trace_id: { type: string }
                      usage:
                        type: object
                        properties:
                          tokens: { type: integer }
        "429":
          description: 超出限流配额
        "401":
          description: 鉴权失败
  /agents/{agent_id}/streamMsg:
    post:
      summary: 向该 Agent 发送一条消息（SSE 流式）
      responses:
        "200":
          description: text/event-stream 流式返回，chunk 结构同 sendMsg 的 data 字段
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: Authorization
```

### 6.4 核心接口设计（API Contract，示意）

```
POST /api/v1/agents/{agent_id}/sendMsg
描述：向已发布 Agent 发送一条消息（同步）
Headers: Authorization: Bearer sk-xxx

Request:
{
  "message": "帮我总结这份报告",     // 用户输入
  "session_id": "xxx",             // 会话标识，可选
  "context": {}                    // 额外业务上下文，可选
}

Response（成功 200）:
{
  "code": 0,
  "data": {
    "reply": "...",
    "trace_id": "xxx",             // 用于查询调用链路
    "usage": { "tokens": 1234 }
  }
}

POST /api/v1/agents/{agent_id}/streamMsg
描述：流式返回（SSE），字段同上，按 chunk 推送

错误码：
40001 - Agent 未发布或已下线    40002 - 超出调用配额
40003 - 能力挂载版本不兼容      40004 - 递归深度超限
40101 - API Key 无效或已吊销   40102 - Key 无该 Agent 的调用 scope
```

Workbench 通道复用同一 `agent_id` 和执行引擎，仅将 `session_id` 的会话管理放在前端界面维护，并额外开启调试模式，暴露每一步 Hook 触发情况、Tool/Subagent 调用详情。Workbench 走平台登录态，不需要 API Key，因此不计入 API Key 维度的限流，但仍计入 Agent 级和账号级限流与用量统计。

---

## 七、核心数据模型（简要）

```
Capability（能力，Tool/Skill 通用）
  id, type(tool|skill), name, schema(input/output), version, publisher_id, status

Agent
  id, name, config(挂载的 capability_id 列表 + hook 配置 + loop 模板), version, status

AgentRun
  id, agent_id, channel(api|workbench), trace_id, usage, started_at, status

Subscription
  id, agent_id(消费方), capability_id 或 agent_id(被挂载方), status

UsageRecord
  id, agent_id, channel, tokens, cost, revenue_share, created_at
```

---

## 八、非功能需求

| 需求 | 指标 | 方案 |
|------|------|------|
| 通道一致性 | API 与 Workbench 行为 100% 一致 | 共用同一执行引擎，禁止双实现 |
| 可观测性 | 全链路 Trace，含多层 Subagent | 按 trace_id 串联 Hub + Subagent 调用树 |
| 多租户隔离 | 沙箱级隔离 | 每次执行独立沙箱，防止跨租户数据泄漏 |
| 计费准确性 | 跨通道用量误差 < 1% | 用量统计下沉到基础设施层统一上报 |
| 递归安全 | 防止无限递归 | 默认最大调用深度 3 层，超限强制拦截 |

---

## 九、风险与演进路径

```
当前方案的局限性：
- 风险 1：能力粒度拆分标准不统一，可能导致市场内容质量参差 → 需要明确的发布规范和审核标准
- 风险 2：多 Agent 递归调用的成本不透明，用户可能在不知情下产生高额 Token 消耗 → 需要调用前的成本预估提示
- 风险 3：确定性 Pipeline 若过早投入，可能分散 MVP 阶段的资源 → 建议后置

演进路径：
阶段一（MVP）：能力市场（Tool/Skill 发布检索） + 能力装配式创建 Agent + Agentic Loop 运行时 + 双通道消费（API + Workbench）
阶段二：Subagent-as-Tool 多 Agent 协同、智能装配推荐、分成计费机制
阶段三：确定性 Pipeline 编排画布、企业级合规/审批流程支持
```

---

## 十、待确认事项清单

- Marketplace 是否对外部第三方开发者完全开放发布权限
- Subagent 挂载是否需要显式订阅/授权
- 被挂载能力版本更新后的默认跟随策略
- 是否支持对外暴露 Agent 内部单个 Skill（而非只暴露整体对话接口）
