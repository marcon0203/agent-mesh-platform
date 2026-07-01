package infrastructure

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/model"

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
}

func NewEinoRuntime(modelProviders domain.ModelProviderRepository) *EinoRuntime {
	return &EinoRuntime{ModelProviders: modelProviders}
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

	if agent.ModelProviderID() == "" {
		return fmt.Errorf("agent %s has no model provider configured", agent.ID())
	}
	provider, err := r.ModelProviders.FindByID(agent.ModelProviderID())
	if err != nil {
		return fmt.Errorf("load model provider: %w", err)
	}
	if !provider.IsEnabled() {
		return fmt.Errorf("model provider %s is disabled", provider.ID())
	}

	chatModel, err := buildChatModel(ctx, provider)
	if err != nil {
		return fmt.Errorf("build chat model: %w", err)
	}

	chatModelAgent, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:          agent.Name(),
		Instruction:   fmt.Sprintf("你是 %s，一个由枢络 AgentMesh 平台编排的智能助理，请根据用户请求和可用工具完成任务。", agent.Name()),
		Model:         chatModel,
		ToolsConfig:   r.buildToolsConfig(agent, runCtx),
		MaxIterations: maxIterationsForTemplate(agent.LoopTemplate()),
	})
	if err != nil {
		return fmt.Errorf("build chat model agent: %w", err)
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

// buildToolsConfig 遍历 agent.Capabilities()，把 IsSubagent=false 的能力
// 解析成内置 Tool（builtin_tools.go 里的静态注册表）。
// Subagent-as-Tool 递归包装和第三方 MCP 能力接入属于 M2 范围，这里先跳过，
// 避免因为个别未接入的能力导致整个 Agent 无法运行。
func (r *EinoRuntime) buildToolsConfig(agent *domain.Agent, _ domain.RunContext) adk.ToolsConfig {
	cfg := adk.ToolsConfig{}
	for _, cap := range agent.Capabilities() {
		if cap.IsSubagent {
			continue
		}
		if t, ok := builtinTools[cap.CapabilityID]; ok {
			cfg.ToolsNodeConfig.Tools = append(cfg.ToolsNodeConfig.Tools, t)
		}
	}
	return cfg
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
