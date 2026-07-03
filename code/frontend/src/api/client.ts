// 统一的 API 调用封装。/api/v1 按路径前缀被 vite.config.ts 分流到
// gateway-service（sendMsg/streamMsg/openapi.json）、marketplace-service
// （capabilities）、orchestration-service 的 admin HTTP（agents/model-providers），
// 页面组件不需要关心具体打到哪个后端服务。
// 对应 docs/Agent开放平台_产品规格文档.md 第六章 API Contract。

import type {
  AgentDetail,
  AgentSummary,
  Capability,
  CapabilityType,
  HookConfig,
  LoopTemplate,
  ModelProvider,
  MountedCapability,
  ProviderType,
} from "@/types"

const BASE_URL = "/api/v1"

// 平台还没有账号登录体系，暂时用一个固定的开发者 ID 作为模型供应商 / Agent 的归属者。
export const DEV_OWNER_ID = "1"

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  })
  if (!res.ok) {
    const body = await res.text().catch(() => "")
    throw new Error(`API ${path} failed: ${res.status}${body ? ` - ${body}` : ""}`)
  }
  // 有的接口（比如 configureAgent）成功时返回 200 但 body 是空的，不只是
  // 204 会没有 body——统一按"空 body 就当 void 处理"，不强行 res.json()。
  const text = await res.text()
  return (text ? JSON.parse(text) : undefined) as T
}

export interface CreateModelProviderInput {
  name: string
  provider_type: ProviderType
  base_url?: string
  api_key: string
  model_name: string
}

export interface ConfigureAgentInput {
  capabilities: MountedCapability[]
  max_depth: number
  hooks: HookConfig
  model_provider_id: string
  publish: boolean
}

export interface StreamChunk {
  reply: string
  trace_id: string
  is_final: boolean
  usage?: { tokens: number }
}

export const api = {
  // 能力市场
  listCapabilities: (type?: CapabilityType) =>
    request<Capability[]>(`/capabilities${type ? `?type=${type}` : ""}`),
  getCapability: (id: string) => request<Capability>(`/capabilities/${id}`),

  // Agent 装配 / 发布
  createAgent: (name: string, loopTemplate: LoopTemplate = "react") =>
    request<{ id: string; name: string; loop_template: LoopTemplate; status: string }>("/agents", {
      method: "POST",
      body: JSON.stringify({ name, loop_template: loopTemplate }),
    }),
  getAgent: (agentId: string) => request<AgentDetail>(`/agents/${agentId}`),
  // 账号下所有 Agent（含草稿），供 Agent 列表页展示。
  listAllAgents: () => request<AgentSummary[]>(`/agents`),
  // 只列出已发布的 Agent，供挂载 Subagent 时选择（见 AgentEditorPage）。
  listPublishedAgents: () => request<AgentSummary[]>(`/agents?status=published`),
  configureAgent: (agentId: string, input: ConfigureAgentInput) =>
    request<void>(`/agents/${agentId}/config`, {
      method: "POST",
      body: JSON.stringify(input),
    }),

  // 模型供应商
  listModelProviders: () => request<ModelProvider[]>(`/model-providers?owner_id=${DEV_OWNER_ID}`),
  createModelProvider: (input: CreateModelProviderInput) =>
    request<ModelProvider>("/model-providers", {
      method: "POST",
      body: JSON.stringify({ ...input, owner_id: DEV_OWNER_ID }),
    }),

  // Agent 调用（同步）
  sendMsg: (agentId: string, message: string, sessionId: string | undefined, apiKey: string) =>
    request<{ code: number; data: { reply: string; trace_id: string; usage: { tokens: number } } }>(
      `/agents/${agentId}/sendMsg`,
      {
        method: "POST",
        headers: { Authorization: `Bearer ${apiKey}` },
        body: JSON.stringify({ message, session_id: sessionId }),
      }
    ),

  // Agent 调用（SSE 流式）。streamMsg 是 POST 端点，原生 EventSource 只支持 GET
  // 且不能带自定义 Header，所以这里用 fetch + ReadableStream 手动解析
  // `data: {...}\n\n` 帧，而不是返回 EventSource 实例。返回一个取消函数。
  streamMsg: (
    agentId: string,
    message: string,
    sessionId: string | undefined,
    apiKey: string,
    onChunk: (chunk: StreamChunk) => void,
    onError?: (err: Error) => void,
    onDone?: () => void
  ): (() => void) => {
    const controller = new AbortController()

    ;(async () => {
      try {
        const res = await fetch(`${BASE_URL}/agents/${agentId}/streamMsg`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${apiKey}`,
          },
          body: JSON.stringify({ message, session_id: sessionId }),
          signal: controller.signal,
        })
        if (!res.ok || !res.body) {
          const body = await res.text().catch(() => "")
          throw new Error(`streamMsg failed: ${res.status}${body ? ` - ${body}` : ""}`)
        }

        const reader = res.body.getReader()
        const decoder = new TextDecoder()
        let buffer = ""

        for (;;) {
          const { done, value } = await reader.read()
          if (done) break
          buffer += decoder.decode(value, { stream: true })

          let sepIndex
          while ((sepIndex = buffer.indexOf("\n\n")) !== -1) {
            const rawEvent = buffer.slice(0, sepIndex)
            buffer = buffer.slice(sepIndex + 2)
            const dataLine = rawEvent.split("\n").find((line) => line.startsWith("data:"))
            if (!dataLine) continue
            try {
              onChunk(JSON.parse(dataLine.slice(5).trim()) as StreamChunk)
            } catch {
              // 忽略无法解析的帧
            }
          }
        }
        onDone?.()
      } catch (err) {
        if ((err as Error).name !== "AbortError") {
          onError?.(err as Error)
        }
      }
    })()

    return () => controller.abort()
  },
}
