import { useQuery } from "@tanstack/react-query"
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { api } from "@/api/client"

const TYPE_LABEL: Record<string, string> = { tool: "TOOL", skill: "SKILL", agent: "AGENT" }
const STATUS_LABEL: Record<string, string> = {
  pending_review: "待审核",
  published: "已上架",
  offline: "已下线",
}

export default function MarketplacePage() {
  const toolsQuery = useQuery({ queryKey: ["capabilities", "tool"], queryFn: () => api.listCapabilities("tool") })
  const skillsQuery = useQuery({ queryKey: ["capabilities", "skill"], queryFn: () => api.listCapabilities("skill") })
  const agentsQuery = useQuery({ queryKey: ["capabilities", "agent"], queryFn: () => api.listCapabilities("agent") })

  const isLoading = toolsQuery.isLoading || skillsQuery.isLoading || agentsQuery.isLoading
  const error = toolsQuery.error || skillsQuery.error || agentsQuery.error
  const capabilities = [
    ...(toolsQuery.data ?? []),
    ...(skillsQuery.data ?? []),
    ...(agentsQuery.data ?? []),
  ]

  return (
    <div>
      <div className="mb-8">
        <p className="mb-2 font-mono text-xs uppercase tracking-wide text-primary">Capability Marketplace</p>
        <h1 className="font-display text-3xl font-semibold">能力市场</h1>
        <p className="mt-2 max-w-xl text-sm text-muted-foreground">
          Tool 是原子能力，Skill 是多工具组合而成的复合流程，Agent 也可以作为一种能力被其他 Agent 挂载。
        </p>
      </div>

      <Input placeholder="按名称或用途搜索能力…" className="mb-8 max-w-md" disabled />

      {isLoading && <p className="text-sm text-muted-foreground">加载中…</p>}

      {error && (
        <Card className="border-destructive/40">
          <CardTitle className="text-destructive">能力列表加载失败</CardTitle>
          <CardDescription>{(error as Error).message}</CardDescription>
        </Card>
      )}

      {!isLoading && !error && capabilities.length === 0 && (
        <Card>
          <CardTitle>市场里还没有已上架的能力</CardTitle>
          <CardDescription>
            marketplace-service 启动时会自动预置两个内置 Tool（网页检索、日历解析），
            如果这里是空的，检查一下 marketplace-service 是否已经启动并连上了 MySQL。
          </CardDescription>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
        {capabilities.map((cap) => (
          <Card key={cap.id}>
            <CardHeader>
              <div className="flex items-center justify-between">
                <Badge variant={cap.type}>{TYPE_LABEL[cap.type]}</Badge>
                {cap.is_builtin && <span className="text-xs font-medium text-accent">平台内置</span>}
              </div>
              <CardTitle>{cap.name}</CardTitle>
              <CardDescription>
                ID：<span className="font-mono">{cap.id}</span>（装配 Agent 时用它来挂载这个能力）
              </CardDescription>
            </CardHeader>
            <CardFooter>
              <span>{cap.version || "未发布版本"}</span>
              <span className="text-accent">{STATUS_LABEL[cap.status] ?? cap.status}</span>
            </CardFooter>
          </Card>
        ))}
      </div>
    </div>
  )
}
