import { useState } from "react"
import { Card } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import type { ChatMessage } from "@/types"

const SEED: ChatMessage[] = [
  { role: "user", content: "帮我把这份财报生成一份摘要" },
  { role: "agent", content: "已调用「表格解析」「财报摘要生成」，摘要草稿已就绪。" },
]

export default function WorkbenchPage() {
  const [messages, setMessages] = useState<ChatMessage[]>(SEED)
  const [input, setInput] = useState("")

  const send = () => {
    if (!input.trim()) return
    setMessages((m) => [...m, { role: "user", content: input }])
    setInput("")
    // TODO: 接入 api.sendMsg / streamMsg，走与开放 API 相同的执行引擎
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

        <Card className="mb-4 min-h-[420px] justify-between">
          <div className="flex flex-col gap-3">
            {messages.map((m, i) => (
              <div
                key={i}
                className={`max-w-[80%] rounded-xl px-4 py-2.5 text-sm ${
                  m.role === "user"
                    ? "ml-auto bg-primary/15 text-foreground"
                    : "border border-glass-border bg-white/5 text-muted-foreground"
                }`}
              >
                {m.content}
              </div>
            ))}
          </div>
          <div className="flex gap-2 pt-4">
            <Input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && send()}
              placeholder="输入消息…"
            />
            <Button onClick={send}>发送</Button>
          </div>
        </Card>
      </div>

      <Card>
        <p className="mb-3 text-sm font-medium">调试面板</p>
        <p className="mb-4 text-xs text-muted-foreground">
          展示本轮 Hook 触发情况与 Tool / Subagent 调用详情。
        </p>
        <div className="space-y-2 font-mono text-xs text-muted-foreground">
          <div>→ Hook: 记忆加载</div>
          <div>→ Hook: 权限校验</div>
          <div>→ Tool: 表格解析</div>
          <div>→ Skill: 财报摘要生成</div>
          <div>→ Hook: 消费上报</div>
        </div>
      </Card>
    </div>
  )
}
