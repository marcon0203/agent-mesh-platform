package service

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/agentmesh/billing-service/internal/repository"
)

// usageEventsQueue 必须和 orchestration-service 的
// internal/infrastructure/usage_reporter_async.go 里的 UsageEventsQueue 保持一致——
// 两个服务通过这个约定的队列名 + JSON payload 结构对接，没有引入额外的
// 共享 Go 包（用量事件是简单 DTO，不像 gRPC 契约那样需要强类型生成代码）。
const usageEventsQueue = "agentmesh.usage_events"

type usageEvent struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id"`
	Depth     int32  `json:"depth"`
	Tokens    int32  `json:"tokens"`
}

// UsageConsumer 消费用量事件队列，写入 usage_record 表（技术规格文档 §6.6）。
type UsageConsumer struct {
	channel *amqp.Channel
	repo    *repository.UsageRepository
}

func NewUsageConsumer(amqpURL string, repo *repository.UsageRepository) (*UsageConsumer, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare(usageEventsQueue, true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &UsageConsumer{channel: ch, repo: repo}, nil
}

// Run 阻塞式地消费队列，调用方应该在单独的 goroutine 里调用。
func (c *UsageConsumer) Run() error {
	msgs, err := c.channel.Consume(usageEventsQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for delivery := range msgs {
		var evt usageEvent
		if err := json.Unmarshal(delivery.Body, &evt); err != nil {
			log.Printf("usage consumer: drop malformed message: %v", err)
			_ = delivery.Nack(false, false)
			continue
		}

		agentRunID, err := c.repo.UpsertAgentRun(evt.AgentID, evt.TraceID, evt.Tokens)
		if err != nil {
			log.Printf("usage consumer: upsert agent_run failed, will retry: %v", err)
			_ = delivery.Nack(false, true)
			continue
		}
		if err := c.repo.InsertUsageRecord(agentRunID, evt.Tokens); err != nil {
			log.Printf("usage consumer: insert usage_record failed, will retry: %v", err)
			_ = delivery.Nack(false, true)
			continue
		}

		_ = delivery.Ack(false)
	}
	return nil
}
