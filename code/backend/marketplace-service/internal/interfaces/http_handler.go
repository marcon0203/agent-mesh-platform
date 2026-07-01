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

// ListCapabilities 对应 GET /capabilities?type=tool
func (h *CapabilityHTTPHandler) ListCapabilities(w http.ResponseWriter, r *http.Request) {
	capType := domain.CapabilityType(r.URL.Query().Get("type"))
	caps, err := h.Discover.ListByType(capType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_ = json.NewEncoder(w).Encode(caps)
}

// SubmitCapability 对应 POST /capabilities/{id}/submit
func (h *CapabilityHTTPHandler) SubmitCapability(w http.ResponseWriter, r *http.Request) {
	var cmd application.SubmitCommand
	if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.Publish.Submit(cmd); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusAccepted)
}
