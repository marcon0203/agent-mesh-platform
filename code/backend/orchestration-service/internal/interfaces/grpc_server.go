package interfaces

import (
	"github.com/agentmesh/orchestration-service/internal/application"
	"github.com/agentmesh/orchestration-service/internal/domain"
)

// OrchestrationGRPCServer 对应 shared/proto/orchestration.proto 里定义的
// OrchestrationService。这一层只负责 gRPC 消息 <-> 用例入参/出参 的转换，
// 不包含任何业务规则 —— 规则已经在 domain.Agent 和 application 用例里了。
type OrchestrationGRPCServer struct {
	InvokeUseCase *application.InvokeUseCase
}

func NewOrchestrationGRPCServer(invoke *application.InvokeUseCase) *OrchestrationGRPCServer {
	return &OrchestrationGRPCServer{InvokeUseCase: invoke}
}

// InvokeRequest/InvokeChunk 字段对应 proto 定义，这里用 Go struct 示意，
// 实际生成代码由 protoc 产出后替换。
type InvokeRequest struct {
	AgentID   string
	SessionID string
	Message   string
	Depth     int32
	TraceID   string
}

func (s *OrchestrationGRPCServer) Invoke(req *InvokeRequest, streamOut chan<- domain.InvokeChunk) error {
	cmd := application.InvokeCommand{
		AgentID:   req.AgentID,
		SessionID: req.SessionID,
		Message:   req.Message,
		Depth:     req.Depth,
		TraceID:   req.TraceID,
	}
	return s.InvokeUseCase.Execute(cmd, streamOut)
}
