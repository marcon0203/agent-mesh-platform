package service

import (
	"context"
	"encoding/json"
	"log"

	"github.com/hibiken/asynq"

	"github.com/agentmesh/billing-service/internal/repository"
)

// usageReportTaskType 必须和 orchestration-service 的
// internal/infrastructure/usage_reporter_async.go 里的 UsageReportTaskType 保持一致——
// 两个服务通过这个约定的任务类型名 + JSON payload 结构对接，共用同一个 Redis 实例
// （用量事件是简单 DTO，不像 gRPC 契约那样需要强类型生成代码，没有引入额外的共享 Go 包）。
const usageReportTaskType = "usage:report"

type usageEvent struct {
	AgentID   string `json:"agent_id"`
	SessionID string `json:"session_id"`
	TraceID   string `json:"trace_id"`
	Depth     int32  `json:"depth"`
	Tokens    int32  `json:"tokens"`
}

// UsageConsumer 消费用量事件任务，写入 usage_record 表（技术规格文档 §6.6）。
type UsageConsumer struct {
	server *asynq.Server
	repo   *repository.UsageRepository
}

func NewUsageConsumer(redisAddr string, repo *repository.UsageRepository) *UsageConsumer {
	server := asynq.NewServer(
		asynq.RedisClientOpt{Addr: redisAddr},
		asynq.Config{Concurrency: 5},
	)
	return &UsageConsumer{server: server, repo: repo}
}

// Run 阻塞式地消费任务队列，调用方应该在单独的 goroutine 里调用。
func (c *UsageConsumer) Run() error {
	mux := asynq.NewServeMux()
	mux.HandleFunc(usageReportTaskType, c.handleUsageReport)
	return c.server.Run(mux)
}

func (c *UsageConsumer) handleUsageReport(_ context.Context, task *asynq.Task) error {
	var evt usageEvent
	if err := json.Unmarshal(task.Payload(), &evt); err != nil {
		log.Printf("usage consumer: drop malformed task: %v", err)
		return nil // 不重试：payload 本身就是坏的，重试也不会变好
	}

	agentRunID, err := c.repo.UpsertAgentRun(evt.AgentID, evt.TraceID, evt.Tokens)
	if err != nil {
		log.Printf("usage consumer: upsert agent_run failed, will retry: %v", err)
		return err // 返回 err 让 asynq 按重试策略重新入队
	}
	if err := c.repo.InsertUsageRecord(agentRunID, evt.Tokens); err != nil {
		log.Printf("usage consumer: insert usage_record failed, will retry: %v", err)
		return err
	}
	return nil
}
