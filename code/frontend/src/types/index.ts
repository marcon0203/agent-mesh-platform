// 对应 docs/Agent开放平台_技术规格文档.md 第四章数据模型

export type CapabilityType = "tool" | "skill" | "agent"

export interface Capability {
  id: string
  type: CapabilityType
  name: string
  description: string
  version: string
  callsPerDay: number
}

export interface AgentConfig {
  id: string
  name: string
  mountedCapabilityIds: string[]
  maxDepth: number
  hooks: {
    pre: string[]
    post: string[]
  }
  status: "draft" | "published" | "offline"
}

export interface ChatMessage {
  role: "user" | "agent"
  content: string
}
