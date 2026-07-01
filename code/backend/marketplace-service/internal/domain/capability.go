package domain

import "errors"

var (
	ErrInvalidStateTransition = errors.New("invalid capability status transition")
	ErrMissingSchema          = errors.New("capability schema is required before publishing")
	ErrMissingMCPEndpoint     = errors.New("third-party capability requires an MCP endpoint")
)

type CapabilityType string

const (
	CapabilityTypeTool  CapabilityType = "tool"
	CapabilityTypeSkill CapabilityType = "skill"
	CapabilityTypeAgent CapabilityType = "agent"
)

type CapabilityStatus string

const (
	StatusPendingReview CapabilityStatus = "pending_review"
	StatusPublished     CapabilityStatus = "published"
	StatusOffline        CapabilityStatus = "offline"
)

// Capability 是本服务的聚合根。核心不变量：
//   1. 状态机只能沿 待审核 → 已上架 → 已下线 单向流转，不能跳过审核直接上架，
//      也不能从已下线状态复活（要复活必须发新版本重新走审核）。
//   2. 第三方能力（非平台内置）必须提供 MCP Endpoint 才允许提交审核，
//      对应技术规格文档 §6.2 的动态发现机制。
type Capability struct {
	id          string
	name        string
	capType     CapabilityType
	publisherID string
	schemaJSON  string
	mcpEndpoint string
	isBuiltin   bool // 内置能力走进程内调用，不强制要求 MCP，见技术规格文档 §6.2 待确认项
	status      CapabilityStatus
	version     string
}

func NewCapability(id, name string, capType CapabilityType, publisherID string, isBuiltin bool) *Capability {
	return &Capability{
		id:          id,
		name:        name,
		capType:     capType,
		publisherID: publisherID,
		isBuiltin:   isBuiltin,
		status:      StatusPendingReview,
	}
}

// DefineSchema 和 SetMCPEndpoint 是提交审核前的必要准备步骤。
func (c *Capability) DefineSchema(schemaJSON string) { c.schemaJSON = schemaJSON }
func (c *Capability) SetMCPEndpoint(endpoint string) { c.mcpEndpoint = endpoint }

// SubmitForReview 校验发布前置条件，这是聚合根对"合法状态"的把关。
func (c *Capability) SubmitForReview() error {
	if c.schemaJSON == "" {
		return ErrMissingSchema
	}
	if !c.isBuiltin && c.mcpEndpoint == "" {
		return ErrMissingMCPEndpoint
	}
	c.status = StatusPendingReview
	return nil
}

// Approve 只能从"待审核"流转到"已上架"，其他任何状态调用都会被拒绝。
func (c *Capability) Approve(version string) error {
	if c.status != StatusPendingReview {
		return ErrInvalidStateTransition
	}
	c.status = StatusPublished
	c.version = version
	return nil
}

// TakeOffline 只能从"已上架"流转到"已下线"。
func (c *Capability) TakeOffline() error {
	if c.status != StatusPublished {
		return ErrInvalidStateTransition
	}
	c.status = StatusOffline
	return nil
}

func (c *Capability) ID() string                 { return c.id }
func (c *Capability) Name() string                { return c.name }
func (c *Capability) Type() CapabilityType         { return c.capType }
func (c *Capability) Status() CapabilityStatus     { return c.status }
func (c *Capability) Version() string              { return c.version }
func (c *Capability) MCPEndpoint() string          { return c.mcpEndpoint }
func (c *Capability) IsBuiltin() bool              { return c.isBuiltin }
