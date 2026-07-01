package infrastructure

import (
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// UsageReportTaskType 是用量事件的 asynq 任务类型名，orchestration-service（生产者）
// 和 billing-service（消费者）通过这个约定的任务类型名 + JSON payload 结构对接，
// 共用同一个 Redis 实例（无需单独部署消息队列），详见 docs/Agent开放平台_技术规格文档.md §6.6。
const UsageReportTaskType = "usage:report"

// usageEvent 是任务 payload 结构，字段覆盖技术规格文档要求的
// AgentID/TraceID/Depth/Tokens。
type usageEvent struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id"`
	Depth     int32  `json:"depth"`
	Tokens    int32  `json:"tokens"`
}

// AsyncUsageReporter 实现 domain.UsageReporter，把用量事件作为 asynq 任务
// enqueue 到 Redis，由 billing-service 异步消费，避免同步计费拖慢主链路。
// 详见 docs/Agent开放平台_技术规格文档.md 6.6 节。
type AsyncUsageReporter struct {
	client *asynq.Client
}

// NewAsyncUsageReporter 连接 Redis（对应 infra/docker-compose.yml 的 redis 服务）。
func NewAsyncUsageReporter(redisAddr string) *AsyncUsageReporter {
	client := asynq.NewClient(asynq.RedisClientOpt{Addr: redisAddr})
	return &AsyncUsageReporter{client: client}
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
	if _, err := r.client.Enqueue(asynq.NewTask(UsageReportTaskType, body)); err != nil {
		// 用量上报是旁路观测数据，失败不应该影响主调用链路，这里只记录日志。
		log.Printf("failed to enqueue usage event: %v", err)
	}
	return nil
}
