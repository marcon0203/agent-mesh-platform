package domain

// CapabilityRepository 是领域层定义的仓储接口。
type CapabilityRepository interface {
	FindByID(id string) (*Capability, error)
	Save(capability *Capability) error
	ListPublished(capType CapabilityType) ([]*Capability, error)
}

// MCPRegistry 是领域层对"能力动态发现"的抽象端口。
// 具体如何连接 MCP Server、如何做健康检查，属于技术细节，留给 infrastructure 层。
// 详见 docs/Agent开放平台_技术规格文档.md 6.2 节。
type MCPRegistry interface {
	Register(capabilityID, endpoint string) error
	Resolve(capabilityID string) (endpoint string, err error)
}
