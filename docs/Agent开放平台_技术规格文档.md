# Agent 开放平台技术规格文档（Technical Spec）

版本：v0.1 draft　　日期：2026-07-01　　配套文档：《Agent开放平台_产品规格文档.md》

---

## 一、系统概览

【系统定位】
Agent 开放平台的技术实现，核心是"用统一执行引擎跑通能力市场→编排→双通道消费"的完整链路，技术栈以 Go 为主，编排引擎基于字节跳动开源的 Eino/Eino ADK 构建。

【核心挑战】
- 挑战 1：能力（Tool/Skill/Agent）需要支持动态发现和热插拔，不能每次新增能力都改代码重新发布 Agent。
- 挑战 2：多 Agent 递归调用（Subagent-as-Tool）的深度和成本必须可控，防止失控消耗。
- 挑战 3：开放 API 和 Workbench 两条通道必须共享同一套执行状态，用量、计费、Trace 要跨通道对得上。

【设计原则】
- 编排引擎与中间件 Hooks 复用 Eino/Eino ADK 现成能力，不重复造轮子；Eino 未覆盖的能力市场、计费、网关、多租户隔离自建。
- 所有对外可调用单元（Tool/Skill/Subagent）统一走标准化 Schema 接口，保证 Eino 组件体系和市场发布流程使用同一套契约。
- 优先保证工程可维护性（强类型、可观测、可中断恢复），而不是短期堆功能。

【假设与约束】
- 【假设】技术栈：Go 1.21+，编排引擎 Eino + Eino ADK，网关层 Hertz（CloudWeGo 同生态，与 Eino 配合成本低）。
- 【假设】初期规模：DAU 级 Agent 调用量 10 万级，QPS 峰值 ~500，暂不做超大规模分片设计。
- 【假设】LLM 推理能力通过外部模型服务（Claude API / 火山方舟等）接入，Eino ChatModel 组件做统一封装，平台自身不训练模型。
- 【待确认】是否需要同时支持 Python 生态的第三方 Tool（通过 MCP 协议接入可绕开语言限制，建议作为主要集成方式）。

---

## 二、整体架构（技术选型版）

```
┌───────────────────────────────────────────────────────────┐
│  客户端 / 集成方                                              │
│  第三方业务系统（HTTP/SSE）　　Workbench Web（浏览器）           │
└──────────────────────────┬──────────────────────────────────┘
                            │ HTTPS
┌──────────────────────────▼──────────────────────────────────┐
│  网关层 gateway-service（Hertz）                                │
│  职责：鉴权（API Key/OAuth2/Session）、限流、路由、SSE 转发、       │
│         OpenAPI 文档生成                                       │
└──────┬──────────────┬───────────────┬───────────────────────┘
       │              │               │
┌──────▼──────┐ ┌─────▼──────────┐ ┌──▼─────────────────┐
│marketplace- │ │orchestration-   │ │ billing-service     │
│service      │ │service          │ │（计费/分成/用量）      │
│（能力发布/   │ │（Eino + Eino ADK）│ └──────────┬──────────┘
│ 审核/发现）  │ │ Agentic Loop     │            │
└──────┬──────┘ │ Host Multi-Agent │            │
       │        │ Hook 切面注入     │            │
       │        └────────┬────────┘            │
       │                 │                      │
┌──────▼─────────────────▼──────────────────────▼──────────────┐
│  基础设施层                                                     │
│  MySQL（元数据） Redis（限流/会话缓存） 向量库（长期记忆）           │
│  对象存储（VFS）  消息队列（异步计费/审核）  沙箱运行时（隔离执行）    │
│  OpenTelemetry + Eino Callback 事件流（可观测性/Trace）           │
└───────────────────────────────────────────────────────────────┘
```

各层职责说明：

| 层级 | 组件 | 核心职责 | 技术选型 |
|------|------|---------|---------|
| 网关层 | gateway-service | 鉴权、限流、路由、SSE、文档生成 | Hertz |
| 编排引擎层 | orchestration-service | Agentic Loop、多 Agent 调度、Hook 注入 | Eino / Eino ADK |
| 能力市场层 | marketplace-service | 能力发布、审核、动态发现 | Go + MCP 协议 |
| 计费层 | billing-service | 用量统计、分成结算 | Go + 消息队列异步处理 |
| 数据层 | MySQL/Redis/向量库/对象存储 | 元数据、缓存、记忆、文件 | MySQL 8 / Redis 7 / Milvus / S3 兼容存储 |

