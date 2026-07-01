package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/google/uuid"

	orchestrationpb "github.com/agentmesh/shared/pkg/orchestration"
)

// AgentHandler 把 gateway 收到的 HTTP/SSE 请求转成到 orchestration-service 的
// gRPC Invoke 调用，是网关"唯一入口"职责里真正做路由转发的一层。
type AgentHandler struct {
	Orchestration orchestrationpb.OrchestrationServiceClient
}

func NewAgentHandler(client orchestrationpb.OrchestrationServiceClient) *AgentHandler {
	return &AgentHandler{Orchestration: client}
}

type sendMsgRequest struct {
	Message   string `json:"message"`
	SessionID string `json:"session_id"`
}

// SendMsg 同步调用已发布 Agent。
// 请求/响应结构对应产品规格文档 §6.4 的 API Contract。
func (h *AgentHandler) SendMsg(ctx context.Context, c *app.RequestContext) {
	agentID := c.Param("agent_id")
	var req sendMsgRequest
	if err := c.BindJSON(&req); err != nil {
		abortWithCode(c, 400, 40000, err.Error())
		return
	}

	traceID := uuid.NewString()
	stream, err := h.Orchestration.Invoke(ctx, &orchestrationpb.InvokeRequest{
		AgentId:   agentID,
		SessionId: req.SessionID,
		Message:   req.Message,
		Depth:     0,
		TraceId:   traceID,
	})
	if err != nil {
		abortWithCode(c, 502, 40001, err.Error())
		return
	}

	var reply strings.Builder
	var totalTokens int32
	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			abortWithCode(c, 502, 40001, err.Error())
			return
		}
		reply.WriteString(chunk.GetContent())
		if usage := chunk.GetUsage(); usage != nil {
			totalTokens += usage.GetTokens()
		}
	}

	c.JSON(200, map[string]any{
		"code": 0,
		"data": map[string]any{
			"reply":    reply.String(),
			"trace_id": traceID,
			"usage":    map[string]any{"tokens": totalTokens},
		},
	})
}

// StreamMsg 以 SSE 方式流式调用已发布 Agent，把 gRPC 服务端流逐个转成
// `data: {...}\n\n` 帧写回客户端，供前端用 EventSource 消费。
func (h *AgentHandler) StreamMsg(ctx context.Context, c *app.RequestContext) {
	agentID := c.Param("agent_id")
	var req sendMsgRequest
	if err := c.BindJSON(&req); err != nil {
		abortWithCode(c, 400, 40000, err.Error())
		return
	}

	traceID := uuid.NewString()
	stream, err := h.Orchestration.Invoke(ctx, &orchestrationpb.InvokeRequest{
		AgentId:   agentID,
		SessionId: req.SessionID,
		Message:   req.Message,
		Depth:     0,
		TraceId:   traceID,
	})
	if err != nil {
		abortWithCode(c, 502, 40001, err.Error())
		return
	}

	c.SetContentType("text/event-stream")
	c.Response.Header.Set("Cache-Control", "no-cache")
	c.Response.Header.Set("Connection", "keep-alive")

	for {
		chunk, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			frame, _ := json.Marshal(map[string]any{"code": 40001, "message": err.Error()})
			fmt.Fprintf(c, "event: error\ndata: %s\n\n", frame)
			_ = c.Flush()
			return
		}

		payload := map[string]any{
			"reply":    chunk.GetContent(),
			"trace_id": traceID,
			"is_final": chunk.GetIsFinal(),
		}
		if usage := chunk.GetUsage(); usage != nil {
			payload["usage"] = map[string]any{"tokens": usage.GetTokens()}
		}
		data, _ := json.Marshal(payload)
		fmt.Fprintf(c, "data: %s\n\n", data)
		_ = c.Flush()
	}
}

// OpenAPISpec 根据统一的 API Contract 骨架生成该 Agent 的 OpenAPI 3.0 文档，
// 骨架参考产品规格文档 §6.3。sendMsg 的请求/响应体对所有 Agent 都是同一套
// message/session_id/context 通用契约（Agent 本身不像 Tool/Skill 那样声明
// 自定义 schema），这里动态填充的是 agent_id/title；等 orchestration-service
// 提供 Agent 详情查询接口后，可以把 title 换成 Agent 的真实名称。
func OpenAPISpec(ctx context.Context, c *app.RequestContext) {
	agentID := c.Param("agent_id")
	basePath := fmt.Sprintf("/agents/%s", agentID)

	spec := map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":   fmt.Sprintf("Agent %s API", agentID),
			"version": "1.0.0",
		},
		"servers": []map[string]any{
			{"url": "/api/v1"},
		},
		"security": []map[string]any{
			{"ApiKeyAuth": []string{}},
		},
		"paths": map[string]any{
			basePath + "/sendMsg": map[string]any{
				"post": map[string]any{
					"summary": "向该 Agent 发送一条消息（同步）",
					"requestBody": map[string]any{
						"required": true,
						"content": map[string]any{
							"application/json": map[string]any{
								"schema": map[string]any{
									"type": "object",
									"properties": map[string]any{
										"message":    map[string]any{"type": "string", "description": "用户输入"},
										"session_id": map[string]any{"type": "string"},
										"context":    map[string]any{"type": "object"},
									},
									"required": []string{"message"},
								},
							},
						},
					},
					"responses": map[string]any{
						"200": map[string]any{
							"description": "成功",
							"content": map[string]any{
								"application/json": map[string]any{
									"schema": map[string]any{
										"type": "object",
										"properties": map[string]any{
											"code": map[string]any{"type": "integer"},
											"data": map[string]any{
												"type": "object",
												"properties": map[string]any{
													"reply":    map[string]any{"type": "string"},
													"trace_id": map[string]any{"type": "string"},
													"usage": map[string]any{
														"type": "object",
														"properties": map[string]any{
															"tokens": map[string]any{"type": "integer"},
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"429": map[string]any{"description": "超出限流配额"},
						"401": map[string]any{"description": "鉴权失败"},
					},
				},
			},
			basePath + "/streamMsg": map[string]any{
				"post": map[string]any{
					"summary": "向该 Agent 发送一条消息（SSE 流式）",
					"responses": map[string]any{
						"200": map[string]any{"description": "text/event-stream 流式返回，chunk 结构同 sendMsg 的 data 字段"},
					},
				},
			},
		},
		"components": map[string]any{
			"securitySchemes": map[string]any{
				"ApiKeyAuth": map[string]any{
					"type": "apiKey",
					"in":   "header",
					"name": "Authorization",
				},
			},
		},
	}

	c.JSON(200, spec)
}
