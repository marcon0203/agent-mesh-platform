package application

import "github.com/agentmesh/orchestration-service/internal/domain"

// CreateAgentUseCase 对应前端 AgentBuilderPage 的第一步：创建一个草稿态 Agent，
// 后续能力装配/模型供应商选择/发布都由 ConfigureAgentUseCase 在这个 Agent 上进行。
type CreateAgentUseCase struct {
	Agents domain.AgentRepository
}

func NewCreateAgentUseCase(agents domain.AgentRepository) *CreateAgentUseCase {
	return &CreateAgentUseCase{Agents: agents}
}

type CreateAgentCommand struct {
	Name         string
	LoopTemplate domain.LoopTemplate
}

func (uc *CreateAgentUseCase) Execute(cmd CreateAgentCommand) (*domain.Agent, error) {
	agent, err := domain.NewAgent("", cmd.Name, cmd.LoopTemplate)
	if err != nil {
		return nil, err
	}
	return uc.Agents.Save(agent)
}
