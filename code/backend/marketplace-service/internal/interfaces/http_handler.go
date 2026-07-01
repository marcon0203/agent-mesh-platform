package interfaces

import (
	"encoding/json"
	"net/http"

	"github.com/agentmesh/marketplace-service/internal/application"
	"github.com/agentmesh/marketplace-service/internal/domain"
)

type CapabilityHTTPHandler struct {
	Publish  *application.PublishCapabilityUseCase
	Discover *application.DiscoverCapabilityUseCase
}

func NewCapabilityHTTPHandler(publish *application.PublishCapabilityUseCase, discover *application.DiscoverCapabilityUseCase) *CapabilityHTTPHandler {
	return &CapabilityHTTPHandler{Publish: publish, Discover: discover}
}

// capabilityResponse 是对外的 DTO。domain.Capability 的字段都是私有的（聚合根
// 只读暴露 getter），直接 json.Marshal 聚合根指针只会得到 {}，所以这里显式映射。
type capabilityResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	Version     string `json:"version"`
	IsBuiltin   bool   `json:"is_builtin"`
	SchemaJSON  string `json:"schema_json"`
	MCPEndpoint string `json:"mcp_endpoint"`
}

func toCapabilityResponse(c *domain.Capability) capabilityResponse {
	return capabilityResponse{
		ID:          c.ID(),
		Name:        c.Name(),
		Type:        string(c.Type()),
		Status:      string(c.Status()),
		Version:     c.Version(),
		IsBuiltin:   c.IsBuiltin(),
		SchemaJSON:  c.SchemaJSON(),
		MCPEndpoint: c.MCPEndpoint(),
	}
}

// ListCapabilities 对应 GET /capabilities?type=tool
func (h *CapabilityHTTPHandler) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	capType := domain.CapabilityType(r.URL.Query().Get("type"))
	caps, err := h.Discover.ListByType(capType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := make([]capabilityResponse, 0, len(caps))
	for _, c := range caps {
		resp = append(resp, toCapabilityResponse(c))
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

// GetCapability 对应 GET /capabilities/{id}，供市场详情页使用。
func (h *CapabilityHTTPHandler) GetCapability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cap, err := h.Discover.FindByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(toCapabilityResponse(cap))
}

// SubmitCapability 对应 POST /capabilities/{id}/submit
func (h *CapabilityHTTPHandler) SubmitCapability(w http.ResponseWriter, r *http.Request) {
	var cmd application.SubmitCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	cmd.CapabilityID = r.PathValue("id")
	if err := h.Publish.Submit(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}

type approveCapabilityRequest struct {
	Version string `json:"version"`
}

// ApproveCapability 对应 POST /capabilities/{id}/approve，
// 覆盖产品规格文档 §5.1 发布审核流程里"审核通过 → 上架"的一步。
func (h *CapabilityHTTPHandler) ApproveCapability(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var req approveCapabilityRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Publish.Approve(id, req.Version); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
