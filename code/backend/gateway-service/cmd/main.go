// gateway-service 是平台唯一对外入口。
// 职责：鉴权（API Key / OAuth2 / Session）、限流、路由到 orchestration-service、
// SSE 转发、按 Agent 自动生成 OpenAPI 文档。
// 详细设计见 docs/Agent开放平台_技术规格文档.md 第三、六章。
package main

import (
	"context"
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"

	"github.com/agentmesh/gateway-service/internal/handler"
)

func main() {
	h := server.Default(server.WithHostPorts(":8080"))

	// TODO: 注册鉴权中间件（API Key 哈希校验 + Scope 校验），见 internal/service/auth.go
	// TODO: 注册限流中间件（Redis + Lua 令牌桶），见 internal/service/ratelimit.go
	h.Use(handler.AuthMiddleware(), handler.RateLimitMiddleware())

	v1 := h.Group("/api/v1")
	{
		v1.POST("/agents/:agent_id/sendMsg", handler.SendMsg)
		v1.POST("/agents/:agent_id/streamMsg", handler.StreamMsg)
		v1.GET("/agents/:agent_id/openapi.json", handler.OpenAPISpec)
	}

	log.Println("gateway-service listening on :8080")
	h.Spin()
	_ = context.Background()
}
