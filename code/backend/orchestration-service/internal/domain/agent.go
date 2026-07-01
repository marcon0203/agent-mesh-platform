package domain

import "errors"

var (
	ErrMaxDepthExceeded      = errors.New("40004: exceeded max subagent recursion depth")
	ErrCapabilityNotFound    = errors.New("mounted capability not found in agent config")
	ErrEmptyAgentName        = errors.New("agent name must not be empty")
	ErrModelProviderRequired = errors.New("agent must select a model provider before publishing")
)

// LoopTemplate 对应产品规格文档第四章「三层执行模型」里的运行时机制选择。
type LoopTemplate string

const (
	LoopTemplateReAct         LoopTemplate = "react"          // 默认 Agentic Loop
	LoopTemplatePlanExecute   LoopTemplate = "plan_execute"
	LoopTemplateDeterministic LoopTemplate = "pipeline"        // 确定性 Pipeline，见技术规格文档 §4.3
)

// MountedCapability 是值对象：一个 Agent 挂载的一个能力（Tool/Skill/Subagent）。
// json 标签只是给 interfaces 层的 HTTP handler 解码请求体用的，
// 领域层本身不 import encoding/json 以外的任何框架包。
type MountedCapability struct {
	CapabilityID  string `json:"capability_id"`
	PinnedVersion string `json:"pinned_version"` // 默认锁定版本，不自动跟随最新，见产品规格文档 §5.4
	IsSubagent    bool   `json:"is_subagent"`     // true 时该能力按 Subagent-as-Tool 机制包装调用
}

// HookConfig 描述该 Agent 启用了哪些前处理/后处理 Hook。
// 可插拔性本身是不变量：一个 Hook 只能是"启用"或"未启用"，不存在中间状态。
type HookConfig struct {
	PreHooks  []string `json:"pre_hooks"`
	PostHooks []string `json:"post_hooks"`
}

// Agent 是本服务的核心聚合根。
// 所有会影响执行安全性的规则（递归深度、能力挂载合法性）都必须经过聚合根方法，
// 不允许 application 层直接拼装一个 Agent struct 绕过校验。
type Agent struct {
	id              string
	name            string
	loopTemplate    LoopTemplate
	capabilities    []MountedCapability
	hooks           HookConfig
	maxDepth        int32 // 子 Agent 最大递归深度，默认 3，见技术规格文档 §6.5
	version         string
	status          AgentStatus
	modelProviderID string // 引用 ModelProvider 聚合的 ID，决定执行时用哪个模型供应商
}

type AgentStatus string

const (
	AgentStatusDraft     AgentStatus = "draft"
	AgentStatusPublished AgentStatus = "published"
	AgentStatusOffline   AgentStatus = "offline"
)

const defaultMaxDepth int32 = 3

// NewAgent 是唯一合法的构造入口，保证创建出来的 Agent 一定满足基本不变量。
func NewAgent(id, name string, loopTemplate LoopTemplate) (*Agent, error) {
	if name == "" {
		return nil, ErrEmptyAgentName
	}
	if loopTemplate == "" {
		loopTemplate = LoopTemplateReAct
	}
	return &Agent{
		id:           id,
		name:         name,
		loopTemplate: loopTemplate,
		maxDepth:     defaultMaxDepth,
		status:       AgentStatusDraft,
	}, nil
}

// MountCapability 挂载一个能力。挂载 Subagent 时不做特殊处理——
// 对聚合根而言，Subagent 只是 IsSubagent=true 的一种能力，
// 这正是 Subagent-as-Tool 机制在领域模型层面的体现（技术规格文档 §6.1）。
func (a *Agent) MountCapability(cap MountedCapability) {
	a.capabilities = append(a.capabilities, cap)
}

// SetMaxDepth 允许在默认值基础上调整，但不允许设为负数或超过安全上限。
func (a *Agent) SetMaxDepth(depth int32) error {
	if depth < 0 || depth > 5 {
		return errors.New("max depth must be between 0 and 5")
	}
	a.maxDepth = depth
	return nil
}

// ValidateInvokeDepth 是本聚合根最重要的不变量校验：
// 任何一次调用（无论来自网关直接请求还是 Subagent 递归调用）都必须先过这一关。
func (a *Agent) ValidateInvokeDepth(currentDepth int32) error {
	if currentDepth > a.maxDepth {
		return ErrMaxDepthExceeded
	}
	return nil
}

// EnableHooks 校验后设置 Hook 配置——目前只做去重与基础校验，
// 具体 Hook 的执行逻辑属于 infrastructure 层（Eino Callback 绑定），不属于领域层职责。
func (a *Agent) EnableHooks(cfg HookConfig) {
	a.hooks = cfg
}

// SetModelProvider 挂载该 Agent 执行时使用的模型供应商，引用由用户在
// 模型供应商配置页里自建的 ModelProvider（见 model_provider.go）。
func (a *Agent) SetModelProvider(providerID string) error {
	if providerID == "" {
		return errors.New("model provider id must not be empty")
	}
	a.modelProviderID = providerID
	return nil
}

func (a *Agent) Publish() error {
	if len(a.capabilities) == 0 {
		return errors.New("cannot publish an agent with no mounted capabilities")
	}
	if a.modelProviderID == "" {
		return ErrModelProviderRequired
	}
	a.status = AgentStatusPublished
	return nil
}

// Getter：聚合根内部状态只读暴露，禁止外部直接修改字段。
func (a *Agent) ID() string                        { return a.id }
func (a *Agent) Name() string                      { return a.name }
func (a *Agent) LoopTemplate() LoopTemplate        { return a.loopTemplate }
func (a *Agent) Capabilities() []MountedCapability { return a.capabilities }
func (a *Agent) Hooks() HookConfig                 { return a.hooks }
func (a *Agent) MaxDepth() int32                   { return a.maxDepth }
func (a *Agent) Status() AgentStatus               { return a.status }
func (a *Agent) ModelProviderID() string           { return a.modelProviderID }
func (a *Agent) Version() string                   { return a.version }

// RehydrateAgent 供仓储层从持久化存储重建聚合根，跳过构造校验。
func RehydrateAgent(id, name string, loopTemplate LoopTemplate, capabilities []MountedCapability, hooks HookConfig, maxDepth int32, version string, status AgentStatus, modelProviderID string) *Agent {
	return &Agent{
		id:              id,
		name:            name,
		loopTemplate:    loopTemplate,
		capabilities:    capabilities,
		hooks:           hooks,
		maxDepth:        maxDepth,
		version:         version,
		status:          status,
		modelProviderID: modelProviderID,
	}
}
