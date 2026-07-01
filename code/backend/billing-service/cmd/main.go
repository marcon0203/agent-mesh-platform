// billing-service 异步消费用量事件，按 Agent / 能力维度聚合计费与分成。
// 用量事件由 orchestration-service 在每次运行结束后写入消息队列，
// 避免同步计费拖慢主链路。详见技术规格文档 6.6 节。
//
// 简单三层结构（见 code/backend/README.md）：
//
//	internal/handler     对外用量查询 HTTP 接口
//	internal/service     消息队列消费者
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

func mysqlDSN() string {
	if dsn := os.Getenv("BILLING_MYSQL_DSN"); dsn != "" {
		return dsn
	}
	return "root:agentmesh@tcp(127.0.0.1:3306)/agentmesh?parseTime=true"
}

func amqpURL() string {
	if url := os.Getenv("BILLING_AMQP_URL"); url != "" {
		return url
	}
	return "amqp://guest:guest@127.0.0.1:5672/"
}

func main() {
	db, err := repository.NewMySQLConnection(mysqlDSN())
	if err != nil {
		log.Fatalf("failed to connect to mysql: %v", err)
	}
	usageRepo := repository.NewUsageRepository(db)

	consumer, err := service.NewUsageConsumer(amqpURL(), usageRepo)
	if err != nil {
		log.Fatalf("failed to connect to rabbitmq: %v", err)
	}
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
