package infrastructure

import (
	"github.com/agentmesh/orchestration-service/internal/domain"
)

// EinoRuntime 实现 domain.AgentRuntime 端口，是整个服务里唯一直接
// 依赖 Eino / Eino ADK 具体 API 的地方。领域层和用例层完全不 import Eino 的包，
// 这样即便未来更换执行引擎，改动范围也只在这一个文件。
type EinoRuntime struct {
	// TODO: 注入 Eino ChatModel 客户端配置（模型服务地址、鉴权信息等）
}

func NewEinoRuntime() *EinoRuntime {
	return &EinoRuntime{}
}

// Run 把 domain.Agent 的挂载能力、Hook 配置翻译成 Eino ADK 的具体构造：
//  1. 普通能力（IsSubagent=false）注册为 adk.ToolsConfig 里的 BaseTool
//  2. Subagent 能力（IsSubagent=true）包装成 BaseTool 接口，
//     内部递归调用本服务自身（Depth+1），实现 Subagent-as-Tool 机制
//  3. agent.Hooks() 里启用的 Hook 通过 Eino Callback 的 OnStart/OnEnd 挂载
//  4. agent.LoopTemplate() 决定用 ReAct 还是其他预置范式构建 ChatModelAgent
//
// 详见 docs/Agent开放平台_技术规格文档.md 第六章 6.1 节。
func (r *EinoRuntime) Run(agent *domain.Agent, ctx domain.RunContext, message string, out chan<- domain.InvokeChunk) error {
	// TODO: 用 adk.NewChatModelAgent 构建 Hub Agent
	//   agentCfg := &adk.ChatModelAgentConfig{
	//       Instruction:   ...,
	//       Model:         chatModel,
	//       ToolsConfig:   buildToolsConfig(agent, ctx),
	//       MaxIterations: 15,
	//   }
	//
	// TODO: 用 Eino Callback 注册 agent.Hooks() 里启用的前处理/后处理逻辑

	out <- domain.InvokeChunk{Content: "TODO: 接入 Eino ADK 执行逻辑", IsFinal: true}
	return nil
}

// buildToolsConfig 是关键的翻译函数：遍历 agent.Capabilities()，
// 把 IsSubagent=true 的能力包装成一个会递归调用 EinoRuntime.Run 的 BaseTool。
// 这里是 Subagent-as-Tool 机制在代码层面真正落地的位置。
func (r *EinoRuntime) buildToolsConfig(agent *domain.Agent, ctx domain.RunContext) any {
	// TODO: 实现
	return nil
}
