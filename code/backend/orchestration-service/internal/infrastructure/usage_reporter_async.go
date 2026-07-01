package infrastructure

import "github.com/agentmesh/orchestration-service/internal/domain"

// AsyncUsageReporter 实现 domain.UsageReporter，把用量事件写入消息队列，
// 由 billing-service 异步消费，避免同步计费拖慢主链路。
// 详见 docs/Agent开放平台_技术规格文档.md 6.6 节。
type AsyncUsageReporter struct {
	// TODO: 注入消息队列 producer
}

func NewAsyncUsageReporter() *AsyncUsageReporter {
	return &AsyncUsageReporter{}
}

func (r *AsyncUsageReporter) Report(ctx domain.RunContext, tokens int32) error {
	// TODO: 发布用量事件到消息队列，payload 包含 AgentID/TraceID/Depth/Tokens
	return nil
}