---

## 三、模块划分与职责

```
[gateway-service]
  对外接口：/api/v1/agents/{id}/sendMsg、/streamMsg、/openapi.json
  依赖：orchestration-service（gRPC 内部调用）、鉴权/限流中间件
  关键设计：所有对外流量的唯一入口，Workbench 走同一网关但走 Session 鉴权分支
  数据所有权：api_key、rate_limit_config

[orchestration-service]
  对外接口：内部 gRPC，供 gateway 调用；不直接对外暴露
  依赖：marketplace-service（拉取能力定义）、billing-service（上报用量）、沙箱运行时
  关键设计：基于 Eino ADK 的 ChatModelAgent + Host Multi-Agent 模式实现 Agentic Loop 和 Subagent-as-Tool；
            用 Eino 的 Callback 机制实现前处理/后处理 Hooks；MaxIterations 控制单 Agent 循环上限，
            自定义 depth 字段在调用上下文中传递控制递归深度
  数据所有权：agent_config、agent_run、run_trace（写入，供 trace-service 查询）

[marketplace-service]
  对外接口：能力发布、检索、详情、版本列表（REST）
  依赖：审核服务（同步/异步校验）、MCP Server 注册中心
  关键设计：Tool/Skill/Agent 统一注册为 MCP Server 或标准 Schema 描述，
            orchestration-service 通过 MCP 协议动态发现和调用，不需要重新编译部署
  数据所有权：capability、capability_version、publisher

[billing-service]
  对外接口：内部用量上报接口、对外用量查询接口
  依赖：消息队列（异步消费用量事件）
  关键设计：用量事件异步写入，避免同步计费拖慢主链路；按 Agent 维度聚合，
            涉及分成的按调用链路（Hub → Subagent）拆分记账
  数据所有权：usage_record、revenue_share_record

[sandbox-runtime]
  对外接口：内部调用，供 orchestration-service 在执行工具/代码时调用
  依赖：容器运行时
  关键设计：每次会话级隔离，执行完销毁，防止跨租户数据泄漏
  数据所有权：无持久数据，仅运行时状态
```

模块依赖关系图（文字版）：

```
gateway-service → orchestration-service → marketplace-service（能力发现）
                        │                → sandbox-runtime（工具执行隔离）
                        │                → billing-service（异步用量上报）
                        └→ trace-service（Callback 事件流写入）
```

---

## 四、数据库设计

#### 核心表结构

