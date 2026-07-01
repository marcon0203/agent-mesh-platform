package application

import "github.com/agentmesh/marketplace-service/internal/domain"

type PublishCapabilityUseCase struct {
	Repo     domain.CapabilityRepository
	Registry domain.MCPRegistry
}

func NewPublishCapabilityUseCase(repo domain.CapabilityRepository, registry domain.MCPRegistry) *PublishCapabilityUseCase {
	return &PublishCapabilityUseCase{Repo: repo, Registry: registry}
}

type SubmitCommand struct {
	CapabilityID string
	SchemaJSON   string
	MCPEndpoint  string // 内置能力可为空
}

// Submit 对应「发布与审核」流程的第一步：定义 schema + 注册 MCP（如需要）+ 提交审核。
// 校验逻辑（schema 是否为空、第三方能力是否提供 endpoint）都在聚合根 SubmitForReview 里，
// 这里只负责把步骤串起来。
func (uc *PublishCapabilityUseCase) Submit(cmd SubmitCommand) error {
	cap, err := uc.Repo.FindByID(cmd.CapabilityID)
	if err != nil {
		return err
	}

	cap.DefineSchema(cmd.SchemaJSON)
	if cmd.MCPEndpoint != "" {
		cap.SetMCPEndpoint(cmd.MCPEndpoint)
		if err := uc.Registry.Register(cap.ID(), cmd.MCPEndpoint); err != nil {
			return err
		}
	}

	if err := cap.SubmitForReview(); err != nil {
		return err
	}
	return uc.Repo.Save(cap)
}

// Approve 对应人工/规则审核通过后的上架操作。
func (uc *PublishCapabilityUseCase) Approve(capabilityID, version string) error {
	cap, err := uc.Repo.FindByID(capabilityID)
	if err != nil {
		return err
	}
	if err := cap.Approve(version); err != nil {
		return err
	}
	return uc.Repo.Save(cap)
}
