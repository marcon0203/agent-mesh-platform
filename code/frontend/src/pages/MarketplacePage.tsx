import { Card, CardHeader, CardTitle, CardDescription, CardFooter } from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { MOCK_CAPABILITIES } from "@/data/mock-capabilities"

const TYPE_LABEL: Record<string, string> = { tool: "TOOL", skill: "SKILL", agent: "AGENT" }

export default function MarketplacePage() {
  return (
    <div>
      <div className="mb-8">
        <p className="mb-2 font-mono text-xs uppercase tracking-wide text-primary">Capability Marketplace</p>
        <h1 className="font-display text-3xl font-semibold">能力市场</h1>
        <p className="mt-2 max-w-xl text-sm text-muted-foreground">
          Tool 是原子能力，Skill 是多工具组合而成的复合流程，Agent 也可以作为一种能力被其他 Agent 挂载。
        </p>
      </div>

      <Input placeholder="按名称或用途搜索能力…" className="mb-8 max-w-md" />

      <div className="grid grid-cols-1 gap-5 md:grid-cols-2 lg:grid-cols-3">
        {MOCK_CAPABILITIES.map((cap) => (
          <Card key={cap.id}>
            <CardHeader>
              <Badge variant={cap.type}>{TYPE_LABEL[cap.type]}</Badge>
              <CardTitle>{cap.name}</CardTitle>
              <CardDescription>{cap.description}</CardDescription>
            </CardHeader>
            <CardFooter>
              <span>{cap.version}</span>
              <span className="text-accent">{cap.callsPerDay.toLocaleString()} 次/日</span>
            </CardFooter>
          </Card>
        ))}
      </div>
    </div>
  )
}
