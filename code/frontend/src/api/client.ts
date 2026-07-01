// 统一的 API 调用封装，指向 gateway-service（见 vite.config.ts 的 /api 代理）。
// 对应 docs/Agent开放平台_产品规格文档.md 第六章 API Contract。

const BASE_URL = "/api/v1"

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE_URL}${path}`, {
    headers: { "Content-Type": "application/json" },
    ...options,
  })
  if (!res.ok) {
    throw new Error(`API ${path} failed: ${res.status}`)
  }
  return res.json()
}

export const api = {
  // 能力市场
  listCapabilities: () => request("/capabilities"),

  // Agent 发布/调用
  sendMsg: (agentId: string, message: string, sessionId?: string) =>
    request(`/agents/${agentId}/sendMsg`, {
      method: "POST",
      body: JSON.stringify({ message, session_id: sessionId }),
    }),

  // streamMsg 走 SSE，需用 EventSource 单独实现，此处留空作为接入点
  streamMsg: (agentId: string) => {
    return new EventSource(`${BASE_URL}/agents/${agentId}/streamMsg`)
  },
}
