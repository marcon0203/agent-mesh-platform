// gateway-service 是平台唯一对外入口。
// 职责：鉴权（API Key / OAuth2 / Session）、限流、路由到 orchestration-service、
// SSE 转发、按 Agent 自动生成 OpenAPI 文档。
// 配置来自服务目录下的 config.yaml，环境变量可覆盖同名字段（见
// internal/handler/config.go），本地开发不改配置直接跑就行。
// 详细设计见 docs/Agent开放平台_技术规格文档.md 第三、六章。
package main

import (
	"log"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orchestrationpb "github.com/agentmesh/shared/pkg/orchestration"

	"github.com/agentmesh/gateway-service/internal/handler"
)

func main() {
	cfg, err := handler.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := handler.NewPostgresConnection(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: cfg.RedisAddr})

	conn, err := grpc.NewClient(cfg.OrchestrationGRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial orchestration-service: %v", err)
	}
	orchestrationClient := orchestrationpb.NewOrchestrationServiceClient(conn)

	authService := handler.NewAuthService(db)
	rateLimitService := handler.NewRateLimitService(rdb)
	agentHandler := handler.NewAgentHandler(orchestrationClient)

	h := server.Default(server.WithHostPorts(cfg.HTTPAddr))

	v1 := h.Group("/api/v1")
	{
		// openapi.json 是公开文档端点，不需要鉴权/限流。
		v1.GET("/agents/:agent_id/openapi.json", handler.OpenAPISpec)
		v1.POST("/agents/:agent_id/sendMsg", authService.Middleware(), rateLimitService.Middleware(), agentHandler.SendMsg)
		v1.POST("/agents/:agent_id/streamMsg", authService.Middleware(), rateLimitService.Middleware(), agentHandler.StreamMsg)
	}

	log.Printf("gateway-service listening on %s\n", cfg.HTTPAddr)
	h.Spin()
}
