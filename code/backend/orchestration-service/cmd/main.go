// orchestration-service 是平台的执行引擎，不直接对外暴露，只服务于
// gateway-service 的内部 gRPC 调用。
//
// 分层结构（DDD 风格，见 code/backend/README.md）：
//   internal/domain          Agent 聚合根：递归深度、能力挂载、Hook 配置等不变量
//   internal/application     用例编排：InvokeUseCase / ConfigureAgentUseCase
//   internal/infrastructure  技术细节：Eino ADK 接入、仓储实现、异步用量上报
//   internal/interfaces      gRPC 协议转换层
//
// 详见 docs/Agent开放平台_技术规格文档.md 第六章 6.1 节。
package main

import (
	"log"
	"net"

	"google.golang.org/grpc"

	"github.com/agentmesh/orchestration-service/internal/application"
	"github.com/agentmesh/orchestration-service/internal/infrastructure"
	"github.com/agentmesh/orchestration-service/internal/interfaces"
)

func main() {
	// 依赖注入：infrastructure 实现 domain 定义的端口，application 只依赖端口
	agentRepo := infrastructure.NewInMemoryAgentRepository() // TODO: 替换为 MySQL 实现
	runtime := infrastructure.NewEinoRuntime()
	usageReporter := infrastructure.NewAsyncUsageReporter()

	invokeUseCase := application.NewInvokeUseCase(agentRepo, runtime, usageReporter)
	grpcHandler := interfaces.NewOrchestrationGRPCServer(invokeUseCase)
	_ = grpcHandler

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	// TODO: RegisterOrchestrationServiceServer(grpcServer, grpcHandler)

	log.Println("orchestration-service listening on :9090 (internal gRPC)")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
