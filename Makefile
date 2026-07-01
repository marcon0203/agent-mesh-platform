SHELL := /bin/bash

POSTGRES_DSN ?= postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable
REDIS_ADDR ?= 127.0.0.1:6379
MODEL_PROVIDER_ENC_KEY ?= 000102030405060708090a0b0c0d0e0f
ORCHESTRATION_GRPC_ADDR ?= 127.0.0.1:9090

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "枢络 AgentMesh 常用命令："
	@echo ""
	@echo "  make dev              一键启动：基础设施 + 4 个后端服务(后台) + 前端(前台)，Ctrl+C 全部退出"
	@echo "  make infra-up         起 PostgreSQL / Redis（docker compose）"
	@echo "  make infra-down       停止基础设施容器"
	@echo "  make schema           执行 infra/schema.sql 建表"
	@echo "  make build            编译 4 个后端服务 + 构建前端"
	@echo "  make run-marketplace  前台单独运行 marketplace-service（:8081）"
	@echo "  make run-orchestration 前台单独运行 orchestration-service（gRPC :9090，admin HTTP :8082）"
	@echo "  make run-gateway      前台单独运行 gateway-service（:8080）"
	@echo "  make run-billing      前台单独运行 billing-service（:8083）"
	@echo "  make frontend         前台单独运行前端 dev server（:5173）"
	@echo ""
	@echo "各服务依赖的环境变量都给了本地开发默认值，直接 make dev 即可；"
	@echo "MODEL_PROVIDER_ENC_KEY 只是本地开发用的固定密钥，生产环境需要换成随机密钥。"
	@echo "Redis 身兼两职：gateway-service 的限流令牌桶，以及 orchestration/billing 之间"
	@echo "用 asynq 传递用量事件的任务队列存储（不再需要单独部署 RabbitMQ）。"

.PHONY: dev
dev:
	@./scripts/dev.sh

.PHONY: infra-up
infra-up:
	docker compose -f infra/docker-compose.yml up -d

.PHONY: infra-down
infra-down:
	docker compose -f infra/docker-compose.yml down

.PHONY: schema
schema:
	@echo "等待 PostgreSQL 就绪…"
	@until docker compose -f infra/docker-compose.yml exec -T postgres pg_isready -U agentmesh >/dev/null 2>&1; do sleep 1; done
	docker compose -f infra/docker-compose.yml exec -T postgres psql -U agentmesh -d agentmesh < infra/schema.sql

.PHONY: build build-backend build-frontend
build: build-backend build-frontend

build-backend:
	cd code/backend/shared && go build ./...
	cd code/backend/marketplace-service && go build ./...
	cd code/backend/orchestration-service && go build ./...
	cd code/backend/gateway-service && go build ./...
	cd code/backend/billing-service && go build ./...

build-frontend:
	cd code/frontend && npm install && npm run build

.PHONY: run-marketplace
run-marketplace:
	cd code/backend/marketplace-service && MARKETPLACE_POSTGRES_DSN="$(POSTGRES_DSN)" go run ./cmd

.PHONY: run-orchestration
run-orchestration:
	cd code/backend/orchestration-service && \
		ORCHESTRATION_POSTGRES_DSN="$(POSTGRES_DSN)" \
		ORCHESTRATION_REDIS_ADDR="$(REDIS_ADDR)" \
		MODEL_PROVIDER_ENC_KEY="$(MODEL_PROVIDER_ENC_KEY)" \
		go run ./cmd

.PHONY: run-gateway
run-gateway:
	cd code/backend/gateway-service && \
		GATEWAY_POSTGRES_DSN="$(POSTGRES_DSN)" \
		GATEWAY_REDIS_ADDR="$(REDIS_ADDR)" \
		ORCHESTRATION_GRPC_ADDR="$(ORCHESTRATION_GRPC_ADDR)" \
		go run ./cmd

.PHONY: run-billing
run-billing:
	cd code/backend/billing-service && \
		BILLING_POSTGRES_DSN="$(POSTGRES_DSN)" \
		BILLING_REDIS_ADDR="$(REDIS_ADDR)" \
		go run ./cmd

.PHONY: frontend
frontend:
	cd code/frontend && npm install && npm run dev
