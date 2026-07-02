import { Link } from "react-router-dom"
import { useQuery } from "@tanstack/react-query"
import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { ConstellationGraphic } from "@/components/Logo"
import { api } from "@/api/client"
import { useAuth } from "@/lib/auth"

const STATS = [
  { label: "已上架能力（Tools / Skills）", value: "2,480", suffix: "+" },
  { label: "已发布 Agent", value: "860", suffix: "+" },
  { label: "月调用量（Token）", value: "1.2", suffix: " 亿" },
  { label: "平台可用性", value: "99.9", suffix: "%" },
]

const TYPE_LABEL: Record<string, string> = { tool: "TOOL", skill: "SKILL", agent: "AGENT" }

const PROCESS_STEPS = [
  { num: "01 / SELECT", title: "挑选能力", desc: "从市场检索或让平台按任务描述智能推荐一组 Tool / Skill / Agent 组合。" },
  { num: "02 / CONFIGURE", title: "配置策略", desc: "设定记忆策略、权限范围与递归深度上限，中间件 Hook 均可按需启用。" },
  { num: "03 / PUBLISH", title: "发布调用", desc: "一键发布，自动生成 API 文档与 Workbench 入口，即刻可被集成或对话使用。" },
]

export default function PortalHomePage() {
  const { isAuthenticated } = useAuth()
  const consoleHref = isAuthenticated ? "/console" : "/login"

  const previewQuery = useQuery({ queryKey: ["capabilities", "tool"], queryFn: () => api.listCapabilities("tool") })
  const preview = (previewQuery.data ?? []).slice(0, 6)

  return (
    <div>
      {/* Hero */}
      <section className="mx-auto grid max-w-6xl grid-cols-1 items-center gap-14 px-8 pb-16 pt-20 lg:grid-cols-[1fr_420px]">
        <div>
          <span className="eyebrow-pill mb-6">AGENT MARKETPLACE · 现已支持能力自由组合</span>
          <h1 className="font-display text-5xl font-bold leading-[1.1] tracking-tight">
            把能力拆到最小粒度
            <br />
            让 <span className="gradient-text">Agent</span> 自己拼装
          </h1>
          <p className="mt-6 max-w-lg text-base leading-relaxed text-muted-foreground">
            Tools、Skills 与 Agent 在同一个市场里自由组合。挑选能力、配置策略、一键发布，
            同时开放 API 与 Workbench 对话两种接入方式。
          </p>
          <div className="mt-9 flex flex-wrap gap-3.5">
            <Button size="lg" asChild>
              <Link to="/marketplace">浏览能力市场</Link>
            </Button>
            <Button size="lg" variant="ghost" asChild>
              <Link to={consoleHref}>{"</> "}进入控制台</Link>
            </Button>
          </div>
          <div className="glass-panel mt-10 inline-flex items-center gap-2.5 px-5 py-3 font-mono text-xs text-muted-foreground">
            <span className="flex gap-1.5">
              <span className="size-2 rounded-full bg-[#ed6a5e]" />
              <span className="size-2 rounded-full bg-[#f4bf4f]" />
              <span className="size-2 rounded-full bg-[#61c554]" />
            </span>
            POST /api/v1/agents/{"{agent_id}"}/streamMsg — 平均首字节 &lt; 800ms
          </div>
        </div>

        <div className="glass-panel relative flex aspect-square items-center justify-center overflow-hidden">
          <div
            className="absolute inset-0"
            style={{ background: "radial-gradient(circle at 50% 40%, rgba(47,111,237,0.18), transparent 65%)" }}
          />
          <ConstellationGraphic className="relative z-10 h-[88%] w-[88%]" />
          <div className="absolute inset-x-0 bottom-6 text-center font-mono text-[11px] tracking-wide text-muted-foreground">
            TOOLS · SKILLS · AGENTS 实时编排
          </div>
        </div>
      </section>

      {/* Stats */}
      <section className="mx-auto max-w-6xl px-8">
        <div className="grid grid-cols-2 gap-4 md:grid-cols-4">
          {STATS.map((s) => (
            <Card key={s.label} className="p-5">
              <div className="font-display text-2xl font-bold">
                <span className="gradient-text">{s.value}</span>
                {s.suffix}
              </div>
              <div className="mt-1 text-xs text-muted-foreground">{s.label}</div>
            </Card>
          ))}
        </div>
      </section>

      {/* Capability preview */}
      <section className="mx-auto max-w-6xl px-8 py-28">
        <div className="mb-12 max-w-xl">
          <p className="eyebrow-pill mb-4">CAPABILITY MARKETPLACE</p>
          <h2 className="font-display text-3xl font-semibold">能力市场，按最小粒度发布与订阅</h2>
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
            Tool 是原子能力，Skill 是多工具组合而成的复合流程，Agent 也可以作为一种能力被其他 Agent 挂载调用。
          </p>
        </div>

        {preview.length > 0 && (
          <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
            {preview.map((cap) => (
              <Card key={cap.id}>
                <CardHeader>
                  <Badge variant={cap.type}>{TYPE_LABEL[cap.type]}</Badge>
                  <CardTitle>{cap.name}</CardTitle>
                  <CardDescription>{cap.is_builtin ? "平台内置能力" : "第三方发布能力"}</CardDescription>
                </CardHeader>
                <CardFooter>
                  <span>{cap.version || "未发布版本"}</span>
                </CardFooter>
              </Card>
            ))}
          </div>
        )}

        <div className="mt-10 text-center">
          <Button variant="outline" asChild>
            <Link to="/marketplace">查看完整能力市场 →</Link>
          </Button>
        </div>
      </section>

      {/* Dual channel */}
      <section className="mx-auto max-w-6xl px-8 py-28">
        <div className="mb-12 max-w-xl">
          <p className="eyebrow-pill mb-4">DUAL CONSUMPTION CHANNELS</p>
          <h2 className="font-display text-3xl font-semibold">发布后，两种方式即可接入</h2>
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
            开放 API 与 Workbench 对话共享同一套执行引擎，行为完全一致，用量统一计入同一份账单。
          </p>
        </div>

        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          <Card className="p-8">
            <h3 className="mb-2 font-display text-xl font-semibold">开放 API</h3>
            <p className="mb-6 text-sm text-muted-foreground">标准 REST 接口，支持同步与 SSE 流式返回，自动生成 OpenAPI 文档。</p>
            <div className="overflow-hidden rounded-xl border border-glass-border bg-black/30">
              <div className="flex items-center gap-1.5 border-b border-glass-border px-3.5 py-2.5">
                <span className="size-2 rounded-full bg-muted-foreground/40" />
                <span className="size-2 rounded-full bg-muted-foreground/40" />
                <span className="size-2 rounded-full bg-muted-foreground/40" />
              </div>
              <div className="space-y-1.5 p-4 font-mono text-xs leading-relaxed">
                <div>
                  <span className="text-muted-foreground">POST</span> /api/v1/agents/<span className="text-accent">rsch-01</span>/streamMsg
                </div>
                <div className="text-[#5aa6ff]">Authorization: Bearer sk-********</div>
                <div className="text-[#5aa6ff]">{"{"}</div>
                <div className="pl-4 text-[#5aa6ff]">
                  "message": <span className="text-accent">"帮我调研三家竞品定价"</span>
                </div>
                <div className="text-[#5aa6ff]">{"}"}</div>
                <div className="pt-2 text-muted-foreground">← 200 event-stream</div>
                <div className="text-accent">data: {"{"}"reply":"正在检索…"{"}"}</div>
              </div>
            </div>
          </Card>

          <Card className="p-8">
            <h3 className="mb-2 font-display text-xl font-semibold">Workbench 对话</h3>
            <p className="mb-6 text-sm text-muted-foreground">
              无需集成即可直接使用，内置调试模式，逐步展示工具调用与子 Agent 调度过程。
            </p>
            <div className="overflow-hidden rounded-xl border border-glass-border bg-black/30">
              <div className="flex items-center gap-1.5 border-b border-glass-border px-3.5 py-2.5">
                <span className="size-2 rounded-full bg-muted-foreground/40" />
                <span className="size-2 rounded-full bg-muted-foreground/40" />
                <span className="size-2 rounded-full bg-muted-foreground/40" />
              </div>
              <div className="space-y-2.5 p-4">
                <div className="ml-auto max-w-[82%] rounded-lg bg-primary/15 px-3.5 py-2 text-xs text-foreground">
                  帮我把这份财报生成一份摘要
                </div>
                <div className="max-w-[82%] rounded-lg border border-glass-border bg-white/5 px-3.5 py-2 text-xs text-muted-foreground">
                  已调用「表格解析」「财报摘要生成」，摘要草稿已就绪 ↓
                </div>
              </div>
            </div>
          </Card>
        </div>
      </section>

      {/* Process */}
      <section className="mx-auto max-w-6xl px-8 py-28">
        <div className="mb-12 max-w-xl">
          <p className="eyebrow-pill mb-4">BUILD PROCESS</p>
          <h2 className="font-display text-3xl font-semibold">三步构建一个可用的 Agent</h2>
          <p className="mt-3 text-sm leading-relaxed text-muted-foreground">
            不需要理解编排逻辑，勾选能力、确认配置、发布即可，执行顺序交给 Agent 自己决定。
          </p>
        </div>

        <div className="grid grid-cols-1 gap-5 md:grid-cols-3">
          {PROCESS_STEPS.map((step) => (
            <Card key={step.num} className="p-7">
              <p className="mb-5 font-mono text-xs font-semibold tracking-widest text-muted-foreground">{step.num}</p>
              <h3 className="mb-2.5 font-display text-lg font-semibold">{step.title}</h3>
              <p className="text-sm leading-relaxed text-muted-foreground">{step.desc}</p>
            </Card>
          ))}
        </div>
      </section>

      {/* Footer CTA */}
      <section className="mx-auto max-w-6xl px-8 pb-24">
        <Card className="relative overflow-hidden p-14 text-center">
          <div
            className="pointer-events-none absolute inset-0"
            style={{ background: "radial-gradient(circle at 50% 0%, rgba(47,111,237,0.22), transparent 60%)" }}
          />
          <h2 className="relative font-display text-3xl font-semibold">把第一个 Agent 拼装起来</h2>
          <p className="relative mt-3.5 text-sm text-muted-foreground">免费额度足够跑通从组装到发布的完整流程。</p>
          <div className="relative mt-8 flex justify-center gap-3.5">
            <Button size="lg" asChild>
              <Link to={consoleHref}>立即开始</Link>
            </Button>
            <Button size="lg" variant="ghost" asChild>
              <Link to="/marketplace">浏览能力市场</Link>
            </Button>
          </div>
        </Card>
      </section>

      <footer className="border-t border-glass-border px-8 py-10">
        <div className="mx-auto flex max-w-6xl items-center justify-between font-mono text-xs text-muted-foreground">
          <span>© 2026 枢络 AgentMesh</span>
          <span>ALL SYSTEMS OPERATIONAL</span>
        </div>
      </footer>
    </div>
  )
}
