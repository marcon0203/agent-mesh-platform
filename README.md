# 枢络 AgentMesh · Agent 开放平台

```
agent-mesh-platform/
├── CLAUDE.md                      Claude Code 自动读取的项目上下文与开发约定
├── docs/                          产品与技术规格文档
│   ├── Agent开放平台_产品规格文档.md
│   ├── Agent开放平台_技术规格文档.md
│   ├── Agent开放平台_实现计划.md    面向 Claude Code 的分阶段任务清单（当前应该做什么）
│   └── design/
│       └── portal-reference.html   门户视觉参考（科技蓝毛玻璃风格）
│
├── code/
│   ├── backend/                   Go + Eino / Eino ADK + Hertz
│   │   ├── gateway-service/        对外入口：鉴权/限流/路由/OpenAPI
│   │   ├── orchestration-service/  执行引擎：Agentic Loop / Subagent-as-Tool / Hooks
│   │   ├── marketplace-service/    能力发布/审核/动态发现（MCP）
│   │   ├── billing-service/        用量统计与分成结算
│   │   └── shared/                 跨服务 proto 与公共包
│   │
│   └── frontend/                  React + TypeScript + shadcn/ui + Tailwind v4
│       └── src/pages/              能力市场 / Agent 构建器 / Workbench / 控制台
│
└── infra/
    ├── docker-compose.yml          本地开发基础设施（MySQL/Redis/RabbitMQ）
    └── schema.sql                  核心数据库表结构
```

## 快速开始

**一键启动**（基础设施 + 4 个后端服务 + 前端，Ctrl+C 退出会一并清理后台进程）：
```bash
make dev
```

**分开手动跑**（方便单独调试某个服务）：
```bash
docker compose -f infra/docker-compose.yml up -d
make schema                 # 执行 infra/schema.sql 建表（第一次跑需要）

make run-marketplace        # :8081
make run-orchestration      # gRPC :9090，admin HTTP :8082
make run-gateway            # :8080
make run-billing            # :8083
make frontend                # :5173
```

`make help` 能看到全部命令。各服务需要的环境变量（MySQL DSN、Redis 地址、RabbitMQ URL、
模型供应商 API Key 加密密钥）都在 Makefile 里给了本地开发默认值，无需额外配置即可跑通。

## 文档与代码的对应关系

骨架代码中的每个 `TODO` 都能在 `docs/` 里找到对应的方案说明：

| 代码位置 | 对应文档章节 |
|---------|-------------|
| `gateway-service/internal/handler/middleware.go` | 产品规格文档 §6.1 鉴权、§6.2 限流 |
| `orchestration-service/internal/domain/agent.go` | 产品规格文档 §4 核心机制（递归深度、能力挂载等不变量） |
| `orchestration-service/internal/infrastructure/eino_runtime.go` | 技术规格文档 §6.1 基于 Eino ADK 的编排实现 |
| `orchestration-service/internal/infrastructure/hooks_registry.go` | 产品规格文档 §3 中间件 Hooks 层 |
| `marketplace-service/internal/domain/capability.go` | 产品规格文档 §5 Marketplace 特有机制（状态机） |
| `marketplace-service/internal/infrastructure/mcp_registry_adapter.go` | 技术规格文档 §6.2 能力动态发现 |
| `billing-service/` | 产品规格文档 §5.3 计费与分成、技术规格文档 §6.6 |
| `infra/schema.sql` | 技术规格文档 §4 数据库设计 |
| `code/frontend/src/pages/AgentBuilderPage.tsx` | 产品规格文档 §4.1 能力装配（不是编排） |

`orchestration-service` 与 `marketplace-service` 采用 DDD 四层（domain/application/
infrastructure/interfaces），`gateway-service` 与 `billing-service` 保持简单三层，
取舍理由见 `code/backend/README.md`。

## 当前状态

这是一个可编译、可运行的骨架（前端已验证 `npm install` / `tsc` / `npm run build` 全部通过），
业务逻辑均为 TODO 占位，用于团队按文档分工并行开发。
