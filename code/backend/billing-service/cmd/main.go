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
// 分成结算（按 capability_id 聚合生成结算单）属于 M2 范围，这里先跑通
// "消费用量事件 → 落库 → 可查询"的链路。
package main

import (
	"log"
	"net/http"
	"os"

	"github.com/agentmesh/billing-service/internal/handler"
	"github.com/agentmesh/billing-service/internal/repository"
	"github.com/agentmesh/billing-service/internal/service"
)

func postgresDSN() string {
	if dsn := os.Getenv("BILLING_POSTGRES_DSN"); dsn != "" {
		return dsn
	}
	return "postgres://agentmesh:agentmesh@127.0.0.1:5432/agentmesh?sslmode=disable"
}

func redisAddr() string {
	if addr := os.Getenv("BILLING_REDIS_ADDR"); addr != "" {
		return addr
	}
	return "127.0.0.1:6379"
}

func main() {
	db, err := repository.NewPostgresConnection(postgresDSN())
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	usageRepo := repository.NewUsageRepository(db)

	consumer := service.NewUsageConsumer(redisAddr(), usageRepo)
	go func() {
		log.Println("billing-service: consuming usage events")
		if err := consumer.Run(); err != nil {
			log.Fatalf("usage consumer stopped: %v", err)
		}
	}()

	usageHandler := handler.NewUsageHandler(usageRepo)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /usage", usageHandler.GetUsageByAgent)

	log.Println("billing-service listening on :8083 (usage query)")
	log.Fatal(http.ListenAndServe(":8083", mux))
}
