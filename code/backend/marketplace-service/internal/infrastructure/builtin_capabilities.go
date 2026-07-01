package infrastructure

import (
	"github.com/agentmesh/marketplace-service/internal/domain"
)

// platformPublisherID 代表"平台官方"发布者，内置能力都挂在这个虚拟发布者名下。
const platformPublisherID = "1"

// builtinCapabilitySeed 描述一个启动时预写入数据库的内置能力。
// ID 必须和 orchestration-service 的 internal/infrastructure/builtin_tools.go
// 里 builtinTools 映射表的 key 保持一致——两边通过这个固定数值 ID 隐式约定
// "同一个内置能力"，因为内置能力免走 MCP，orchestration 直接按 ID 解析成
// 进程内 Tool 实现，不经过市场的动态发现（技术规格文档 §6.2）。
type builtinCapabilitySeed struct {
	id         string
	name       string
	capType    domain.CapabilityType
	schemaJSON string
}

var builtinCapabilitySeeds = []builtinCapabilitySeed{
	{
		id:      "1",
		name:    "网页检索",
		capType: domain.CapabilityTypeTool,
		schemaJSON: `{
			"input": {"type": "object", "properties": {"url": {"type": "string", "description": "要抓取的网页 URL"}}, "required": ["url"]},
			"output": {"type": "object", "properties": {"text": {"type": "string"}}}
		}`,
	},
	{
		id:      "2",
		name:    "日历解析",
		capType: domain.CapabilityTypeTool,
		schemaJSON: `{
			"input": {"type": "object", "properties": {"text": {"type": "string", "description": "包含日期时间的自然语言"}}, "required": ["text"]},
			"output": {"type": "object", "properties": {"iso8601": {"type": "string"}, "matched": {"type": "boolean"}}}
		}`,
	},
}

// SeedBuiltinCapabilities 在服务启动时把内置能力预写入数据库并直接置为已上架状态
// （内置能力由平台自己维护，不需要走人工审核）。isBuiltin=true 让聚合根的
// SubmitForReview 跳过 MCP Endpoint 校验。可重复调用（Save 是 upsert）。
func SeedBuiltinCapabilities(repo domain.CapabilityRepository) error {
	for _, seed := range builtinCapabilitySeeds {
		cap := domain.NewCapability(seed.id, seed.name, seed.capType, platformPublisherID, true)
		cap.DefineSchema(seed.schemaJSON)
		if err := cap.SubmitForReview(); err != nil {
			return err
		}
		if err := cap.Approve("1.0.0"); err != nil {
			return err
		}
		if err := repo.Save(cap); err != nil {
			return err
		}
	}
	return nil
}
