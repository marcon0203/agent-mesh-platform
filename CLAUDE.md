# 项目上下文（Claude Code 自动读取）

这是「枢络 AgentMesh」Agent 开放平台的代码仓库。开始任何开发任务前，先读：

1. `docs/Agent开放平台_产品规格文档.md` —— 产品设计，回答"为什么这么做"
2. `docs/Agent开放平台_技术规格文档.md` —— 技术方案，回答"具体怎么做"
3. `docs/Agent开放平台_实现计划.md` —— **当前应该做哪个任务**，按里程碑（M0/M1/M2/M3）和依赖顺序列了任务表格
4. `code/backend/README.md` —— 为什么 orchestration-service / marketplace-service 用 DDD 四层，
   gateway-service / billing-service 用简单三层
5. `code/frontend/README.md` —— 前端技术选型与目录结构

## 开发约定

- 按 `docs/Agent开放平台_实现计划.md` 里的任务表格顺序开发，每个任务改完后跑对应服务的
  `go build ./...`（后端）或 `npm run build`（前端）验证再继续下一个任务。
- `orchestration-service` 和 `marketplace-service` 的业务规则优先写进聚合根
  （`internal/domain/agent.go` 的 `Agent`、`internal/domain/capability.go` 的 `Capability`），
  不要在 `application`/`infrastructure` 层散写校验逻辑绕过聚合根。
- `internal/domain` 和 `internal/application` 两层不允许 import 任何第三方框架包
  （不 import Eino、Hertz、gRPC 等），这些技术细节只能出现在 `internal/infrastructure`
  和 `internal/interfaces` 里。
- 实现计划里标 `【需人工决策】` 的任务，先跟人确认方案再写代码，不要自行假设。
- Eino / Eino ADK 的具体 API 用法以本地 `go doc` 查到的真实签名为准，不要凭训练记忆硬编——
  框架版本可能已经变化。

## 常用命令

```bash
# 基础设施
docker compose -f infra/docker-compose.yml up -d

# 后端（任一服务）
cd code/backend/<service-name> && go build ./... && go run ./cmd

# 前端
cd code/frontend && npm install && npm run dev
```
