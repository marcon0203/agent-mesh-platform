import { Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { PageHeader } from "@/components/PageHeader"
import { api } from "@/api/client"

const STATUS_LABEL: Record<string, string> = {
  draft: "草稿",
  published: "已发布",
  offline: "已下线",
}

export default function AgentListPage() {
  const agentsQuery = useQuery({ queryKey: ["agents", "all"], queryFn: api.listAllAgents })

  return (
    <div>
      <PageHeader
        eyebrow="AGENT MANAGEMENT"
        title="Agent 管理"
        description="装配能力、配置 Hook、选择模型供应商并发布——每个 Agent 都有自己独立的编辑页。"
      />

      <div className="mb-6 flex items-center justify-between">
        <h2 className="font-display text-lg font-semibold">我的 Agent</h2>
        <Button asChild>
          <Link to="/console/builder/new">新建 Agent</Link>
        </Button>
      </div>

      {agentsQuery.isLoading && <p className="text-sm text-muted-foreground">加载中…</p>}
      {agentsQuery.error && (
        <Card className="border-destructive/40">
          <CardTitle className="text-destructive">加载失败</CardTitle>
          <CardDescription>{(agentsQuery.error as Error).message}</CardDescription>
        </Card>
      )}
      {!agentsQuery.isLoading && !agentsQuery.error && (agentsQuery.data?.length ?? 0) === 0 && (
        <Card>
          <CardTitle>还没有创建任何 Agent</CardTitle>
          <CardDescription>点右上角"新建 Agent"，挑能力、配策略、选模型供应商，几步就能发布。</CardDescription>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
        {agentsQuery.data?.map((a) => (
          <Link key={a.id} to={`/console/builder/${a.id}`}>
            <Card className="h-full cursor-pointer">
              <CardHeader>
                <div className="flex items-center justify-between">
                  <Badge variant={a.status === "published" ? "agent" : "neutral"}>{STATUS_LABEL[a.status] ?? a.status}</Badge>
                </div>
                <CardTitle>{a.name}</CardTitle>
                <CardDescription>执行机制：{a.loop_template}</CardDescription>
              </CardHeader>
              <CardFooter>
                <span>
                  Agent ID：<span className="font-mono">{a.id}</span>
                </span>
              </CardFooter>
            </Card>
          </Link>
        ))}
      </div>
    </div>
  )
}
