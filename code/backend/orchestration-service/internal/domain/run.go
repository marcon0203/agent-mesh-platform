package domain

// RunContext 是一次调用在领域层的上下文表示，贯穿 Hub 与递归的 Subagent 调用。
type RunContext struct {
	AgentID   string
	SessionID string
	TraceID   string
	Depth     int32
}

// InvokeChunk 是领域层对"一次流式输出片段"的抽象，
// 具体如何转成 gRPC/SSE 帧是 interfaces 层的职责。
type InvokeChunk struct {
	Content string
	IsFinal bool
	Tokens  int32
}

// AgentRepository 是领域层定义、infrastructure 层实现的仓储接口。
// 领域层只依赖接口，不知道背后是 MySQL 还是别的存储。
// Save 返回持久化后的聚合根：新建（ID 为空）时由存储层生成 ID 并通过返回值回填，
// 因为聚合根本身不暴露 ID 的 setter。
type AgentRepository interface {
	FindByID(id string) (*Agent, error)
	Save(agent *Agent) (*Agent, error)
}

// UsageReporter 是领域事件的一种简化表达：一次调用结束后需要上报用量，
// 具体如何上报（同步 HTTP 还是异步消息队列）由 infrastructure 层决定。
type UsageReporter interface {
	Report(ctx RunContext, tokens int32) error
}
