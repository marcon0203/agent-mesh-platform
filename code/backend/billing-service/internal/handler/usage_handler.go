package handler

import (
	"encoding/json"
	"net/http"

	"github.com/agentmesh/billing-service/internal/repository"
)

// UsageHandler 暴露对外用量查询接口（技术规格文档模块划分里
// billing-service 的"对外用量查询接口"职责）。分成结算等更复杂的聚合
// 属于 M2 范围，这里先提供最基础的按 Agent 汇总。
type UsageHandler struct {
	Repo *repository.UsageRepository
}

func NewUsageHandler(repo *repository.UsageRepository) *UsageHandler {
	return &UsageHandler{Repo: repo}
}

// GetUsageByAgent 对应 GET /usage?agent_id=xxx。
func (h *UsageHandler) GetUsageByAgent(w http.ResponseWriter, r *http.Request) {
	agentID := r.URL.Query().Get("agent_id")
	if agentID == "" {
		http.Error(w, "agent_id is required", http.StatusBadRequest)
		return
	}
	tokens, err := h.Repo.SumTokensByAgent(agentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"agent_id": agentID, "tokens": tokens})
}