```sql
-- 表名：capability（能力主表，Tool/Skill/Agent 统一注册）
CREATE TABLE capability (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    type          TINYINT         NOT NULL COMMENT '1=tool 2=skill 3=agent',
    name          VARCHAR(100)    NOT NULL,
    publisher_id  BIGINT UNSIGNED NOT NULL,
    schema_json   JSON            NOT NULL COMMENT '输入输出 JSON Schema',
    mcp_endpoint  VARCHAR(255)             COMMENT 'MCP Server 地址，动态发现用',
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1待审核 2已上架 3已下线',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_publisher (publisher_id),
    KEY idx_type_status (type, status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力主表';

-- 表名：capability_version（能力版本表）
CREATE TABLE capability_version (
    id             BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    capability_id  BIGINT UNSIGNED NOT NULL,
    version        VARCHAR(20)     NOT NULL COMMENT '语义化版本号',
    changelog      TEXT,
    is_latest      TINYINT         NOT NULL DEFAULT 0,
    created_at     DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_cap_version (capability_id, version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力版本表';

-- 表名：agent（已发布 Agent）
CREATE TABLE agent (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name          VARCHAR(100)    NOT NULL,
    owner_id      BIGINT UNSIGNED NOT NULL,
    config_json   JSON            NOT NULL COMMENT '挂载能力列表 + Hook 配置 + Loop 模板',
    version       VARCHAR(20)     NOT NULL,
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1开发中 2已发布 3已下线',
    max_depth     TINYINT         NOT NULL DEFAULT 3 COMMENT '子 Agent 最大递归深度',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_owner (owner_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent 主表';

-- 表名：agent_run（运行记录）
CREATE TABLE agent_run (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    agent_id      BIGINT UNSIGNED NOT NULL,
    channel       TINYINT         NOT NULL COMMENT '1=api 2=workbench',
    trace_id      VARCHAR(64)     NOT NULL,
    caller_key_id BIGINT UNSIGNED          COMMENT '发起调用的 API Key，Workbench 场景为空',
    tokens_used   INT UNSIGNED    NOT NULL DEFAULT 0,
    status        TINYINT         NOT NULL COMMENT '1成功 2失败 3限流拒绝',
    started_at    DATETIME        NOT NULL,
    PRIMARY KEY (id),
    KEY idx_agent_time (agent_id, started_at),
    KEY idx_trace (trace_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Agent 运行记录表';

-- 表名：api_key
CREATE TABLE api_key (
    id            BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    developer_id  BIGINT UNSIGNED NOT NULL,
    key_prefix    VARCHAR(12)     NOT NULL COMMENT '展示用前缀，如 sk-abc1',
    key_hash      VARCHAR(64)     NOT NULL COMMENT 'Key 的哈希值，不存明文',
    scopes        JSON            NOT NULL COMMENT '如 ["agent:123:invoke"]',
    status        TINYINT         NOT NULL DEFAULT 1 COMMENT '1启用 2已吊销',
    created_at    DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_key_hash (key_hash),
    KEY idx_developer (developer_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='开发者 API Key 表';

-- 表名：subscription（能力挂载订阅关系）
CREATE TABLE subscription (
    id                BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    consumer_agent_id BIGINT UNSIGNED NOT NULL COMMENT '发起挂载的 Agent（Hub）',
    capability_id     BIGINT UNSIGNED NOT NULL COMMENT '被挂载的能力（Tool/Skill/Agent）',
    pinned_version    VARCHAR(20)     NOT NULL COMMENT '锁定版本，默认不跟随最新',
    status            TINYINT         NOT NULL DEFAULT 1,
    created_at        DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_consumer_cap (consumer_agent_id, capability_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='能力挂载订阅表';

-- 表名：usage_record（用量与分成记录）
CREATE TABLE usage_record (
    id              BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    agent_run_id    BIGINT UNSIGNED NOT NULL,
    capability_id   BIGINT UNSIGNED          COMMENT '若该次消耗来自某个被调用能力，记录归属',
    tokens          INT UNSIGNED    NOT NULL DEFAULT 0,
    cost_cents      INT UNSIGNED    NOT NULL DEFAULT 0,
    revenue_share   INT UNSIGNED    NOT NULL DEFAULT 0 COMMENT '分给能力提供方的金额（分）',
    created_at      DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_run (agent_run_id),
    KEY idx_capability (capability_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用量与分成记录表';
```

#### 索引策略

| 表 | 索引 | 使用场景 |
|----|------|---------|
| agent_run | idx_agent_time | 查询某 Agent 一段时间内的运行记录，用于计费对账、监控看板 |
| capability | idx_type_status | 市场检索页按类型+上架状态过滤 |
| usage_record | idx_capability | 按能力维度汇总分成，用于给发布方结算 |

---

## 五、核心接口设计

对外 REST API（网关层，鉴权/限流/OpenAPI 生成细节见产品规格文档第六章，此处不重复）；内部服务间调用采用 gRPC，示例：

```protobuf
service OrchestrationService {
  rpc Invoke(InvokeRequest) returns (stream InvokeChunk);
}

message InvokeRequest {
  string agent_id = 1;
  string session_id = 2;
  string message = 3;
  int32  depth = 4;        // 当前递归深度，Subagent 调用时 +1 传递
  string trace_id = 5;
}

message InvokeChunk {
  string content = 1;
  bool   is_final = 2;
  UsageInfo usage = 3;
}
```

`depth` 字段是控制递归安全的关键：网关首次请求 `depth=0`，orchestration-service 每次以 Subagent-as-Tool 方式发起内部调用时 `depth+1`，超过 `agent.max_depth` 直接拒绝并返回 `40004` 错误码。

---

## 六、关键技术方案

### 6.1 基于 Eino ADK 的多 Agent 编排实现

