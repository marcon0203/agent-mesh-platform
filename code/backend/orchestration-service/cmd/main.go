// orchestration-service 是平台的执行引擎。gRPC :9090 服务于 gateway-service 的
// 内部调用（数据面，Invoke）；HTTP :8082 是配置面（Agent 装配/发布、模型供应商
// 增删查），供前端 AgentBuilderPage 和模型供应商配置页直接调用。
//
// 分层结构（DDD 风格，见 code/backend/README.md）：
//
//	internal/domain          Agent / ModelProvider 聚合根：递归深度、能力挂载、Hook 配置等不变量
//	internal/application     用例编排：InvokeUseCase / ConfigureAgentUseCase / ModelProviderUseCase
//	internal/infrastructure  技术细节：Eino ADK 接入、MySQL 仓储实现、异步用量上报
//	internal/interfaces      gRPC / HTTP 协议转换层
//
// 详见 docs/Agent开放平台_技术规格文档.md 第六章 6.1 节。
package main

import (
	"log"
	"net"
	"net/http"
	"os"

	"google.golang.org/grpc"

	orchestrationpb "github.com/agentmesh/shared/pkg/orchestration"

	"github.com/agentmesh/orchestration-service/internal/application"
	"github.com/agentmesh/orchestration-service/internal/infrastructure"
	"github.com/agentmesh/orchestration-service/internal/interfaces"
)

func postgresDSN() string {
	if dsn := os.Getenv("ORCHESTRATION_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable"
}

func redisAddr() string {
	if addr := os.Getenv("ORCHESTRATION_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}

func main() {
	db, err := infrastructure.NewPostgresConnection(postgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	// 依赖注入：infrastructure 实现 domain 定义的端口，application 只依赖端口
	agentRepo := infrastructure.NewAgentMySQLRepository(db)
	modelProviderRepo, err := infrastructure.NewModelProviderMySQLRepository(db)
	if err != nil {
		log.Fatalf("failed to init model provider repository (check MODEL_PROVIDER_ENC_KEY): %v", err)
	}
	runtime := infrastructure.NewEinoRuntime(modelProviderRepo)
	usageReporter := infrastructure.NewAsyncUsageReporter(redisAddr())

	invokeUseCase := application.NewInvokeUseCase(agentRepo, runtime, usageReporter)
	createAgentUseCase := application.NewCreateAgentUseCase(agentRepo)
	configureAgentUseCase := application.NewConfigureAgentUseCase(agentRepo)
	modelProviderUseCase := application.NewModelProviderUseCase(modelProviderRepo)

	grpcHandler := interfaces.NewOrchestrationGRPCServer(invokeUseCase)
	adminHandler := interfaces.NewAdminHTTPHandler(createAgentUseCase, configureAgentUseCase, modelProviderUseCase)

	go serveAdminHTTP(adminHandler)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	orchestrationpb.RegisterOrchestrationServiceServer(grpcServer, grpcHandler)

	log.Println("orchestration-service listening on :9090 (internal gRPC)")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func serveAdminHTTP(handler *interfaces.AdminHTTPHandler) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /agents", handler.CreateAgentHandler)
	mux.HandleFunc("GET /agents/{id}", handler.GetAgent)
	mux.HandleFunc("POST /agents/{id}/config", handler.ConfigureAgentHandler)
	mux.HandleFunc("POST /model-providers", handler.CreateModelProvider)
	mux.HandleFunc("GET /model-providers", handler.ListModelProviders)

	log.Println("orchestration-service listening on :8082 (admin HTTP: agent config / model providers)")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
