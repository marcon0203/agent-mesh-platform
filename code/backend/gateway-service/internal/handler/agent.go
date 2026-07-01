package handler

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
)

// SendMsg 同步调用已发布 Agent。
// 请求/响应结构对应技术规格文档第六章的 API Contract。
func SendMsg(ctx context.Context, c *app.RequestContext) {
	agentID := c.Param("agent_id")

	// TODO:
	// 1. 从 c 中解析 request body（message/session_id/context）
	// 2. 校验该 API Key 对该 agentID 的 scope（agent:{id}:invoke）
	// 3. 组装 gRPC InvokeRequest，depth=0，调用 orchestration-service
	// 4. 将结果和 trace_id/usage 一并返回

	c.JSON(200, map[string]any{
		"code": 0,
		"data": map[string]any{
			"reply":    "TODO: 接入 orchestration-service",
			"agent_id": agentID,
		},
	})
}

// StreamMsg 以 SSE 方式流式调用已发布 Agent。
func StreamMsg(ctx context.Context, c *app.RequestContext) {
	// TODO: 建立与 orchestration-service 的 gRPC 流式连接，
	// 将 InvokeChunk 逐个转换为 SSE data: 帧写回客户端。
	c.SetContentType("text/event-stream")
}

// OpenAPISpec 根据 Agent 的输入输出 schema 自动生成 OpenAPI 3.0 文档。
// 骨架 yaml 参考 docs/Agent开放平台_产品规格文档.md 第六章 6.3 节。
func OpenAPISpec(ctx context.Context, c *app.RequestContext) {
	c.JSON(200, map[string]any{"openapi": "3.0.3", "info": map[string]string{"title": "TODO"}})
}
