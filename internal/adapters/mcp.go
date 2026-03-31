package adapters

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mkh/rice-railing/internal/exec"
)

// MCPAdapter fetches external context via MCP servers.
// Implements graceful degradation — fails silently when unavailable.
type MCPAdapter struct {
	runner  *exec.Runner
	servers []string
	enabled bool
}

// NewMCPAdapter creates an MCP adapter for the given server list.
func NewMCPAdapter(runner *exec.Runner, servers []string, enabled bool) *MCPAdapter {
	return &MCPAdapter{
		runner:  runner,
		servers: servers,
		enabled: enabled,
	}
}

// MCPResource represents data fetched from an MCP server.
type MCPResource struct {
	Server  string `json:"server"`
	Type    string `json:"type"` // policy_pack, docs, rules, metadata
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// FetchPolicyPack attempts to fetch a company policy pack from an MCP server.
func (m *MCPAdapter) FetchPolicyPack(ctx context.Context, serverName string) (*MCPResource, error) {
	if !m.enabled {
		return nil, fmt.Errorf("MCP is disabled in constitution")
	}
	return m.callServer(ctx, serverName, "policy_pack", "fetch_policy_pack")
}

// FetchRules attempts to fetch reusable rules from an MCP server.
func (m *MCPAdapter) FetchRules(ctx context.Context, serverName string) (*MCPResource, error) {
	if !m.enabled {
		return nil, fmt.Errorf("MCP is disabled in constitution")
	}
	return m.callServer(ctx, serverName, "rules", "fetch_rules")
}

// FetchDocs attempts to fetch documentation from an MCP server.
func (m *MCPAdapter) FetchDocs(ctx context.Context, serverName string) (*MCPResource, error) {
	if !m.enabled {
		return nil, fmt.Errorf("MCP is disabled in constitution")
	}
	return m.callServer(ctx, serverName, "docs", "fetch_docs")
}

// ListAvailable checks which configured MCP servers are reachable.
func (m *MCPAdapter) ListAvailable(ctx context.Context) []MCPResource {
	var results []MCPResource
	for _, server := range m.servers {
		res, err := m.callServer(ctx, server, "ping", "ping")
		if err != nil {
			results = append(results, MCPResource{
				Server: server,
				Type:   "ping",
				Error:  err.Error(),
			})
		} else {
			results = append(results, *res)
		}
	}
	return results
}

// callServer invokes an MCP server via the npx/mcp-client CLI pattern.
// MCP servers are typically invoked via stdio JSON-RPC. We use a thin
// CLI wrapper approach: `npx @modelcontextprotocol/client <server> <method>`.
// Falls back gracefully if the client is not installed.
func (m *MCPAdapter) callServer(ctx context.Context, serverName, resourceType, method string) (*MCPResource, error) {
	// Try direct npx MCP client invocation
	result, err := m.runner.Run(ctx, "npx", "--yes", "@anthropic-ai/mcp-client",
		"--server", serverName,
		"--method", method,
		"--output", "json")

	if err != nil {
		// Fallback: try curl-based HTTP MCP endpoint
		return m.tryHTTPFallback(ctx, serverName, method)
	}

	if result.ExitCode != 0 {
		return &MCPResource{
			Server: serverName,
			Type:   resourceType,
			Error:  fmt.Sprintf("MCP call failed (exit %d): %s", result.ExitCode, strings.TrimSpace(result.Stderr)),
		}, nil
	}

	return &MCPResource{
		Server:  serverName,
		Type:    resourceType,
		Content: result.Stdout,
	}, nil
}

func (m *MCPAdapter) tryHTTPFallback(ctx context.Context, serverName, method string) (*MCPResource, error) {
	// MCP servers can also expose HTTP endpoints
	// Try a simple JSON-RPC call
	payload := fmt.Sprintf(`{"jsonrpc":"2.0","id":1,"method":"%s","params":{}}`, method)
	result, err := m.runner.Run(ctx, "curl", "-sf",
		"-H", "Content-Type: application/json",
		"-d", payload,
		fmt.Sprintf("http://%s/rpc", serverName))

	if err != nil {
		return nil, fmt.Errorf("MCP server %s unreachable: %w", serverName, err)
	}

	if result.ExitCode != 0 {
		return nil, fmt.Errorf("MCP server %s returned error", serverName)
	}

	// Parse JSON-RPC response
	var rpcResponse struct {
		Result json.RawMessage `json:"result"`
		Error  *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal([]byte(result.Stdout), &rpcResponse); err != nil {
		return nil, fmt.Errorf("MCP server %s returned invalid JSON: %w", serverName, err)
	}

	if rpcResponse.Error != nil {
		return nil, fmt.Errorf("MCP server %s: %s", serverName, rpcResponse.Error.Message)
	}

	return &MCPResource{
		Server:  serverName,
		Type:    method,
		Content: string(rpcResponse.Result),
	}, nil
}
