package infrastructure

import (
	"errors"
	"sync"
)

// MCPRegistryAdapter 实现 domain.MCPRegistry。
// MVP 阶段用内存映射表占位，后续替换为真正的 MCP Server 健康检查 + 连接池。
type MCPRegistryAdapter struct {
	mu        sync.RWMutex
	endpoints map[string]string
}

func NewMCPRegistryAdapter() *MCPRegistryAdapter {
	return &MCPRegistryAdapter{endpoints: make(map[string]string)}
}

func (m *MCPRegistryAdapter) Register(capabilityID, endpoint string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.endpoints[capabilityID] = endpoint
	// TODO: 对 endpoint 做一次连通性健康检查，失败则拒绝注册
	return nil
}

func (m *MCPRegistryAdapter) Resolve(capabilityID string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	endpoint, ok := m.endpoints[capabilityID]
	if !ok {
		return "", errors.New("no MCP endpoint registered for capability: " + capabilityID)
	}
	return endpoint, nil
}
