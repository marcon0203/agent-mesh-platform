package application

import (
	"github.com/agentmesh/orchestration-service/internal/domain"
)

// InvokeUseCase 是 gRPC handler 与领域模型之间的编排层。
// 规则本身（深度校验、Hook 是否启用）都在 domain.Agent 里，这里只负责按顺序把步骤串起来。
type InvokeUseCase struct {
	Agents  domain.AgentRepository
	Runtime domain.AgentRuntime
	Usage   domain.UsageReporter
}

func NewInvokeUseCase(agents domain.AgentRepository, runtime domain.AgentRuntime, usage domain.UsageReporter) *InvokeUseCase {
	return &InvokeUseCase{Agents: agents, Runtime: runtime, Usage: usage}
}

type InvokeCommand struct {
	AgentID   string
	SessionID string
	Message   string
	Depth     int32
	TraceID   string
}

// Execute 对应技术规格文档 §6.1 的执行步骤：
//  1. 加载 Agent 聚合根
//  2. 校验递归深度（规则在聚合根内部，这里只是调用）
//  3. 委托 Runtime 端口执行 Agentic Loop（技术细节在 infrastructure/eino_runtime.go）
//  4. 执行结束后异步上报用量
func (uc *InvokeUseCase) Execute(cmd InvokeCommand, out chan<- domain.InvokeChunk) error {
	agent, err := uc.Agents.FindByID(cmd.AgentID)
	if err != nil {
		return err
	}

	if err := agent.ValidateInvokeDepth(cmd.Depth); err != nil {
		return err
	}

	runCtx := domain.RunContext{
		AgentID:   cmd.AgentID,
		SessionID: cmd.SessionID,
		TraceID:   cmd.TraceID,
		Depth:     cmd.Depth,
	}

	// relay 是 Runtime.Run 和对外的 out 之间的中转：每个 chunk 先经过这里
	// 累加 tokens，再转发给调用方，这样上报用量不需要 Runtime 关心计费细节。
	relay := make(chan domain.InvokeChunk)
	pumpDone := make(chan struct{})
	var totalTokens int32
	go func() {
		defer close(pumpDone)
		for chunk := range relay {
			totalTokens += chunk.Tokens
			out <- chunk
		}
	}()

	runErr := uc.Runtime.Run(agent, runCtx, cmd.Message, relay)
	close(relay)
	<-pumpDone

	if runErr != nil {
		return runErr
	}

	go uc.Usage.Report(runCtx, totalTokens)

	return nil
}
