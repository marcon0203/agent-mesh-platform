package application

import "github.com/agentmesh/orchestration-service/internal/domain"

// ConfigureAgentUseCase 对应前端 AgentBuilderPage 的"能力装配"操作，
// 产品规格文档 §4.1 强调这一步是静态配置，不涉及执行顺序 —— 所以这里
// 没有编排逻辑，只是把请求参数转成对聚合根方法的调用。
type ConfigureAgentUseCase struct {
	Agents domain.AgentRepository
}

func NewConfigureAgentUseCase(agents domain.AgentRepository) *ConfigureAgentUseCase {
	return &ConfigureAgentUseCase{Agents: agents}
}

type ConfigureAgentCommand struct {
	AgentID         string
	Capabilities    []domain.MountedCapability
	MaxDepth        int32
	Hooks           domain.HookConfig
	ModelProviderID string
	Publish         bool
}

func (uc *ConfigureAgentUseCase) Execute(cmd ConfigureAgentCommand) error {
	agent, err := uc.Agents.FindByID(cmd.AgentID)
	if err != nil {
		return err
	}

	for _, cap := range cmd.Capabilities {
		agent.MountCapability(cap)
	}
	if err := agent.SetMaxDepth(cmd.MaxDepth); err != nil {
		return err
	}
	agent.EnableHooks(cmd.Hooks)

	if cmd.ModelProviderID != "" {
		if err := agent.SetModelProvider(cmd.ModelProviderID); err != nil {
			return err
		}
	}

	if cmd.Publish {
		if err := agent.Publish(); err != nil {
			return err
		}
	}

	_, err = uc.Agents.Save(agent)
	return err
}