- 使用 `adk.NewChatModelAgent` 构建 Hub Agent，`ToolsConfig` 中挂载普通 Tool 组件；Subagent 同样包装为 `BaseTool` 接口注入到 Hub 的工具列表中，Hub 的 LLM 在 `tool_call` 阶段无差别调用。
- `MaxIterations` 参数控制单个 Agent 自身循环轮数上限，防止在同一层陷入死循环；跨层递归深度用 6.5 节的 `depth` 字段在 gRPC 调用链中显式传递并校验，两者共同构成安全边界。
- Eino 的 Callback 机制承接前处理/后处理 Hook：在 `OnStart`（对应记忆加载、权限校验）和 `OnEnd`（对应消费上报、状态回传）阶段注册自定义回调，实现产品规格文档第三章描述的"可插拔 Hooks"。

### 6.2 能力动态发现（MCP 协议集成）

- 市场里的 Tool/Skill 统一要求发布方提供 MCP Server 地址；`capability` 表的 `mcp_endpoint` 字段记录该地址。
- orchestration-service 启动 Agent 时，根据 `agent.config_json` 里挂载的 `capability_id` 列表，实时从 marketplace-service 拉取 MCP 连接信息并建立会话，无需重新编译部署即可让 Agent 使用新上架的能力。
- 【待确认】是否所有能力都强制要求 MCP 接入，还是允许平台内置能力走进程内直接调用（性能更优但耦合更紧），建议内置能力走进程内调用、第三方能力强制走 MCP。

### 6.3 限流实现

- 令牌桶算法，基于 Redis + Lua 脚本保证原子性，Key 设计为 `ratelimit:{dimension}:{id}`（dimension 包括 apikey/agent/tenant/token）。
- Token 消耗类限流采用滑动窗口计数，避免固定窗口边界突刺；窗口数据同样存 Redis，TTL 与窗口周期对齐。
- 网关层在鉴权通过后、转发到 orchestration-service 前完成限流校验，避免无效请求打到编排引擎。

### 6.4 鉴权实现

- API Key 只存哈希值（`key_hash`），校验时对请求 Header 中的 Key 做同样哈希后比对，避免数据库泄露导致 Key 泄露。
- Scope 校验实现为网关中间件：解析 JWT/Key 关联的 `scopes` 字段，与请求路径中的 `agent_id` 及操作类型比对，不匹配直接返回 401。
- Workbench 走独立的 Session 鉴权分支，不复用 API Key 校验逻辑，但复用同一个限流和用量统计中间件。

### 6.5 沙箱隔离方案

- 每次 Agent 执行涉及代码执行/文件操作类工具时，在独立容器中运行，会话结束即销毁，容器间无共享文件系统，防止跨租户/跨会话数据泄漏。
- 【待确认】容器隔离粒度：按会话隔离（成本较高但隔离性强）还是按租户隔离（成本较低但会话间存在弱隔离风险），建议高风险操作（代码执行）按会话隔离，低风险只读操作（网页抓取）可放宽到租户级隔离池以控制成本。

### 6.6 计费与分成实现

- 用量事件（Token 消耗、调用次数）在每次 Agent 运行结束后异步写入消息队列，由 billing-service 消费并落库到 `usage_record`，避免同步计费拖慢主链路响应时间。
- 涉及 Subagent 调用的场景，用量按调用链路拆分记账：Hub 自身消耗记入 Hub 的 `usage_record`，Subagent 消耗记入其 `capability_id`，用于后续按能力维度给发布方结算分成。

### 6.7 高可用设计

```
服务层：
- 无状态设计（orchestration-service 不在本地保存会话状态，会话状态存 Redis/Checkpoint 存储）
- 熔断降级：单个 Subagent 调用失败率过高时自动降级为跳过该能力，返回部分结果 + 提示
- 超时控制：Subagent 调用统一设置超时（建议 30s），超时视为该分支失败但不影响 Hub 主流程

数据层：
- MySQL 主从，元数据读多写少场景走从库
- Redis：限流和会话缓存用 Cluster 模式，避免单点

基础设施：
- 多可用区部署，orchestration-service 无状态可任意扩容
```

---

## 七、非功能需求方案

