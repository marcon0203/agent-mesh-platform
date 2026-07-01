package application

import "github.com/agentmesh/orchestration-service/internal/domain"

// ModelProviderUseCase 对应前端"模型供应商配置"页：用户自建 provider（含 API Key），
// Agent 发布时引用某个 provider。加密/解密等技术细节在 infrastructure 层的仓储实现里，
// 这里只编排领域对象的创建与查询。
type ModelProviderUseCase struct {
	Providers domain.ModelProviderRepository
}

func NewModelProviderUseCase(providers domain.ModelProviderRepository) *ModelProviderUseCase {
	return &ModelProviderUseCase{Providers: providers}
}

type CreateModelProviderCommand struct {
	OwnerID      string
	Name         string
	ProviderType domain.ProviderType
	BaseURL      string
	APIKey       string
	ModelName    string
}

// Create 校验规则在聚合根的 NewModelProvider 构造函数里，这里只负责持久化。
// ID 为空表示新建，由仓储层（对应 MySQL 自增列）生成后回填到返回的聚合根。
func (uc *ModelProviderUseCase) Create(cmd CreateModelProviderCommand) (*domain.ModelProvider, error) {
	provider, err := domain.NewModelProvider("", cmd.OwnerID, cmd.Name, cmd.ProviderType, cmd.BaseURL, cmd.APIKey, cmd.ModelName)
	if err != nil {
		return nil, err
	}
	return uc.Providers.Save(provider)
}

func (uc *ModelProviderUseCase) ListByOwner(ownerID string) ([]*domain.ModelProvider, error) {
	return uc.Providers.ListByOwner(ownerID)
}

func (uc *ModelProviderUseCase) FindByID(id string) (*domain.ModelProvider, error) {
	return uc.Providers.FindByID(id)
}

func (uc *ModelProviderUseCase) Disable(id string) error {
	provider, err := uc.Providers.FindByID(id)
	if err != nil {
		return err
	}
	provider.Disable()
	_, err = uc.Providers.Save(provider)
	return err
}
