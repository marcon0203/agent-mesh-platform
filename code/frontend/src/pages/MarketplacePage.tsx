import { useMemo, useState } from "react"
import { useQuery } from "@tanstack/react-query"
import { LayoutGrid, Bot, Wrench, Sparkles, Plug, Search, type LucideIcon } from "lucide-react"
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { PageHeader } from "@/components/PageHeader"
import { api } from "@/api/client"
import { cn } from "@/lib/utils"
import type { Capability, CapabilityType } from "@/types"

const TYPE_LABEL: Record<CapabilityType, string> = { tool: "TOOL", skill: "SKILL", agent: "AGENT" }
const STATUS_LABEL: Record<string, string> = {
  pending_review: "待审核",
  published: "已上架",
  offline: "已下线",
}

type CategoryKey = "all" | CapabilityType | "mcp"

interface CategoryDef {
  key: CategoryKey
  label: string
  icon: LucideIcon
}

// 侧边分类参考百炼「模型广场」左侧的模态筛选栏，把我们的能力目录按
// Agent / Tool / Skill 三种类型分开，再加一个横切的「MCP 接入」——
// 对应技术规格文档 §6.2：MCP 是第三方能力的接入协议，不是和 Tool/Skill
// 并列的第四种能力类型，所以这里过滤的是"非内置且带 mcp_endpoint"的能力，
// 跨 Tool/Skill 都可能命中。
const CATEGORIES: CategoryDef[] = [
  { key: "all", label: "全部能力", icon: LayoutGrid },
  { key: "agent", label: "Agent", icon: Bot },
  { key: "tool", label: "Tool", icon: Wrench },
  { key: "skill", label: "Skill", icon: Sparkles },
  { key: "mcp", label: "MCP 接入", icon: Plug },
]

const SECTION_ORDER: { type: CapabilityType; label: string; icon: LucideIcon }[] = [
  { type: "agent", label: "Agent", icon: Bot },
  { type: "tool", label: "Tool", icon: Wrench },
  { type: "skill", label: "Skill", icon: Sparkles },
]

const HERO_GRADIENTS = [
  "from-[#2f6fed]/40 via-[#2f6fed]/10 to-transparent",
  "from-[#3fe0d0]/35 via-[#3fe0d0]/8 to-transparent",
  "from-[#b48cff]/40 via-[#b48cff]/10 to-transparent",
]

function SkeletonGrid() {
  return (
    <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
      {Array.from({ length: 6 }).map((_, i) => (
        <div key={i} className="glass-panel h-40 animate-pulse p-6">
          <div className="h-4 w-16 rounded-full bg-white/10" />
          <div className="mt-4 h-4 w-2/3 rounded bg-white/10" />
          <div className="mt-3 h-3 w-full rounded bg-white/5" />
        </div>
      ))}
    </div>
  )
}

function CapabilityCard({ cap }: { cap: Capability }) {
  return (
    <Card key={cap.id}>
      <CardHeader>
        <div className="flex items-center justify-between">
          <Badge variant={cap.type}>{TYPE_LABEL[cap.type]}</Badge>
          <div className="flex items-center gap-2">
            {!cap.is_builtin && cap.mcp_endpoint && <Badge variant="neutral">MCP</Badge>}
            {cap.is_builtin && <span className="text-xs font-medium text-accent">平台内置</span>}
          </div>
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
  )
}

function HeroCard({ cap, gradient }: { cap: Capability; gradient: string }) {
  return (
    <div
      className={cn(
        "glass-panel relative flex h-40 flex-col justify-end overflow-hidden bg-gradient-to-br p-6",
        gradient
      )}
    >
      <Badge variant={cap.type} className="absolute right-5 top-5">
        {TYPE_LABEL[cap.type]}
      </Badge>
      <div className="font-display text-xl font-bold">{cap.name}</div>
      <div className="mt-1 text-xs text-muted-foreground">
        {cap.is_builtin ? "平台内置 · 免 MCP 直接调用" : "第三方能力 · MCP 接入"}
      </div>
    </div>
  )
}