| 需求 | 指标 | 方案 |
|------|------|------|
| 可用性 | 99.9% | 无状态服务 + 多副本 + 自动故障转移 |
| 响应时间 | 首字节 P99 < 2s（流式场景） | SSE 流式返回，编排引擎首 token 尽快下发 |
| 递归安全 | 最大深度 3 层，超限 100% 拦截 | gRPC 调用链显式传递 depth 字段并校验 |
| 计费准确性 | 跨通道用量误差 < 1% | 用量统计下沉到 orchestration-service 统一上报，不由网关层各自统计 |
| 隔离安全 | 沙箱逃逸零容忍 | 容器化 + 最小权限运行时 |
| 可观测性 | 全链路 Trace 覆盖多层 Subagent | Eino Callback 事件流 + OpenTelemetry 串联 trace_id |

---

## 八、部署架构

```
生产环境：
  Kubernetes 集群
  ├── gateway-service：Deployment，副本数 ≥ 3，前置 Ingress + 负载均衡
  ├── orchestration-service：Deployment，无状态，按 CPU/内存自动扩缩容（HPA）
  ├── marketplace-service / billing-service：Deployment，副本数 ≥ 2
  ├── sandbox-runtime：独立节点池，容器执行隔离，限制资源配额
  └── 数据层：MySQL 主从（云托管）、Redis Cluster、消息队列（云托管）

CI/CD 流程：
  代码提交 → 单测（含 Eino Agent 配置的 mock 测试）→ 构建镜像
  → 灰度发布（先在 orchestration-service 验证新 Hook/编排逻辑）→ 全量
```

---

## 九、前端技术方案

### 9.1 技术选型与理由

| 选型 | 说明 |
|------|------|
| Vite + React 19 + TypeScript | 冷启动和 HMR 速度快，团队若已有 React 经验可直接复用；TypeScript 与后端 Go 的强类型风格呼应，减少接口联调时的隐性错误 |
| Tailwind CSS v4 | 用 `@tailwindcss/vite` 插件，主题变量直接写在 CSS 里（`@theme inline`），不需要额外的 `tailwind.config.js`，降低配置层级 |
| shadcn/ui 风格组件 | 不是一个 npm 依赖包，而是"把组件源码复制进项目"的模式——组件可以按需修改样式（比如骨架里给 Card/Button 加了毛玻璃效果），不受限于第三方组件库的封装边界 |
| react-router-dom | 客户端路由，MVP 阶段页面数量少（4个），暂不需要引入更重的路由框架 |

【假设】团队具备 React 生态经验；如果团队更熟悉 Vue，架构原则（状态管理分层、API 层封装）同样适用，只需替换实现框架。

### 9.2 状态管理分层

不引入 Redux 这类全局状态管理框架，采用按数据生命周期分层的策略：

| 数据类型 | 方案 | 说明 |
|---------|------|------|
| 服务端数据（能力列表、Agent 配置、用量统计） | TanStack Query | 自带缓存、重试、失效策略，覆盖产品规格文档里"能力检索""用量查询"等场景，避免手写 loading/error 状态模板代码 |
| 表单临时状态（Agent 构建器里勾选的能力、Hook 开关） | React 组件内 `useState` | 生命周期仅限当前页面，不需要跨页面共享 |
| 会话级状态（Workbench 对话历史） | 组件内 `useState` + 后续可选 `sessionStorage`（注意：不能用浏览器 storage API 存敏感信息，且 Claude 生成的 artifact 环境本身禁用浏览器存储，正式项目中不受此限制，可正常使用） |
| 全局但轻量的状态（当前登录用户、API Key 选择） | React Context | 数据量小、更新频率低，不需要 TanStack Query 或额外状态库 |

引入 TanStack Query 是 M1 阶段的任务（见实现计划 §3.5），当前骨架代码里 `src/api/client.ts` 是裸 `fetch` 封装，接入真实后端时替换为 Query 的 `useQuery`/`useMutation`。

### 9.3 API 层与类型同步

- `src/api/client.ts` 是唯一与 gateway-service 通信的出口，页面组件不直接调用 `fetch`。
- 【建议】gateway-service 按技术规格文档 §6.3 自动生成 OpenAPI 文档后，前端用 `openapi-typescript` 之类的工具从 `openapi.json` 生成 TS 类型，替换 `src/types/index.ts` 里手写的类型定义，避免前后端字段漂移。
- SSE（`streamMsg`）在 `client.ts` 里返回原生 `EventSource`，页面组件负责监听 `onmessage` 并增量更新 UI；后续如果多个页面都要处理流式响应，可以抽成一个 `useAgentStream` 自定义 Hook 复用。

