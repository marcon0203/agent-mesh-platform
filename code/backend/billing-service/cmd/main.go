// billing-service 异步消费用量事件，按 Agent / 能力维度聚合计费与分成。
// 用量事件由 orchestration-service 在每次运行结束后写入 asynq 任务队列（Redis），
// 避免同步计费拖慢主链路。详见技术规格文档 6.6 节。
//
// 简单三层结构（见 code/backend/README.md）：
//
//	internal/handler     对外用量查询 HTTP 接口
//	internal/service     asynq 任务消费者
//	internal/repository  usage_record / agent_run 数据访问
//
// 配置来自服务目录下的 config.yaml，环境变量可覆盖同名字段（见
// internal/repository/config.go），本地开发不改配置直接跑就行。
//
// 分成结算（按 capability_id 聚合生成结算单）属于 M2 范围，这里先跑通
// "消费用量事件 → 落库 → 可查询"的链路。
package main

import (
	"log"
	"net/http"

	"github.com/agentmesh/billing-service/internal/handler"
	"github.com/agentmesh/billing-service/internal/repository"
	"github.com/agentmesh/billing-service/internal/service"
)

func main() {
	cfg, err := repository.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := repository.NewPostgresConnection(cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	usageRepo := repository.NewUsageRepository(db)

	consumer := service.NewUsageConsumer(cfg.RedisAddr, usageRepo)
	go func() {
		log.Println("billing-service: consuming usage events")
		if err := consumer.Run(); err != nil {
			log.Fatalf("usage consumer stopped: %v", err)
		}
	}()

	usageHandler := handler.NewUsageHandler(usageRepo)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /usage", usageHandler.GetUsageByAgent)

	log.Printf("billing-service listening on %s (usage query)\n", cfg.HTTPAddr)
	log.Fatal(http.ListenAndServe(cfg.HTTPAddr, mux))
}
