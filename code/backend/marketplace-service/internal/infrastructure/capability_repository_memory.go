package infrastructure

import (
	"errors"
	"sync"

	"github.com/agentmesh/marketplace-service/internal/domain"
)

type InMemoryCapabilityRepository struct {
	mu           sync.RWMutex
	capabilities map[string]*domain.Capability
}

func NewInMemoryCapabilityRepository() *InMemoryCapabilityRepository {
	return &InMemoryCapabilityRepository{capabilities: make(map[string]*domain.Capability)}
}

func (r *InMemoryCapabilityRepository) FindByID(id string) (*domain.Capability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	cap, ok := r.capabilities[id]
	if !ok {
		return nil, errors.New("capability not found: " + id)
	}
	return cap, nil
}

func (r *InMemoryCapabilityRepository) Save(cap *domain.Capability) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.capabilities[cap.ID()] = cap
	return nil
}

func (r *InMemoryCapabilityRepository) ListPublished(capType domain.CapabilityType) ([]*domain.Capability, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []*domain.Capability
	for _, cap := range r.capabilities {
		if cap.Status() == domain.StatusPublished && cap.Type() == capType {
			result = append(result, cap)
		}
	}
	return result, nil
}
