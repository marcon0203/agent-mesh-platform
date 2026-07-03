package infrastructure

import (
	"context"
	"fmt"
	"log"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"

	"github.com/cloudwego/eino-ext/components/model/claude"
	"github.com/cloudwego/eino-ext/components/model/openai"

	"github.com/agentmesh/orchestration-service/internal/domain"
)

// defaultClaudeMaxTokens 是 Anthropic API 的必填参数，MVP 阶段用固定值，
// 后续如需按 Agent 精细调整可以加到 ModelProvider 聚合根里。
const defaultClaudeMaxTokens = 4096

// EinoRuntime 实现 domain.AgentRuntime 端口，是整个服务里唯一直接
// 依赖 Eino / Eino ADK 具体 API 的地方。领域层和用例层完全不 import Eino 的包，
// 这样即便未来更换执行引擎，改动范围也只在这一个文件。
type EinoRuntime struct {
	ModelProviders domain.ModelProviderRepository
	Agents         domain.AgentRepository
}

func NewEinoRuntime(modelProviders domain.ModelProviderRepository, agents domain.AgentRepository) *EinoRuntime {
	return &EinoRuntime{ModelProviders: modelProviders, Agents: agents}
}

// buildChatModel 把 ModelProvider 聚合根翻译成具体的 Eino ChatModel 组件实例。
// 【需人工决策】的落地结果：MVP 支持 OpenAI 兼容协议（覆盖大多数可自建 base_url
// 的模型服务）和 Anthropic 原生协议两种，具体选哪种由用户在模型供应商配置页决定。
func buildChatModel(ctx context.Context, provider *domain.ModelProvider) (model.ToolCallingChatModel, error) {
	switch provider.Type() {
	case domain.ProviderTypeOpenAICompatible:
		return openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  provider.APIKey(),
			BaseURL: provider.BaseURL(),
			Model:   provider.ModelName(),
		})
	case domain.ProviderTypeAnthropic:
		cfg := &claude.Config{
			APIKey:    provider.APIKey(),
			Model:     provider.ModelName(),
			MaxTokens: defaultClaudeMaxTokens,
		}
		if provider.BaseURL() != "" {
			baseURL := provider.BaseURL()
			cfg.BaseURL = &baseURL
		}
		return claude.NewChatModel(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported model provider type: %s", provider.Type())
	}
}

// maxIterationsForTemplate 把 agent.LoopTemplate() 映射到 Eino ADK 的
// MaxIterations，对应技术规格文档 §6.1：ReAct 允许多轮工具调用循环，
// 确定性 Pipeline 只允许单轮（M2 才会换成 adk.NewSequentialAgent 之类的编排）。
func maxIterationsForTemplate(t domain.LoopTemplate) int {
	switch t {
	case domain.LoopTemplatePlanExecute:
		return 10
	case domain.LoopTemplateDeterministic:
		return 1
	default:
		return 15
	}
}

// Run 把 domain.Agent 的挂载能力、Hook 配置翻译成 Eino ADK 的具体构造，
// 跑一次完整的 Agentic Loop 并把中间结果流式写入 out。
func (r *EinoRuntime) Run(agent *domain.Agent, runCtx domain.RunContext, message string, out chan<- domain.InvokeChunk) error {
	ctx := context.Background()

	chatModelAgent, err := r.buildAgentNode(ctx, agent, runCtx.Depth)
	if err != nil {
		return err
	}

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           chatModelAgent,
		EnableStreaming: true,
	})

	handler := buildHookCallbackHandler(agent.Hooks(), runCtx)
	iterator := runner.Query(ctx, message, adk.WithCallbacks(handler))

	// Eino 的事件流是"逐事件"而非"逐 token"粒度，这里把每个事件当作一个
	// InvokeChunk，最后一个事件到达时才标记 IsFinal，避免过早把中间事件当结尾。
	var pending *domain.InvokeChunk
	flush := func(isLast bool) {
		if pending == nil {
			return
		}
		if isLast {
			pending.IsFinal = true
		}
		out <- *pending
		pending = nil
	}

	for {
		event, ok := iterator.Next()
		if !ok {
			break
		}
		if event.Err != nil {
			return event.Err
		}

		msg, _, err := adk.GetMessage(event)
		if err != nil {
			return err
		}
		flush(false)
		chunk := domain.InvokeChunk{Content: msg.Content}
		if msg.ResponseMeta != nil && msg.ResponseMeta.Usage != nil {
			chunk.Tokens = int32(msg.ResponseMeta.Usage.TotalTokens)
		}
		pending = &chunk
	}
	flush(true)

	return nil
}

