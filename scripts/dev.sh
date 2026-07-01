#!/usr/bin/env bash
# 一键启动 AgentMesh 全部服务：基础设施 + 4 个后端服务（后台）+ 前端 dev server（前台）。
# Ctrl+C 退出时会自动清理所有后台进程。供 `make dev` 调用，也可以直接执行。
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
LOG_DIR="$ROOT_DIR/.run/logs"
mkdir -p "$LOG_DIR"

: "${POSTGRES_DSN:=postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable}"
: "${REDIS_ADDR:=127.0.0.1:6379}"
# 仅供本地开发使用的固定密钥（16 字节 hex，对应 AES-128）。
# 生产环境务必换成随机生成、妥善保管的密钥。
: "${MODEL_PROVIDER_ENC_KEY:=000102030405060708090a0b0c0d0e0f}"
: "${ORCHESTRATION_GRPC_ADDR:=127.0.0.1:9090}"

echo "==> 启动基础设施 (PostgreSQL / Redis)"
docker compose -f "$ROOT_DIR/infra/docker-compose.yml" up -d

echo "==> 等待 PostgreSQL 就绪…"
until docker compose -f "$ROOT_DIR/infra/docker-compose.yml" exec -T postgres \
  pg_isready -U agentmesh >/dev/null 2>&1; do
  sleep 1
done

echo "==> 执行 infra/schema.sql（表已存在会报错，属于正常现象，可忽略）"
docker compose -f "$ROOT_DIR/infra/docker-compose.yml" exec -T postgres \
  psql -U agentmesh -d agentmesh < "$ROOT_DIR/infra/schema.sql" || true

PIDS=()
cleanup() {
  echo ""
  echo "==> 停止所有后台服务"
  for pid in "${PIDS[@]:-}"; do
    [ -z "$pid" ] && continue
    pkill -P "$pid" 2>/dev/null || true
    kill "$pid" 2>/dev/null || true
  done
}
trap cleanup EXIT INT TERM

start_service() {
  local name="$1"; shift
  echo "==> 启动 $name（日志：.run/logs/$name.log）"
  ( cd "$ROOT_DIR/code/backend/$name" && "$@" ) > "$LOG_DIR/$name.log" 2>&1 &
  PIDS+=("$!")
}

MARKETPLACE_POSTGRES_DSN="$POSTGRES_DSN" \
  start_service marketplace-service go run ./cmd

ORCHESTRATION_POSTGRES_DSN="$POSTGRES_DSN" \
  ORCHESTRATION_REDIS_ADDR="$REDIS_ADDR" \
  MODEL_PROVIDER_ENC_KEY="$MODEL_PROVIDER_ENC_KEY" \
  start_service orchestration-service go run ./cmd

GATEWAY_POSTGRES_DSN="$POSTGRES_DSN" \
  GATEWAY_REDIS_ADDR="$REDIS_ADDR" \
  ORCHESTRATION_GRPC_ADDR="$ORCHESTRATION_GRPC_ADDR" \
  start_service gateway-service go run ./cmd

BILLING_POSTGRES_DSN="$POSTGRES_DSN" \
  BILLING_REDIS_ADDR="$REDIS_ADDR" \
  start_service billing-service go run ./cmd

echo "==> 等待后端服务起来…"
sleep 3
echo "    marketplace-service   http://localhost:8081"
echo "    orchestration-service admin http://localhost:8082, gRPC :9090"
echo "    gateway-service       http://localhost:8080"
echo "    billing-service       http://localhost:8083"

echo "==> 启动前端 dev server（前台运行，Ctrl+C 会一并清理上面的后端进程）"
cd "$ROOT_DIR/code/frontend"
if [ ! -d node_modules ]; then
  npm install
fi
npm run dev
