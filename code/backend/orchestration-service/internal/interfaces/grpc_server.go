package interfaces

import (
	orchestrationpb "github.com/agentmesh/shared/pkg/orchestration"

	"github.com/agentmesh/orchestration-service/internal/application"
	"github.com/agentmesh/orchestration-service/internal/domain"
)

// OrchestrationGRPCServer 对应 shared/proto/orchestration.proto 里定义的
// OrchestrationService。这一层只负责 gRPC 消息 <-> 用例入参/出参 的转换，
// 不包含任何业务规则 —— 规则已经在 domain.Agent 和 application 用例里了。
type OrchestrationGRPCServer struct {
	orchestrationpb.UnimplementedOrchestrationServiceServer
	InvokeUseCase *application.InvokeUseCase
}

func NewOrchestrationGRPCServer(invoke *application.InvokeUseCase) *OrchestrationGRPCServer {
	return &OrchestrationGRPCServer{InvokeUseCase: invoke}
}

// Invoke 实现 orchestrationpb.OrchestrationServiceServer，把 gRPC 服务端流
// 转成 domain.InvokeChunk 的 channel 消费，交给 InvokeUseCase 执行。
func (s *OrchestrationGRPCServer) Invoke(req *orchestrationpb.InvokeRequest, stream orchestrationpb.OrchestrationService_InvokeServer) error {
	cmd := application.InvokeCommand{
		AgentID:   req.GetAgentId(),
		SessionID: req.GetSessionId(),
		Message:   req.GetMessage(),
		Depth:     req.GetDepth(),
		TraceID:   req.GetTraceId(),
	}

	out := make(chan domain.InvokeChunk)
	errCh := make(chan error, 1)
	go func() {
		errCh <- s.InvokeUseCase.Execute(cmd, out)
		close(out)
	}()

	for chunk := range out {
		pbChunk := &orchestrationpb.InvokeChunk{
			Content: chunk.Content,
			IsFinal: chunk.IsFinal,
			Usage:   &orchestrationpb.UsageInfo{Tokens: chunk.Tokens},
		}
		if err := stream.Send(pbChunk); err != nil {
			return err
		}
	}

	return <-errCh
}