// buildAgentNode 递归构建一个 Eino ChatModelAgent：先校验递归深度，再把
// agent.Capabilities() 解析成 ToolsConfig（含更深一层的 Subagent-as-Tool 包装）。
// depth 语义与 gRPC InvokeRequest.depth 一致（技术规格文档 §6.4）：网关对 Hub
// 的首次调用 depth=0，每往下挂一层 Subagent-as-Tool，depth 传给子 Agent 时 +1，
// 子 Agent 自身的 max_depth 一旦被超过就在这里直接拒绝，不会真的去跑模型。
func (r *EinoRuntime) buildAgentNode(ctx context.Context, agent *domain.Agent, depth int32) (adk.Agent, error) {
	if err := agent.ValidateInvokeDepth(depth); err != nil {
		return nil, err
	}
	if agent.ModelProviderID() == "" {
		return nil, fmt.Errorf("agent %s has no model provider configured", agent.ID())
	}
	provider, err := r.ModelProviders.FindByID(agent.ModelProviderID())
	if err != nil {
		return nil, fmt.Errorf("load model provider: %w", err)
	}
	if !provider.IsEnabled() {
		return nil, fmt.Errorf("model provider %s is disabled", provider.ID())
	}

	chatModel, err := buildChatModel(ctx, provider)
	if err != nil {
		return nil, fmt.Errorf("build chat model: %w", err)
	}

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          agentToolName(agent),
		Description:   fmt.Sprintf("由枢络 AgentMesh 平台编排的智能助理「%s」，可作为子 Agent 处理独立子任务。", agent.Name()),
		Instruction:   fmt.Sprintf("你是 %s，一个由枢络 AgentMesh 平台编排的智能助理，请根据用户请求和可用工具完成任务。", agent.Name()),
		Model:         chatModel,
		ToolsConfig:   r.buildToolsConfig(ctx, agent, depth),
		MaxIterations: maxIterationsForTemplate(agent.LoopTemplate()),
	})
}

// agentToolName 把 Agent 转成 Eino 工具调用要求的合法标识符（LLM 侧的 function
// name 一般只允许字母数字下划线短横线）。Agent.Name() 可以是任意中文，
// 不能直接当工具名用，人类可读的名字放进 Description 里给模型看。
func agentToolName(agent *domain.Agent) string {
	return fmt.Sprintf("agent_%s", agent.ID())
}

// buildToolsConfig 遍历 agent.Capabilities()：IsSubagent=false 的能力解析成
// 内置 Tool（builtin_tools.go 里的静态注册表）；IsSubagent=true 的能力
// 递归包装成 Subagent-as-Tool（技术规格文档 §6.1）。
// 第三方 MCP 能力接入属于 M2 的另一个任务，这里先跳过，避免因为个别未接入的
// 能力导致整个 Agent 无法运行；单个 Subagent 装配失败（比如目标 Agent 已下线、
// 递归深度超限）也只是跳过它，不影响其余能力正常挂载。
func (r *EinoRuntime) buildToolsConfig(ctx context.Context, agent *domain.Agent, depth int32) adk.ToolsConfig {
	cfg := adk.ToolsConfig{}
	for _, cap := range agent.Capabilities() {
		if cap.IsSubagent {
			subTool, err := r.buildSubagentTool(ctx, cap, depth)
			if err != nil {
				log.Printf("skip subagent capability %s for agent %s: %v", cap.CapabilityID, agent.ID(), err)
				continue
			}
			cfg.ToolsNodeConfig.Tools = append(cfg.ToolsNodeConfig.Tools, subTool)
			continue
		}
		if t, ok := builtinTools[cap.CapabilityID]; ok {
			cfg.ToolsNodeConfig.Tools = append(cfg.ToolsNodeConfig.Tools, t)
		}
	}
	return cfg
}

// buildSubagentTool 把一个 IsSubagent=true 的挂载能力包装成 Eino BaseTool：
// CapabilityID 直接引用本服务内另一个已发布 Agent 的 ID（内部递归调用本服务，
// 不经过 marketplace-service），Hub 的 LLM 在 tool_call 阶段像调用普通 Tool
// 一样调用它。用 adk.NewAgentTool 默认行为（不加 WithFullChatHistoryAsInput），
// 子 Agent 只收到精简后的任务描述，不会看到 Hub 完整的对话历史。
func (r *EinoRuntime) buildSubagentTool(ctx context.Context, cap domain.MountedCapability, parentDepth int32) (tool.BaseTool, error) {
	target, err := r.Agents.FindByID(cap.CapabilityID)
	if err != nil {
		return nil, fmt.Errorf("load subagent %s: %w", cap.CapabilityID, err)
	}
	if target.Status() != domain.AgentStatusPublished {
		return nil, fmt.Errorf("subagent %s is not published", cap.CapabilityID)
	}

	subAgent, err := r.buildAgentNode(ctx, target, parentDepth+1)
	if err != nil {
		return nil, fmt.Errorf("build subagent %s: %w", cap.CapabilityID, err)
	}
	return adk.NewAgentTool(ctx, subAgent), nil
}

// buildHookCallbackHandler 用 Eino Callback 的 OnStart/OnEnd 挂载
// agent.Hooks() 里启用的前处理/后处理逻辑，对应技术规格文档 §6.1。
func buildHookCallbackHandler(hooks domain.HookConfig, runCtx domain.RunContext) callbacks.Handler {
	return callbacks.NewHandlerBuilder().
		OnStartFn(func(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
			for _, name := range hooks.PreHooks {
				if fn, ok := PreHookRegistry[name]; ok {
					_ = fn(runCtx)
				}
			}
			return ctx
		}).
		OnEndFn(func(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
			for _, name := range hooks.PostHooks {
				if fn, ok := PostHookRegistry[name]; ok {
					_ = fn(runCtx)
				}
			}
			return ctx
		}).
		Build()
}
