package infrastructure

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// MCPRegistryAdapter 实现 domain.MCPRegistry。
// Register 时会先对 endpoint 做一次 MCP 握手健康检查，失败直接拒绝注册，
// 不再是内存映射表占位（技术规格文档 §6.2）。
type MCPRegistryAdapter struct {
	mu        sync.RWMutex
	endpoints map[string]string
	client    *http.Client
}

func NewMCPRegistryAdapter() *MCPRegistryAdapter {
	return &MCPRegistryAdapter{
		endpoints: make(map[string]string),
		client:    &http.Client{Timeout: 5 * time.Second},
	}
}

// Register 对 endpoint 做连通性检查后再登记；检查失败时不写入映射表，
// 调用方（publish_usecase.go 的 Submit）会把这个 error 原样返回给发布方，
// 阻止一个连不上的 MCP 地址进入待审核状态。
func (m *MCPRegistryAdapter) Register(capabilityID, endpoint string) error {
	if err := checkMCPEndpointHealth(m.client, endpoint); err != nil {
		return fmt.Errorf("MCP endpoint health check failed: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.endpoints[capabilityID] = endpoint
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

// mcpProtocolVersion 是 MCP Streamable HTTP 传输握手用的协议版本号，
// 健康检查只关心握手能不能成功，不代表本服务已经实现完整的 MCP 客户端。
const mcpProtocolVersion = "2025-03-26"

type mcpInitializeRequest struct {
	JSONRPC string              `json:"jsonrpc"`
	ID      int                 `json:"id"`
	Method  string              `json:"method"`
	Params  mcpInitializeParams `json:"params"`
}

type mcpInitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      mcpClientInfo  `json:"clientInfo"`
}

type mcpClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type mcpJSONRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// checkMCPEndpointHealth 发一次 MCP "initialize" JSON-RPC 请求（Streamable HTTP
// 传输），验证对方是不是一个会按 MCP 语义应答的服务，而不只是随便一个能连上的
// HTTP 地址。SSE 传输（Content-Type: text/event-stream）只要能成功建连就算通过，
// 不深入解析后续 SSE 帧——完整的 MCP 会话协商属于真正调用能力时才需要做的事，
// 超出"注册时连通性检查"这一步的范围。
func checkMCPEndpointHealth(client *http.Client, endpoint string) error {
	u, err := url.ParseRequestURI(endpoint)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return fmt.Errorf("invalid MCP endpoint URL: %s", endpoint)
	}

	body, err := json.Marshal(mcpInitializeRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: mcpInitializeParams{
			ProtocolVersion: mcpProtocolVersion,
			Capabilities:    map[string]any{},
			ClientInfo:      mcpClientInfo{Name: "agentmesh-marketplace", Version: "0.1.0"},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot reach MCP endpoint: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("MCP endpoint responded with status %d", resp.StatusCode)
	}

	if strings.Contains(resp.Header.Get("Content-Type"), "text/event-stream") {
		return nil
	}

	var rpcResp mcpJSONRPCResponse
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return fmt.Errorf("MCP endpoint did not return a valid JSON-RPC response: %w", err)
	}
	if rpcResp.JSONRPC != "2.0" {
		return errors.New("MCP endpoint response is not JSON-RPC 2.0")
	}
	if rpcResp.Result == nil && rpcResp.Error == nil {
		return errors.New("MCP endpoint response missing both result and error")
	}
	return nil
}
