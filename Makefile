SHELL := /bin/bash

MYSQL_DSN ?= root:agentmesh@tcp(127.0.0.1:3306)/agentmesh?parseTime=true
REDIS_ADDR ?= 127.0.0.1:6379
AMQP_URL ?= amqp://guest:guest@127.0.0.1:5672/
MODEL_PROVIDER_ENC_KEY ?= 000102030405060708090a0b0c0d0e0f
ORCHESTRATION_GRPC_ADDR ?= 127.0.0.1:9090

.DEFAULT_GOAL := help

.PHONY: help
help:
	@echo "枢络 AgentMesh 常用命令："
	@echo ""
	@echo "  make dev              一键启动：基础设施 + 4 个后端服务(后台) + 前端(前台)，Ctrl+C 全部退出"
	@echo "  make infra-up         起 MySQL / Redis / RabbitMQ（docker compose）"
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
	@echo "等待 MySQL 就绪…"
	@until docker compose -f infra/docker-compose.yml exec -T mysql mysqladmin ping -h127.0.0.1 -uroot -pagentmesh --silent >/dev/null 2>&1; do sleep 1; done
	docker compose -f infra/docker-compose.yml exec -T mysql mysql -uroot -pagentmesh agentmesh < infra/schema.sql

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
	cd code/backend/marketplace-service && MARKETPLACE_MYSQL_DSN="$(MYSQL_DSN)" go run ./cmd

.PHONY: run-orchestration
run-orchestration:
	cd code/backend/orchestration-service && \
		ORCHESTRATION_MYSQL_DSN="$(MYSQL_DSN)" \
		ORCHESTRATION_AMQP_URL="$(AMQP_URL)" \
		MODEL_PROVIDER_ENC_KEY="$(MODEL_PROVIDER_ENC_KEY)" \
		go run ./cmd

.PHONY: run-gateway
run-gateway:
	cd code/backend/gateway-service && \
		GATEWAY_MYSQL_DSN="$(MYSQL_DSN)" \
		GATEWAY_REDIS_ADDR="$(REDIS_ADDR)" \
		ORCHESTRATION_GRPC_ADDR="$(ORCHESTRATION_GRPC_ADDR)" \
		go run ./cmd

.PHONY: run-billing
run-billing:
	cd code/backend/billing-service && \
		BILLING_MYSQL_DSN="$(MYSQL_DSN)" \
		BILLING_AMQP_URL="$(AMQP_URL)" \
		go run ./cmd

.PHONY: frontend
frontend:
	cd code/frontend && npm install && npm run dev
