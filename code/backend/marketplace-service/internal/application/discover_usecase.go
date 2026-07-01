package application

import "github.com/agentmesh/marketplace-service/internal/domain"

type DiscoverCapabilityUseCase struct {
	Repo     domain.CapabilityRepository
	Registry domain.MCPRegistry
}

func NewDiscoverCapabilityUseCase(repo domain.CapabilityRepository, registry domain.MCPRegistry) *DiscoverCapabilityUseCase {
	return &DiscoverCapabilityUseCase{Repo: repo, Registry: registry}
}

// ListByType 供市场检索页 / 智能装配推荐使用。
func (uc *DiscoverCapabilityUseCase) ListByType(capType domain.CapabilityType) ([]*domain.Capability, error) {
	return uc.Repo.ListPublished(capType)
}

// FindByID 供市场详情页使用。
func (uc *DiscoverCapabilityUseCase) FindByID(id string) (*domain.Capability, error) {
	return uc.Repo.FindByID(id)
}

// ResolveEndpoint 供 orchestration-service 在构建 Agent 运行时，
// 根据挂载的 capabilityID 实时拉取 MCP 连接信息。
func (uc *DiscoverCapabilityUseCase) ResolveEndpoint(capabilityID string) (string, error) {
	cap, err := uc.Repo.FindByID(capabilityID)
	if err != nil {
		return "", err
	}
	if cap.IsBuiltin() {
		return "", nil // 内置能力走进程内调用，无需 MCP endpoint
	}
	return uc.Registry.Resolve(capabilityID)
}
