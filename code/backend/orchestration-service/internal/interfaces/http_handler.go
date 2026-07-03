package interfaces

import (
	"encoding/json"
	"net/http"

	"github.com/agentmesh/orchestration-service/internal/application"
	"github.com/agentmesh/orchestration-service/internal/domain"
)

// AdminHTTPHandler 是 orchestration-service 的配置面 HTTP 接口：Agent 能力装配、
// 发布，以及模型供应商的增删查。这是内部/控制台使用的接口，不同于面向最终调用方的
// gRPC Invoke（数据面），也不同于 gateway-service 对外暴露的 sendMsg/streamMsg。
type AdminHTTPHandler struct {
	CreateAgent    *application.CreateAgentUseCase
	ConfigureAgent *application.ConfigureAgentUseCase
	ModelProviders *application.ModelProviderUseCase
}

func NewAdminHTTPHandler(createAgent *application.CreateAgentUseCase, configureAgent *application.ConfigureAgentUseCase, modelProviders *application.ModelProviderUseCase) *AdminHTTPHandler {
	return &AdminHTTPHandler{CreateAgent: createAgent, ConfigureAgent: configureAgent, ModelProviders: modelProviders}
}

type createAgentRequest struct {
	Name         string              `json:"name"`
	LoopTemplate domain.LoopTemplate `json:"loop_template"`
}

type agentResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	LoopTemplate domain.LoopTemplate `json:"loop_template"`
	Status       string              `json:"status"`
}

// CreateAgentHandler 对应 POST /agents，创建一个草稿态 Agent 供后续装配/发布使用。
func (h *AdminHTTPHandler) CreateAgentHandler(w http.ResponseWriter, r *http.Request) {
	var req createAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	agent, err := h.CreateAgent.Execute(application.CreateAgentCommand{Name: req.Name, LoopTemplate: req.LoopTemplate})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agentResponse{ID: agent.ID(), Name: agent.Name(), LoopTemplate: agent.LoopTemplate(), Status: string(agent.Status())})
}

// ListAgents 对应 GET /agents，不带 status 参数时返回所有状态的 Agent
// （供前端 Agent 列表页展示，含草稿态），status=published 时只返回已发布的
// （供 Agent 构建器挑选"挂载哪个已发布 Agent 作为 Subagent"时使用）。
func (h *AdminHTTPHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	var (
		agents []*domain.Agent
		err    error
	)
	if r.URL.Query().Get("status") == "published" {
		agents, err = h.ConfigureAgent.Agents.ListPublished()
	} else {
		agents, err = h.ConfigureAgent.Agents.ListAll()
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]agentResponse, 0, len(agents))
	for _, agent := range agents {
		resp = append(resp, agentResponse{ID: agent.ID(), Name: agent.Name(), LoopTemplate: agent.LoopTemplate(), Status: string(agent.Status())})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

type agentDetailResponse struct {
	ID              string                     `json:"id"`
	Name            string                     `json:"name"`
	LoopTemplate    domain.LoopTemplate        `json:"loop_template"`
	Capabilities    []domain.MountedCapability `json:"capabilities"`
	Hooks           domain.HookConfig          `json:"hooks"`
	MaxDepth        int32                      `json:"max_depth"`
	Status          string                     `json:"status"`
	ModelProviderID string                     `json:"model_provider_id"`
}

// GetAgent 对应 GET /agents/{id}，供 AgentBuilderPage 回显当前配置状态。
func (h *AdminHTTPHandler) GetAgent(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	agent, err := h.ConfigureAgent.Agents.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(agentDetailResponse{
		ID: agent.ID(), Name: agent.Name(), LoopTemplate: agent.LoopTemplate(),
		Capabilities: agent.Capabilities(), Hooks: agent.Hooks(), MaxDepth: agent.MaxDepth(),
		Status: string(agent.Status()), ModelProviderID: agent.ModelProviderID(),
	})
}

type configureAgentRequest struct {
	Capabilities    []domain.MountedCapability `json:"capabilities"`
	MaxDepth        int32                      `json:"max_depth"`
	Hooks           domain.HookConfig          `json:"hooks"`
	ModelProviderID string                     `json:"model_provider_id"`
	Publish         bool                       `json:"publish"`
}

// ConfigureAgentHandler 对应 POST /agents/{id}/config，供前端 AgentBuilderPage 的
// "装配能力 + 选择模型供应商 + 发布" 操作使用。
func (h *AdminHTTPHandler) ConfigureAgentHandler(w http.ResponseWriter, r *http.Request) {
	agentID := r.PathValue("id")
	var req configureAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	cmd := application.ConfigureAgentCommand{
		AgentID:         agentID,
		Capabilities:    req.Capabilities,
		MaxDepth:        req.MaxDepth,
		Hooks:           req.Hooks,
		ModelProviderID: req.ModelProviderID,
		Publish:         req.Publish,
	}
	if err := h.ConfigureAgent.Execute(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type createModelProviderRequest struct {
	OwnerID      string              `json:"owner_id"`
	Name         string              `json:"name"`
	ProviderType domain.ProviderType `json:"provider_type"`
	BaseURL      string              `json:"base_url"`
	APIKey       string              `json:"api_key"`
	ModelName    string              `json:"model_name"`
}

// modelProviderResponse 特意不包含 APIKey 字段——供应商配置一旦创建，
// 前端后续查询列表/详情时不应该再看到明文 Key。
type modelProviderResponse struct {
	ID           string              `json:"id"`
	Name         string              `json:"name"`
	ProviderType domain.ProviderType `json:"provider_type"`
	BaseURL      string              `json:"base_url"`
	ModelName    string              `json:"model_name"`
	Status       string              `json:"status"`
}

func toModelProviderResponse(p *domain.ModelProvider) modelProviderResponse {
	return modelProviderResponse{
		ID:           p.ID(),
		Name:         p.Name(),
		ProviderType: p.Type(),
		BaseURL:      p.BaseURL(),
		ModelName:    p.ModelName(),
		Status:       string(p.Status()),
	}
}

// CreateModelProvider 对应 POST /model-providers。
func (h *AdminHTTPHandler) CreateModelProvider(w http.ResponseWriter, r *http.Request) {
	var req createModelProviderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	provider, err := h.ModelProviders.Create(application.CreateModelProviderCommand{
		OwnerID:      req.OwnerID,
		Name:         req.Name,
		ProviderType: req.ProviderType,
		BaseURL:      req.BaseURL,
		APIKey:       req.APIKey,
		ModelName:    req.ModelName,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toModelProviderResponse(provider))
}

// ListModelProviders 对应 GET /model-providers?owner_id=xxx。
func (h *AdminHTTPHandler) ListModelProviders(w http.ResponseWriter, r *http.Request) {
	ownerID := r.URL.Query().Get("owner_id")
	providers, err := h.ModelProviders.ListByOwner(ownerID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]modelProviderResponse, 0, len(providers))
	for _, p := range providers {
		resp = append(resp, toModelProviderResponse(p))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
