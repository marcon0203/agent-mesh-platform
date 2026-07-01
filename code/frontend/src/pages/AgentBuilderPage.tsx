import { useState } from "react"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { MOCK_CAPABILITIES } from "@/data/mock-capabilities"

const PRE_HOOKS = ["记忆加载", "权限校验", "提示词构建"]
const POST_HOOKS = ["消费上报", "沙箱清理", "状态回传"]

export default function AgentBuilderPage() {
  const [name, setName] = useState("")
  const [mounted, setMounted] = useState<string[]>([])
  const [maxDepth, setMaxDepth] = useState(3)
  const [enabledHooks, setEnabledHooks] = useState<string[]>(["记忆加载", "权限校验", "消费上报"])

  const toggle = (list: string[], setList: (v: string[]) => void, id: string) =>
    setList(list.includes(id) ? list.filter((x) => x !== id) : [...list, id])

  return (
    <div>
      <div className="mb-8">
        <p className="mb-2 font-mono text-xs uppercase tracking-wide text-primary">Agent Builder</p>
        <h1 className="font-display text-3xl font-semibold">构建 Agent</h1>
        <p className="mt-2 max-w-xl text-sm text-muted-foreground">
          勾选能力、配置策略即可，执行顺序交给 Agentic Loop 自主决定 —— 不需要手工画流程图。
        </p>
      </div>

      <div className="mb-6 max-w-md">
        <label className="mb-2 block text-sm text-muted-foreground">Agent 名称</label>
        <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="例如：调研助理" />
      </div>

      <h2 className="mb-4 font-display text-lg font-semibold">第一步 · 挑选能力</h2>
      <div className="mb-10 grid grid-cols-1 gap-4 md:grid-cols-2">
        {MOCK_CAPABILITIES.map((cap) => {
          const active = mounted.includes(cap.id)
          return (
            <Card
              key={cap.id}
              onClick={() => toggle(mounted, setMounted, cap.id)}
              className={active ? "cursor-pointer border-primary/60" : "cursor-pointer"}
            >
              <CardHeader>
                <div className="flex items-center justify-between">
                  <Badge variant={cap.type}>{cap.type.toUpperCase()}</Badge>
                  {active && <span className="text-xs font-medium text-accent">已挂载</span>}
                </div>
                <CardTitle>{cap.name}</CardTitle>
                <CardDescription>{cap.description}</CardDescription>
              </CardHeader>
            </Card>
          )
        })}
      </div>

      <h2 className="mb-4 font-display text-lg font-semibold">第二步 · 配置策略</h2>
      <div className="mb-10 grid grid-cols-1 gap-6 md:grid-cols-2">
        <div className="glass-panel p-6">
          <p className="mb-3 text-sm font-medium">前处理 Hooks</p>
          <div className="flex flex-wrap gap-2">
            {PRE_HOOKS.map((h) => (
              <button
                key={h}
                onClick={() => toggle(enabledHooks, setEnabledHooks, h)}
                className={`rounded-full border px-3 py-1.5 text-xs transition-colors ${
                  enabledHooks.includes(h)
                    ? "border-primary/60 bg-primary/15 text-[#5aa6ff]"
                    : "border-glass-border text-muted-foreground"
                }`}
              >
                {h}
              </button>
            ))}
          </div>
          <p className="mb-3 mt-5 text-sm font-medium">后处理 Hooks</p>
          <div className="flex flex-wrap gap-2">
            {POST_HOOKS.map((h) => (
              <button
                key={h}
                onClick={() => toggle(enabledHooks, setEnabledHooks, h)}
                className={`rounded-full border px-3 py-1.5 text-xs transition-colors ${
                  enabledHooks.includes(h)
                    ? "border-accent/50 bg-accent/15 text-accent"
                    : "border-glass-border text-muted-foreground"
                }`}
              >
                {h}
              </button>
            ))}
          </div>
        </div>

        <div className="glass-panel p-6">
          <p className="mb-3 text-sm font-medium">子 Agent 最大递归深度</p>
          <p className="mb-4 text-xs text-muted-foreground">
            当挂载的能力中包含其他 Agent（Subagent-as-Tool）时生效，防止无限递归消耗。
          </p>
          <Input
            type="number"
            min={0}
            max={5}
            value={maxDepth}
            onChange={(e) => setMaxDepth(Number(e.target.value))}
            className="max-w-[120px]"
          />
        </div>
      </div>

      <Button size="lg" disabled={!name || mounted.length === 0}>
        发布 Agent
      </Button>
    </div>
  )
}
