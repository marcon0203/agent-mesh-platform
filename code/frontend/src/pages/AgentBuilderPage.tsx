import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { api } from "@/api/client"
import type { MountedCapability } from "@/types"

const PRE_HOOKS = ["记忆加载", "权限校验", "提示词构建"]
const POST_HOOKS = ["消费上报", "沙箱清理", "状态回传"]

export default function AgentBuilderPage() {
  const queryClient = useQueryClient()

  const [agentId, setAgentId] = useState<string | null>(null)
  const [name, setName] = useState("")
  const [mounted, setMounted] = useState<string[]>([])
  const [maxDepth, setMaxDepth] = useState(3)
  const [enabledHooks, setEnabledHooks] = useState<string[]>(["记忆加载", "权限校验", "消费上报"])
  const [modelProviderId, setModelProviderId] = useState<string>("")

  const toolsQuery = useQuery({ queryKey: ["capabilities", "tool"], queryFn: () => api.listCapabilities("tool") })
  const skillsQuery = useQuery({ queryKey: ["capabilities", "skill"], queryFn: () => api.listCapabilities("skill") })
  const agentsQuery = useQuery({ queryKey: ["capabilities", "agent"], queryFn: () => api.listCapabilities("agent") })
  const providersQuery = useQuery({ queryKey: ["model-providers"], queryFn: api.listModelProviders })

  const capabilities = [...(toolsQuery.data ?? []), ...(skillsQuery.data ?? []), ...(agentsQuery.data ?? [])]

  const createAgentMutation = useMutation({
    mutationFn: () => api.createAgent(name),
    onSuccess: (created) => setAgentId(created.id),
  })

  const publishMutation = useMutation({
    mutationFn: () => {
      if (!agentId) throw new Error("请先创建 Agent")
      const capabilityPayload: MountedCapability[] = mounted.map((id) => {
        const cap = capabilities.find((c) => c.id === id)
        return { capability_id: id, pinned_version: cap?.version ?? "", is_subagent: cap?.type === "agent" }
      })
      return api.configureAgent(agentId, {
        capabilities: capabilityPayload,
        max_depth: maxDepth,
        hooks: {
          pre_hooks: enabledHooks.filter((h) => PRE_HOOKS.includes(h)),
          post_hooks: enabledHooks.filter((h) => POST_HOOKS.includes(h)),
        },
        model_provider_id: modelProviderId,
        publish: true,
      })
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ["agent", agentId] }),
  })

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
        <div className="flex gap-2">
          <Input
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="例如：调研助理"
            disabled={!!agentId}
          />
          {!agentId && (
            <Button
              disabled={!name || createAgentMutation.isPending}
              onClick={() => createAgentMutation.mutate()}
            >
              {createAgentMutation.isPending ? "创建中…" : "创建"}
            </Button>
          )}
        </div>
        {agentId && (
          <p className="mt-2 text-xs text-muted-foreground">
            Agent ID：<span className="font-mono text-accent">{agentId}</span>
          </p>
        )}
        {createAgentMutation.error && (
          <p className="mt-2 text-xs text-destructive">{(createAgentMutation.error as Error).message}</p>
        )}
      </div>

      {!agentId && (
        <Card>
          <CardTitle>先创建 Agent 再继续</CardTitle>
          <CardDescription>填写名称并点击"创建"，之后才能挑选能力、配置策略和发布。</CardDescription>
        </Card>
      )}

      {agentId && (
        <>
          <h2 className="mb-4 font-display text-lg font-semibold">第一步 · 挑选能力</h2>
          {capabilities.length === 0 ? (
            <Card className="mb-10">
              <CardDescription>市场里还没有已上架的能力，先去能力市场页确认 marketplace-service 是否正常。</CardDescription>
            </Card>
          ) : (
            <div className="mb-10 grid grid-cols-1 gap-4 md:grid-cols-2">
              {capabilities.map((cap) => {
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
                      <CardDescription>
                        {cap.type === "agent" ? "作为 Subagent 挂载" : "作为 Tool/Skill 挂载"} · {cap.version || "未发布版本"}
                      </CardDescription>
                    </CardHeader>
                  </Card>
                )
              })}
            </div>
          )}

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
                className="mb-6 max-w-[120px]"
              />

              <p className="mb-3 text-sm font-medium">模型供应商</p>
              {providersQuery.data && providersQuery.data.length > 0 ? (
                <div className="flex flex-wrap gap-2">
                  {providersQuery.data.map((p) => (
                    <button
                      key={p.id}
                      onClick={() => setModelProviderId(p.id)}
                      className={`rounded-full border px-3 py-1.5 text-xs transition-colors ${
                        modelProviderId === p.id
                          ? "border-primary/60 bg-primary/15 text-[#5aa6ff]"
                          : "border-glass-border text-muted-foreground"
                      }`}
                    >
                      {p.name}
                    </button>
                  ))}
                </div>
              ) : (
                <p className="text-xs text-muted-foreground">
                  还没有配置模型供应商，去"模型供应商"页创建一个后再回来选择。
                </p>
              )}
            </div>
          </div>

          {publishMutation.error && (
            <p className="mb-3 text-sm text-destructive">{(publishMutation.error as Error).message}</p>
          )}
          {publishMutation.isSuccess && <p className="mb-3 text-sm text-accent">已发布，可以去 Workbench 里对话了。</p>}

          <Button
            size="lg"
            disabled={mounted.length === 0 || !modelProviderId || publishMutation.isPending}
            onClick={() => publishMutation.mutate()}
          >
            {publishMutation.isPending ? "发布中…" : "发布 Agent"}
          </Button>
        </>
      )}
    </div>
  )
}
