import path from "node:path"
import { defineConfig } from "vite"
import react from "@vitejs/plugin-react"
import tailwindcss from "@tailwindcss/vite"

export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  server: {
    // 本地开发时把 /api/v1 按路径前缀分流到三个后端服务：
    // - gateway-service（:8080）只承接对外调用面：sendMsg/streamMsg/openapi.json
    // - marketplace-service（:8081）承接能力市场的检索/发布
    // - orchestration-service 的 admin HTTP（:8082）承接 Agent 配置面和模型供应商增删查
    // 前端代码始终只感知统一的 /api/v1 路径，不需要关心背后是哪个服务，
    // 匹配顺序很重要：更具体的规则要排在通用的 /agents 规则前面。
    proxy: {
      "^/api/v1/capabilities": {
        target: "http://localhost:8081",
        rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
      "^/api/v1/model-providers": {
        target: "http://localhost:8082",
        rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
      // 精确匹配的三条规则要允许可选的查询串（比如 GET /agents?status=published），
      // 不能简单写 [^/]+$ / agents$ —— vite 的 proxy context 是拿完整 req.url
      // （含查询串）去测正则的，末尾锚定 $ 会导致带 query 的请求直接匹配失败、
      // 落到最后的通用 /api 规则上，被错误转发到 gateway-service。
      "^/api/v1/agents/[^/?]+/config(\\?.*)?$": {
        target: "http://localhost:8082",
        rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
      "^/api/v1/agents/[^/?]+(\\?.*)?$": {
        target: "http://localhost:8082",
        rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
      "^/api/v1/agents(\\?.*)?$": {
        target: "http://localhost:8082",
        rewrite: (path) => path.replace(/^\/api\/v1/, ""),
      },
      "/api": "http://localhost:8080",
    },
  },
})
