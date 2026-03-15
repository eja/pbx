// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/eja/pbx/sys"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

var (
	mcpClientMu    sync.Mutex
	mcpClientCache = map[string]*client.Client{}
)

func getMCPClient(serverURL, token string) (*client.Client, error) {
	key := serverURL + "|" + token
	mcpClientMu.Lock()
	defer mcpClientMu.Unlock()
	if c, ok := mcpClientCache[key]; ok {
		return c, nil
	}
	c, err := newMCPClient(serverURL, token)
	if err != nil {
		return nil, err
	}
	mcpClientCache[key] = c
	return c, nil
}

func newMCPClient(serverURL, token string) (*client.Client, error) {
	ctx := context.Background()

	var mcpClient *client.Client
	var err error

	if token != "" {
		headers := map[string]string{"Authorization": "Bearer " + token}
		mcpClient, err = client.NewStreamableHttpClient(serverURL,
			transport.WithHTTPHeaders(headers),
		)
		if err != nil {
			mcpClient, err = client.NewSSEMCPClient(serverURL,
				transport.WithHeaders(headers),
			)
		}
	} else {
		mcpClient, err = client.NewStreamableHttpClient(serverURL)
		if err != nil {
			mcpClient, err = client.NewSSEMCPClient(serverURL)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("create MCP client: %w", err)
	}

	if err := mcpClient.Start(ctx); err != nil {
		return nil, fmt.Errorf("start MCP client: %w", err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: sys.Name, Version: sys.Version}
	initReq.Params.Capabilities = mcp.ClientCapabilities{}

	if _, err := mcpClient.Initialize(ctx, initReq); err != nil {
		_ = mcpClient.Close()
		return nil, fmt.Errorf("initialize MCP client: %w", err)
	}

	return mcpClient, nil
}

func fetchMCPTools(serverURL, token string) (map[string]LLMTool, error) {
	mcpClient, err := getMCPClient(serverURL, token)
	if err != nil {
		return nil, err
	}

	result, err := mcpClient.ListTools(context.Background(), mcp.ListToolsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list MCP tools: %w", err)
	}

	tools := make(map[string]LLMTool)
	for _, t := range result.Tools {
		toolName := t.Name
		schema, err := json.Marshal(t.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("marshal tool schema for %s: %w", toolName, err)
		}
		tools[toolName] = LLMTool{
			Description: t.Description,
			Parameters:  json.RawMessage(schema),
			Callback: func(args string) (string, error) {
				return callMCPTool(serverURL, token, toolName, args)
			},
		}
	}
	return tools, nil
}

func callMCPTool(serverURL, token, toolName, args string) (string, error) {
	mcpClient, err := getMCPClient(serverURL, token)
	if err != nil {
		return "", err
	}

	var arguments map[string]any
	if err := json.Unmarshal([]byte(args), &arguments); err != nil {
		return "", fmt.Errorf("unmarshal tool args: %w", err)
	}

	req := mcp.CallToolRequest{}
	req.Params.Name = toolName
	req.Params.Arguments = arguments

	result, err := mcpClient.CallTool(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("call MCP tool %s: %w", toolName, err)
	}
	if result.IsError {
		var sb strings.Builder
		for _, c := range result.Content {
			if t, ok := c.(mcp.TextContent); ok {
				sb.WriteString(t.Text)
			}
		}
		return "", fmt.Errorf("tool %s error: %s", toolName, sb.String())
	}

	var sb strings.Builder
	for _, c := range result.Content {
		if t, ok := c.(mcp.TextContent); ok {
			sb.WriteString(t.Text)
		}
	}
	return sb.String(), nil
}

func fetchMCPPrompts(serverURL, token string) (map[string]string, error) {
	mcpClient, err := getMCPClient(serverURL, token)
	if err != nil {
		return nil, err
	}

	result, err := mcpClient.ListPrompts(context.Background(), mcp.ListPromptsRequest{})
	if err != nil {
		return nil, fmt.Errorf("list MCP prompts: %w", err)
	}

	prompts := make(map[string]string)
	for _, p := range result.Prompts {
		prompts[p.Name] = p.Description
	}
	return prompts, nil
}

func getMCPPrompt(serverURL, token, promptName string, args map[string]string) (string, error) {
	mcpClient, err := getMCPClient(serverURL, token)
	if err != nil {
		return "", err
	}

	req := mcp.GetPromptRequest{}
	req.Params.Name = promptName
	req.Params.Arguments = args

	result, err := mcpClient.GetPrompt(context.Background(), req)
	if err != nil {
		return "", fmt.Errorf("get MCP prompt %s: %w", promptName, err)
	}

	var sb strings.Builder
	for _, msg := range result.Messages {
		if t, ok := msg.Content.(mcp.TextContent); ok {
			sb.WriteString(t.Text)
			sb.WriteString("\n")
		}
	}
	return strings.TrimSpace(sb.String()), nil
}