export default function MarketplacePage() {
  const toolsQuery = useQuery({ queryKey: ["capabilities", "tool"], queryFn: () => api.listCapabilities("tool") })
  const skillsQuery = useQuery({ queryKey: ["capabilities", "skill"], queryFn: () => api.listCapabilities("skill") })
  const agentsQuery = useQuery({ queryKey: ["capabilities", "agent"], queryFn: () => api.listCapabilities("agent") })

  const [category, setCategory] = useState<CategoryKey>("all")
  const [featuredOnly, setFeaturedOnly] = useState(false)
  const [search, setSearch] = useState("")

  const isLoading = toolsQuery.isLoading || skillsQuery.isLoading || agentsQuery.isLoading
  const error = toolsQuery.error || skillsQuery.error || agentsQuery.error
  const allCapabilities = useMemo(
    () => [...(toolsQuery.data ?? []), ...(skillsQuery.data ?? []), ...(agentsQuery.data ?? [])],
    [toolsQuery.data, skillsQuery.data, agentsQuery.data]
  )

  const filtered = useMemo(() => {
    let pool = allCapabilities
    if (featuredOnly) pool = pool.filter((c) => c.is_builtin)
    if (search.trim()) pool = pool.filter((c) => c.name.toLowerCase().includes(search.trim().toLowerCase()))
    if (category === "mcp") pool = pool.filter((c) => !c.is_builtin && c.mcp_endpoint)
    else if (category !== "all") pool = pool.filter((c) => c.type === category)
    return pool
  }, [allCapabilities, featuredOnly, search, category])

  // 首推能力：只在浏览全部、不搜索时展示，取内置能力里的前三个，
  // 呼应百炼模型广场顶部的「首推模型」大卡片区。
  const featured =
    category === "all" && !search.trim() ? allCapabilities.filter((c) => c.is_builtin).slice(0, 3) : []

  const sections = SECTION_ORDER.map((s) => ({ ...s, items: filtered.filter((c) => c.type === s.type) })).filter(
    (s) => s.items.length > 0
  )

  return (
    <div className="mx-auto max-w-7xl px-8 py-16">
      <PageHeader
        eyebrow="CAPABILITY MARKETPLACE"
        title="能力广场"
        description="Tool 是原子能力，Skill 是多工具组合而成的复合流程，Agent 也可以作为一种能力被其他 Agent 挂载；第三方能力统一走 MCP 协议接入。"
      />

      <div className="flex flex-col gap-8 lg:flex-row">
        <aside className="shrink-0 lg:w-48">
          <div className="glass-panel p-2">
            <ul className="flex flex-row gap-1 overflow-x-auto lg:flex-col lg:overflow-visible">
              {CATEGORIES.map((c) => {
                const Icon = c.icon
                const active = category === c.key
                return (
                  <li key={c.key} className="shrink-0">
                    <button
                      onClick={() => setCategory(c.key)}
                      className={cn(
                        "flex w-full items-center gap-2.5 whitespace-nowrap rounded-lg px-3 py-2 text-sm font-medium text-muted-foreground transition-colors hover:bg-white/5 hover:text-foreground",
                        active && "bg-primary/15 text-[#5aa6ff]"
                      )}
                    >
                      <Icon className="size-4 shrink-0" />
                      {c.label}
                    </button>
                  </li>
                )
              })}
            </ul>
          </div>
        </aside>

        <div className="min-w-0 flex-1">
          <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <div className="inline-flex rounded-full border border-glass-border p-1">
              {(
                [
                  { key: false, label: "全部能力" },
                  { key: true, label: "精选能力" },
                ] as const
              ).map((opt) => (
                <button
                  key={String(opt.key)}
                  onClick={() => setFeaturedOnly(opt.key)}
                  className={cn(
                    "rounded-full px-4 py-1.5 text-sm font-medium transition-colors",
                    featuredOnly === opt.key
                      ? "bg-primary/15 text-[#5aa6ff]"
                      : "text-muted-foreground hover:text-foreground"
                  )}
                >
                  {opt.label}
                </button>
              ))}
            </div>

            <div className="relative max-w-xs">
              <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
              <Input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="按名称搜索能力…"
                className="pl-9"
              />
            </div>
          </div>

          {isLoading && <SkeletonGrid />}

          {error && (
            <Card className="border-destructive/40">
              <CardTitle className="text-destructive">能力列表加载失败</CardTitle>
              <CardDescription>{(error as Error).message}</CardDescription>
            </Card>
          )}

          {!isLoading && !error && allCapabilities.length === 0 && (
            <Card>
              <CardTitle>市场里还没有已上架的能力</CardTitle>
              <CardDescription>
                marketplace-service 启动时会自动预置两个内置 Tool（网页检索、日历解析），
                如果这里是空的，检查一下 marketplace-service 是否已经启动并连上了 PostgreSQL。
              </CardDescription>
            </Card>
          )}

          {!isLoading && !error && allCapabilities.length > 0 && filtered.length === 0 && (
            <Card>
              <CardTitle>没有匹配的能力</CardTitle>
              <CardDescription>换个分类或者清空搜索词试试。</CardDescription>
            </Card>
          )}

          {!isLoading && !error && featured.length > 0 && (
            <div className="mb-10">
              <h2 className="mb-4 font-display text-lg font-semibold">首推能力</h2>
              <div className="grid grid-cols-1 gap-5 md:grid-cols-3">
                {featured.map((cap, i) => (
                  <HeroCard key={cap.id} cap={cap} gradient={HERO_GRADIENTS[i % HERO_GRADIENTS.length]} />
                ))}
              </div>
            </div>
          )}

          {!isLoading && !error && category !== "all" && filtered.length > 0 && (
            <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
              {filtered.map((cap) => (
                <CapabilityCard key={cap.id} cap={cap} />
              ))}
            </div>
          )}

          {!isLoading &&
            !error &&
            category === "all" &&
            sections.map((s) => {
              const Icon = s.icon
              return (
                <div key={s.type} className="mb-10">
                  <h2 className="mb-4 flex items-center gap-2 font-display text-lg font-semibold">
                    <Icon className="size-4 text-muted-foreground" />
                    {s.label}
                  </h2>
                  <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
                    {s.items.map((cap) => (
                      <CapabilityCard key={cap.id} cap={cap} />
                    ))}
                  </div>
                </div>
              )
            })}
        </div>
      </div>
    </div>
  )
}
