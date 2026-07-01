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
docker compose -f ../../infra/docker-compose.yml up -d   # MySQL / Redis / 消息队列

cd gateway-service && go run ./cmd
cd orchestration-service && go run ./cmd
cd marketplace-service && go run ./cmd
cd billing-service && go run ./cmd
```

`marketplace-service` 和其内部 `domain`/`application`/`infrastructure`/`interfaces` 四层
已在沙箱环境验证 `go build` 全部通过（无外部重依赖）。`orchestration-service` 的
`domain`/`application`/`infrastructure`/`interfaces` 四层同样编译通过；其 `cmd/main.go`
依赖 `google.golang.org/grpc`，需要在有完整网络访问的环境里 `go mod tidy` 才能构建。

## 当前状态

骨架代码，`TODO` 标记的位置对应技术规格文档里已确定方案、但还未实现的部分：

- [ ] gateway-service：API Key 鉴权中间件、Redis 令牌桶限流
- [ ] orchestration-service：`infrastructure/eino_runtime.go` 里接入真实 Eino ADK
      `ChatModelAgent` 构造与 Callback 绑定；仓储从内存实现换成 MySQL
- [ ] marketplace-service：`infrastructure/mcp_registry_adapter.go` 接入真实 MCP 健康检查；
      仓储从内存实现换成 MySQL
- [ ] billing-service：消息队列消费者、分成结算逻辑

数据库表结构见技术规格文档第四章 / `infra/schema.sql`，可直接执行建表。
