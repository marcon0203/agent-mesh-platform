import { useEffect, useRef, useState } from "react"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { api } from "@/api/client"
import type { ChatMessage } from "@/types"

function randomSessionId() {
  return crypto.randomUUID ? crypto.randomUUID() : Math.random().toString(36).slice(2)
}

export default function WorkbenchPage() {
  const [agentId, setAgentId] = useState("")
  const [apiKey, setApiKey] = useState("")
  const [sessionId] = useState(randomSessionId)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [input, setInput] = useState("")
  const [isStreaming, setIsStreaming] = useState(false)
  const [debugLog, setDebugLog] = useState<string[]>([])
  const cancelRef = useRef<(() => void) | null>(null)

  useEffect(() => () => cancelRef.current?.(), [])

  const send = () => {
    if (!input.trim() || !agentId || isStreaming) return

    const userMessage = input
    setMessages((m) => [...m, { role: "user", content: userMessage }, { role: "agent", content: "" }])
    setInput("")
    setDebugLog([])
    setIsStreaming(true)

    let agentContent = ""
    cancelRef.current = api.streamMsg(
      agentId,
      userMessage,
      sessionId,
      apiKey,
      (chunk) => {
        agentContent += chunk.reply
        setMessages((m) => {
          const next = [...m]
          next[next.length - 1] = { role: "agent", content: agentContent }
          return next
        })
        setDebugLog((d) => [
          ...d,
          `→ trace=${chunk.trace_id.slice(0, 8)} tokens=${chunk.usage?.tokens ?? 0} final=${chunk.is_final}`,
        ])
      },
      (err) => {
        setMessages((m) => {
          const next = [...m]
          next[next.length - 1] = { role: "agent", content: `出错了：${err.message}` }
          return next
        })
        setIsStreaming(false)
      },
      () => setIsStreaming(false)
    )
  }

  return (
    <div className="grid grid-cols-1 gap-6 lg:grid-cols-[1fr_320px]">
      <div>
        <div className="mb-6">
          <p className="mb-2 font-mono text-xs uppercase tracking-wide text-primary">Workbench</p>
          <h1 className="font-display text-3xl font-semibold">对话调试</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            与开放 API 共用同一套执行引擎，行为完全一致。
          </p>
        </div>

        <div className="mb-4 grid grid-cols-1 gap-3 sm:grid-cols-2">
          <div>
            <label className="mb-1 block text-xs text-muted-foreground">Agent ID</label>
            <Input
              value={agentId}
              onChange={(e) => setAgentId(e.target.value)}
              placeholder="从 Agent Builder 发布后拿到的 ID"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs text-muted-foreground">API Key</label>
            <Input
              type="password"
              value={apiKey}
              onChange={(e) => setApiKey(e.target.value)}
              placeholder="sk-xxx（平台暂无账号体系，先手动粘贴一个 api_key 表里的 Key）"
            />
          </div>
        </div>

        <Card className="mb-4 min-h-[420px] justify-between">
          <div className="flex flex-col gap-3">
            {messages.length === 0 && (
              <p className="text-sm text-muted-foreground">填好 Agent ID 和 API Key 之后就可以开始对话了。</p>
            )}
            {messages.map((m, i) => (
              <div
                key={i}
                className={`max-w-[80%] whitespace-pre-wrap rounded-xl px-4 py-2.5 text-sm ${
                  m.role === "user"
                    ? "ml-auto bg-primary/15 text-foreground"
                    : "border border-glass-border bg-white/5 text-muted-foreground"
                }`}
              >
                {m.content || (isStreaming && i === messages.length - 1 ? "…" : "")}
              </div>
            ))}
          </div>
          <div className="flex gap-2 pt-4">
            <Input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && send()}
              placeholder="输入消息…"
              disabled={isStreaming}
            />
            <Button onClick={send} disabled={!agentId || isStreaming}>
              {isStreaming ? "生成中…" : "发送"}
            </Button>
          </div>
        </Card>
      </div>

      <Card>
        <p className="mb-3 text-sm font-medium">调试面板</p>
        <p className="mb-4 text-xs text-muted-foreground">
          Session：<span className="font-mono">{sessionId.slice(0, 8)}</span>
          <br />
          实时展示 orchestration-service 流式返回的事件（Hook/Tool 调用详情在服务端日志里，暂未透传到这里）。
        </p>
        <div className="space-y-2 font-mono text-xs text-muted-foreground">
          {debugLog.length === 0 ? <div>暂无事件</div> : debugLog.map((line, i) => <div key={i}>{line}</div>)}
        </div>
      </Card>
    </div>
  )
}
