# Backend（Go + Eino / Eino ADK + Hertz）

对应 `docs/Agent开放平台_技术规格文档.md` 的服务划分（第三章）。

```
backend/
├── gateway-service/         对外唯一入口：鉴权、限流、路由、SSE、OpenAPI 生成（Hertz）
├── orchestration-service/   执行引擎：Agentic Loop、Subagent-as-Tool、Hook 切面（Eino / Eino ADK）
├── marketplace-service/     能力发布、审核、动态发现（MCP 协议注册中心）
├── billing-service/         异步用量统计与分成结算
└── shared/                  跨服务共享的 proto 定义与公共包
```

## 分层策略：不是所有服务都用 DDD

四个服务的领域复杂度差别很大，只在真正有业务不变量的服务上用 DDD 四层，其余保持简单的三层结构。

**orchestration-service / marketplace-service —— DDD 四层：**

```
internal/
├── domain/          聚合根 + 值对象 + 仓储/端口接口，不依赖任何框架
│                     orchestration: Agent 聚合（递归深度、能力挂载、Hook 配置等不变量）
│                     marketplace:  Capability 聚合（状态机流转：待审核→已上架→已下线）
├── application/      用例编排：调用 domain + 调 infrastructure，不包含业务规则本身
├── infrastructure/   技术细节：Eino ADK 接入 / MCP 注册中心 / 仓储实现（当前为内存实现占位）
└── interfaces/       对外的 gRPC/HTTP handler，只做协议转换，不含业务逻辑
```

这两个服务的核心价值在于"规则不能被绕过"——比如递归深度限制、能力状态机的合法流转，
封装进聚合根后，只有一个地方能改这些规则，任何入口都必须经过校验。

**gateway-service / billing-service —— 简单三层：**

```
internal/
├── handler/     HTTP handler + 中间件（鉴权/限流）
├── service/     业务逻辑
└── repository/  数据访问
```

gateway-service 本质是基础设施编排（鉴权、限流、路由转发），没有业务不变量，强行分领域层
只会让一个简单的中间件函数绕三层才能改一行逻辑。billing-service 更像"消费事件 → 计算 →
落库"的处理管道，核心复杂度在分成算法而不是状态管理，不需要聚合根和仓储接口整套。

## 本地开发

```bash
docker compose -f ../../infra/docker-compose.yml up -d   # PostgreSQL / Redis

cd gateway-service && go run ./cmd
cd orchestration-service && go run ./cmd
cd marketplace-service && go run ./cmd
cd billing-service && go run ./cmd
```

推荐直接用仓库根目录的 `make dev`（见根 `README.md`），会按依赖顺序把基础设施和四个服务都起好。

## 当前状态

M1 里程碑的任务表格已经全部实现（不再是骨架/TODO 状态）：

- [x] gateway-service：API Key 鉴权中间件、Redis 令牌桶限流、转发到 orchestration-service、
      按 Agent 动态生成 OpenAPI 文档
- [x] orchestration-service：`infrastructure/eino_runtime.go` 接入真实 Eino ADK
      `ChatModelAgent` + Callback 绑定；ModelProvider 聚合根（用户自建模型供应商）；
      仓储从内存实现换成 PostgreSQL；用量上报用 asynq 把事件放进 Redis 任务队列
- [x] marketplace-service：内置能力（免走 MCP）静态注册；仓储从内存实现换成 PostgreSQL
- [x] billing-service：asynq 消费用量事件、落库 `usage_record`、对外用量查询接口

MCP 第三方能力接入、Subagent-as-Tool 递归包装、分成结算算法属于 M2 范围，还没有实现。

数据库表结构见技术规格文档第四章 / `infra/schema.sql`（PostgreSQL 方言），可直接执行建表；
`make schema` 会自动跑这一步。
