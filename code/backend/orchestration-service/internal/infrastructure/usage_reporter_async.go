package infrastructure

import (
	"encoding/json"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// UsageEventsQueue 是用量事件队列名，orchestration-service（生产者）和
// billing-service（消费者）通过这个约定的队列名 + JSON payload 结构对接，
// 详见 docs/Agent开放平台_技术规格文档.md §6.6。
const UsageEventsQueue = "agentmesh.usage_events"

// usageEvent 是发布到消息队列的 payload 结构，字段覆盖技术规格文档要求的
// AgentID/TraceID/Depth/Tokens。
type usageEvent struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id"`
	Depth     int32  `json:"depth"`
	Tokens    int32  `json:"tokens"`
}

// AsyncUsageReporter 实现 domain.UsageReporter，把用量事件写入消息队列，
// 由 billing-service 异步消费，避免同步计费拖慢主链路。
// 详见 docs/Agent开放平台_技术规格文档.md 6.6 节。
type AsyncUsageReporter struct {
	channel *amqp.Channel
}

// NewAsyncUsageReporter 连接 RabbitMQ 并声明用量事件队列。amqpURL 形如
// "amqp://guest:guest@127.0.0.1:5672/"（对应 infra/docker-compose.yml 的 rabbitmq 服务）。
func NewAsyncUsageReporter(amqpURL string) (*AsyncUsageReporter, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}
	if _, err := ch.QueueDeclare(UsageEventsQueue, true, false, false, false, nil); err != nil {
		return nil, err
	}
	return &AsyncUsageReporter{channel: ch}, nil
}

func (r *AsyncUsageReporter) Report(ctx domain.RunContext, tokens int32) error {
	body, err := json.Marshal(usageEvent{
		AgentID:   ctx.AgentID,
		SessionID: ctx.SessionID,
		TraceID:   ctx.TraceID,
		Depth:     ctx.Depth,
		Tokens:    tokens,
	})
	if err != nil {
		return err
	}
	err = r.channel.Publish("", UsageEventsQueue, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        body,
	})
	if err != nil {
		// 用量上报是旁路观测数据，失败不应该影响主调用链路，这里只记录日志。
		log.Printf("failed to publish usage event: %v", err)
	}
	return nil
}
