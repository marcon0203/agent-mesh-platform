import type { Capability } from "@/types"

export const MOCK_CAPABILITIES: Capability[] = [
  { id: "cap-web-search", type: "tool", name: "网页检索", description: "接入搜索引擎与网页抓取，返回结构化正文与来源信息。", version: "v2.3.1", callsPerDay: 32400 },
  { id: "cap-calendar", type: "tool", name: "日历解析", description: "从自然语言中抽取时间、地点、参与人，输出标准 iCal 事件对象。", version: "v1.8.0", callsPerDay: 9100 },
  { id: "cap-report-summary", type: "skill", name: "财报摘要生成", description: "组合表格解析、数值校验与结构化写作三个工具，产出可直接送审的摘要草稿。", version: "v3.0.2", callsPerDay: 4600 },
  { id: "cap-ticket-triage", type: "skill", name: "工单智能分诊", description: "识别工单紧急度与类别，自动路由到对应处理队列。", version: "v1.4.5", callsPerDay: 18900 },
  { id: "cap-research-agent", type: "agent", name: "调研助理", description: "可作为子能力挂载：接收课题描述，自主完成多轮检索与信息整合。", version: "v2.1.0", callsPerDay: 2300 },
  { id: "cap-compliance-agent", type: "agent", name: "合规审阅助理", description: "可作为子能力挂载：对文档逐条比对合规条款，输出风险点与修改建议。", version: "v1.2.3", callsPerDay: 1100 },
]
