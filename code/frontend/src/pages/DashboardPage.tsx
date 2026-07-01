import { Card, CardHeader, CardTitle, CardDescription } from "@/components/ui/card"
import { PageHeader } from "@/components/PageHeader"

const STATS = [
  { label: "已上架能力", value: "2,480", suffix: "+" },
  { label: "已发布 Agent", value: "860", suffix: "+" },
  { label: "月调用量（Token）", value: "1.2", suffix: " 亿" },
  { label: "平台可用性", value: "99.9", suffix: "%" },
]

export default function DashboardPage() {
  return (
    <div>
      <PageHeader eyebrow="DASHBOARD" title="控制台" />

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

      <Card>
        <CardHeader>
          <CardTitle>API Key 与用量</CardTitle>
          <CardDescription>
            按 Agent 维度统一核算，覆盖 API 调用和 Workbench 对话两个入口。TODO：接入 billing-service 用量查询接口。
          </CardDescription>
        </CardHeader>
      </Card>
    </div>
  )
}
