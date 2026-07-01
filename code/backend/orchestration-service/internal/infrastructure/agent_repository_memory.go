package infrastructure

import (
	"errors"
	"sync"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// InMemoryAgentRepository 是 domain.AgentRepository 的临时实现，
// 用于本地开发联调；生产环境替换为对接 MySQL agent 表的实现，
// application 层代码不需要任何改动（这正是依赖倒置的意义）。
type InMemoryAgentRepository struct {
	mu     sync.RWMutex
	agents map[string]*domain.Agent
}

func NewInMemoryAgentRepository() *InMemoryAgentRepository {
	return &InMemoryAgentRepository{agents: make(map[string]*domain.Agent)}
}

func (r *InMemoryAgentRepository) FindByID(id string) (*domain.Agent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	agent, ok := r.agents[id]
	if !ok {
		return nil, errors.New("agent not found: " + id)
	}
	return agent, nil
}

func (r *InMemoryAgentRepository) Save(agent *domain.Agent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.agents[agent.ID()] = agent
	return nil
}
