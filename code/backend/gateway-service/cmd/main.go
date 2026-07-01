// gateway-service 是平台唯一对外入口。
// 职责：鉴权（API Key / OAuth2 / Session）、限流、路由到 orchestration-service、
// SSE 转发、按 Agent 自动生成 OpenAPI 文档。
// 详细设计见 docs/Agent开放平台_技术规格文档.md 第三、六章。
package main

import (
	"log"
	"os"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	orchestrationpb "github.com/agentmesh/shared/pkg/orchestration"

	"github.com/agentmesh/gateway-service/internal/handler"
)

func postgresDSN() string {
	if dsn := os.Getenv("GATEWAY_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable"
}

func redisAddr() string {
	if addr := os.Getenv("GATEWAY_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}

func orchestrationAddr() string {
	if addr := os.Getenv("ORCHESTRATION_GRPC_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:9090"
}

func main() {
	db, err := handler.NewPostgresConnection(postgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr()})

	conn, err := grpc.NewClient(orchestrationAddr(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to dial orchestration-service: %v", err)
	}
	orchestrationClient := orchestrationpb.NewOrchestrationServiceClient(conn)

	authService := handler.NewAuthService(db)
	rateLimitService := handler.NewRateLimitService(rdb)
	agentHandler := handler.NewAgentHandler(orchestrationClient)

	h := server.Default(server.WithHostPorts(":8080"))

	v1 := h.Group("/api/v1")
	{
		// openapi.json 是公开文档端点，不需要鉴权/限流。
		v1.GET("/agents/:agent_id/openapi.json", handler.OpenAPISpec)
		v1.POST("/agents/:agent_id/sendMsg", authService.Middleware(), rateLimitService.Middleware(), agentHandler.SendMsg)
		v1.POST("/agents/:agent_id/streamMsg", authService.Middleware(), rateLimitService.Middleware(), agentHandler.StreamMsg)
	}

	log.Println("gateway-service listening on :8080")
	h.Spin()
}
