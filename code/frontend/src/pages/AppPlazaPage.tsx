import { useMemo, useState } from "react"
import { Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { Search, Bot, MessagesSquare } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { PageHeader } from "@/components/PageHeader"
import { api } from "@/api/client"

const LOOP_TEMPLATE_LABEL: Record<string, string> = {
  react: "ReAct",
  plan_execute: "Plan & Execute",
  pipeline: "确定性 Pipeline",
}

// 应用广场：参考阿里云百炼「应用广场」——把已发布的 Agent 当"应用"摆成卡片，
// 点进去直接带上 agentId 跳到 Workbench 开始对话，不用先去 Workbench 手动
// 填 Agent ID。搜索框对齐参考图右上角的位置，暂时没有"精选/全部"切换——
// 我们的 Agent 没有内置/第三方之分，硬加一个筛选维度意义不大。
export default function AppPlazaPage() {
  const agentsQuery = useQuery({ queryKey: ["agents", "published"], queryFn: api.listPublishedAgents })
  const [search, setSearch] = useState("")

  const agents = useMemo(() => {
    const list = agentsQuery.data ?? []
    if (!search.trim()) return list
    return list.filter((a) => a.name.toLowerCase().includes(search.trim().toLowerCase()))
  }, [agentsQuery.data, search])

  return (
    <div>
      <div className="mb-8 flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
        <PageHeader eyebrow="APP PLAZA" title="应用广场" description="挑一个已发布的 Agent，直接进 Workbench 开始对话。" />
        <div className="relative w-full max-w-xs shrink-0 sm:mt-1">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input value={search} onChange={(e) => setSearch(e.target.value)} placeholder="搜索应用…" className="pl-9" />
        </div>
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
          <CardTitle>还没有已发布的 Agent</CardTitle>
          <CardDescription>
            去"工作区 → Agent 管理"创建并发布一个 Agent，发布后会出现在这里。
          </CardDescription>
        </Card>
      )}

      {!agentsQuery.isLoading && !agentsQuery.error && (agentsQuery.data?.length ?? 0) > 0 && agents.length === 0 && (
        <Card>
          <CardTitle>没有匹配的应用</CardTitle>
          <CardDescription>换个搜索词试试。</CardDescription>
        </Card>
      )}

      {agents.length > 0 && (
        <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
          {agents.map((a) => (
            <Link key={a.id} to={`/console/workbench?agentId=${a.id}`}>
              <Card className="h-full cursor-pointer">
                <CardHeader>
                  <div className="flex items-center justify-between">
                    <span className="flex size-9 items-center justify-center rounded-lg bg-primary/15 text-[#5aa6ff]">
                      <Bot className="size-4.5" />
                    </span>
                    <span className="flex items-center gap-1 text-xs text-muted-foreground">
                      <MessagesSquare className="size-3.5" />
                      开始对话
                    </span>
                  </div>
                  <CardTitle>{a.name}</CardTitle>
                  <CardDescription>
                    {LOOP_TEMPLATE_LABEL[a.loop_template] ?? a.loop_template} · Agent ID：
                    <span className="font-mono">{a.id}</span>
                  </CardDescription>
                </CardHeader>
              </Card>
            </Link>
          ))}
        </div>
      )}
    </div>
  )
}
