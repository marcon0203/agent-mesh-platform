// marketplace-service 负责能力（Tool/Skill/Agent）的发布、审核、检索与动态发现。
//
// 分层结构（DDD 风格，见 code/backend/README.md）：
//   internal/domain          Capability 聚合根：状态机流转、版本规则等不变量
//   internal/application     用例编排：PublishCapabilityUseCase / DiscoverCapabilityUseCase
//   internal/infrastructure  技术细节：仓储实现、MCP 注册中心适配器
//   internal/interfaces      HTTP 协议转换层
//
// 配置来自服务目录下的 config.yaml，环境变量可覆盖同名字段（见
// internal/infrastructure/config.go），本地开发不改配置直接跑就行。
//
// 详见 docs/Agent开放平台_技术规格文档.md 第三章、6.2 节。
package main

import (
	"log"
	"net/http"

	"github.com/agentmesh/marketplace-service/internal/application"
	"github.com/agentmesh/marketplace-service/internal/infrastructure"
	"github.com/agentmesh/marketplace-service/internal/interfaces"
)

func main() {
	cfg, err := infrastructure.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := infrastructure.NewPostgresConnection(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}

	repo := infrastructure.NewCapabilityMySQLRepository(db)
	registry := infrastructure.NewMCPRegistryAdapter()

	if err := infrastructure.SeedBuiltinCapabilities(repo); err != nil {
		log.Fatalf("failed to seed builtin capabilities: %v", err)
	}

	publishUseCase := application.NewPublishCapabilityUseCase(repo, registry)
	discoverUseCase := application.NewDiscoverCapabilityUseCase(repo, registry)
	handler := interfaces.NewCapabilityHTTPHandler(publishUseCase, discoverUseCase)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /capabilities", handler.ListCapabilities)
	mux.HandleFunc("GET /capabilities/{id}", handler.GetCapability)
	mux.HandleFunc("POST /capabilities/{id}/submit", handler.SubmitCapability)
	mux.HandleFunc("POST /capabilities/{id}/approve", handler.ApproveCapability)

	log.Printf("marketplace-service listening on %s\n", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
