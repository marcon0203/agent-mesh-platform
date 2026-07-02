import { Link } from "react-router-dom"
import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { PageHeader } from "@/components/PageHeader"

const STATS = [
  { label: "已上架能力", value: "2,480", suffix: "+" },
  { label: "已发布 Agent", value: "860", suffix: "+" },
  { label: "月调用量（Token）", value: "1.2", suffix: " 亿" },
  { label: "平台可用性", value: "99.9", suffix: "%" },
]

const QUICK_LINKS = [
  { to: "/console/builder", title: "Agent 管理", desc: "装配能力、配置 Hook、选择模型供应商并发布。" },
  { to: "/console/model-providers", title: "模型供应商", desc: "接入自己的模型服务商 API Key，供 Agent 选用。" },
  { to: "/console/workbench", title: "Workbench", desc: "免集成直接对话调试，查看实时事件与 trace。" },
]

export default function DashboardPage() {
  return (
    <div>
      <PageHeader eyebrow="DASHBOARD" title="控制台" description="账号下所有 Agent 与能力的用量、配置入口。" />

      <div className="mb-10 grid grid-cols-2 gap-4 md:grid-cols-4">
        {STATS.map((s) => (
          <Card key={s.label} className="p-5">
            <div className="font-display text-3xl font-bold">
              <span className="gradient-text">{s.value}</span>
              {s.suffix}
            </div>
            <div className="mt-1.5 text-xs text-muted-foreground">{s.label}</div>
          </Card>
        ))}
      </div>

      <div className="mb-10 grid grid-cols-1 gap-4 md:grid-cols-3">
        {QUICK_LINKS.map((link) => (
          <Link key={link.to} to={link.to}>
            <Card className="h-full cursor-pointer p-6">
              <CardTitle className="mb-2">{link.title}</CardTitle>
              <CardDescription>{link.desc}</CardDescription>
            </Card>
          </Link>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>API Key 与用量</CardTitle>
          <CardDescription>
            按 Agent 维度统一核算，覆盖 API 调用和 Workbench 对话两个入口。TODO：接入 billing-service 用量查询接口、API Key 增删管理。
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  )
}