### 9.4 组件分层规范

```
src/
├── components/ui/     shadcn 风格基础组件（button/card/badge/input…），不包含业务逻辑
├── components/        业务组件（跨页面复用的组合，如能力卡片、Hook 开关组），当前骨架尚未拆分，
│                       M1 阶段随着页面复杂度上升再从 pages/ 里提取
├── pages/              路由级页面，负责数据获取（用 TanStack Query）和页面级状态编排
├── layouts/            页面外壳（导航栏等），不含业务数据
└── data/               MVP 阶段的示例数据，接入真实接口后逐步移除
```

原则：`components/ui` 永远不知道"能力""Agent"这些业务概念，只处理样式和交互；业务概念只出现在 `components/`（业务组件）和 `pages/`。

### 9.5 性能与加载体验

- 路由级代码分割：`React.lazy` + `Suspense` 按页面拆包，避免首屏加载四个页面的全部 JS（当前骨架页面少，先不做，页面数量增长到 6+ 后再引入）。
- 空状态与错误态：能力市场为空、Agent 尚未发布等场景需要设计明确的空状态提示（参考产品设计里"空状态是行动邀请"的原则），不能只显示空白网格。
- 流式响应的渲染：Workbench 收到 SSE chunk 时应增量追加文本而不是整体重渲染消息列表，避免长回复时出现明显卡顿。

### 9.6 测试策略

| 层级 | 工具 | 覆盖范围 |
|------|------|---------|
| 单元测试 | Vitest + React Testing Library | `components/ui` 的交互逻辑、`lib/utils.ts` 等纯函数 |
| 集成测试 | Vitest + MSW（Mock Service Worker） | 页面级数据请求流程，mock gateway-service 响应 |
| 端到端测试 | Playwright | 核心用户旅程：浏览市场 → 构建 Agent → 发布 → Workbench 对话 |

当前骨架代码未包含测试文件，属于 M1 阶段任务范畴。

### 9.7 构建与部署

- `npm run build` 产出静态资源（`dist/`），可直接部署到 CDN / 对象存储 + CDN，不需要 Node.js 运行时。
- 环境变量通过 Vite 的 `.env` 机制区分（如 `VITE_API_BASE_URL`），避免把网关地址硬编码在 `client.ts` 里；本地开发走 `vite.config.ts` 里配置的 `/api` 代理。
- 【待确认】是否需要多租户品牌定制（White-label），如果需要，`src/index.css` 里的主题变量要改造成可通过配置接口动态加载，而不是当前的硬编码值。

---

## 十、风险与演进路径

```
当前方案的局限性：
- 风险 1：Eino/Eino ADK 相对较新，社区案例和第三方工具生态不如 Python 体系成熟 → 优先用 MCP 协议解耦，降低对框架本身工具生态的依赖
- 风险 2：MCP 动态发现引入网络调用开销，可能影响首字节延迟 → 高频能力可加本地缓存/连接池
- 风险 3：沙箱隔离粒度的成本与安全权衡尚未最终确定 → 需要压测后再定档

演进路径：
阶段一（MVP）：orchestration-service 跑通 Agentic Loop + 基础 Hook；网关完成鉴权限流；市场支持内置能力，MCP 接入暂缓
阶段二：开放 MCP 第三方能力接入、Subagent-as-Tool 多 Agent 协同上线、计费分成机制上线
阶段三：确定性 Pipeline 编排、多租户精细化隔离、容量按需扩展
```

---

## 十一、待确认事项清单

- 是否所有第三方能力强制走 MCP 协议接入，还是允许平台内置能力走进程内调用
- 沙箱隔离粒度：按会话隔离还是按租户隔离
- Python 生态第三方 Tool 的兼容方案，是否完全依赖 MCP 协议解决语言异构问题
- 前端是否需要支持多租户品牌定制（White-label），影响是否要把设计 token 做成可配置主题而非硬编码
- Workbench 的调试模式（展示 Hook/Tool 调用详情）是否对所有用户开放，还是仅开发者角色可见
