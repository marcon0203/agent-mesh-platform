// 对应 docs/Agent开放平台_技术规格文档.md 第四章数据模型。
// 字段与后端 DTO 保持一致：marketplace-service 的 capabilityResponse、
// orchestration-service 的 modelProviderResponse / agentDetailResponse
// （见对应服务 internal/interfaces/http_handler.go）。

export type CapabilityType = "tool" | "skill" | "agent"

export interface Capability {
  id: string
  type: CapabilityType
  name: string
  status: "pending_review" | "published" | "offline"
  version: string
  is_builtin: boolean
  schema_json: string
  mcp_endpoint: string
}

export type ProviderType = "openai_compatible" | "anthropic"

export interface ModelProvider {
  id: string
  name: string
  provider_type: ProviderType
  base_url: string
  model_name: string
  status: "enabled" | "disabled"
}

export type LoopTemplate = "react" | "plan_execute" | "pipeline"

export interface MountedCapability {
  capability_id: string
  pinned_version: string
  is_subagent: boolean
}

export interface HookConfig {
  pre_hooks: string[]
  post_hooks: string[]
}

export interface AgentDetail {
  id: string
  name: string
  loop_template: LoopTemplate
  capabilities: MountedCapability[]
  hooks: HookConfig
  max_depth: number
  status: "draft" | "published" | "offline"
  model_provider_id: string
}

export interface ChatMessage {
  role: "user" | "agent"
  content: string
}
