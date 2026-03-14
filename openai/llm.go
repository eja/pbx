// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"strings"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/sys"
)

const llmUrl = "https://api.openai.com/v1/chat/completions"
const llmModel = "gpt-4o-mini"

type mcpRequest struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
	ID      int    `json:"id"`
}

type mcpResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *mcpError       `json:"error,omitempty"`
	ID      json.RawMessage `json:"id"`
}

type mcpError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type mcpTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

type mcpListToolsResult struct {
	Tools []mcpTool `json:"tools"`
}

type mcpCallToolParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type mcpCallToolResult struct {
	Content []mcpContent `json:"content"`
	IsError bool         `json:"isError,omitempty"`
}

type mcpContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func LLM(messages []sys.TypeChatMessage, system string, tools map[string]LLMTool) (string, error) {
	aiSettings := db.Settings()
	url := sys.Options.LlmUrl
	if url == "" {
		url = aiSettings["llmUrl"]
	}
	if url == "" {
		url = llmUrl
	}
	model := aiSettings["llmModel"]
	if model == "" {
		model = llmModel
	}
	token := aiSettings["llmToken"]
	if token == "" {
		token = sys.Options.AiToken
	}
	mcpURL := aiSettings["mcpUrl"]
	mcpToken := aiSettings["mcpToken"]

	finalTools := make(map[string]LLMTool)
	maps.Copy(finalTools, tools)

	if mcpURL != "" {
		mcpTools, err := fetchMCPTools(mcpURL, mcpToken)
		if err != nil {
			return "", fmt.Errorf("failed to fetch MCP tools: %w", err)
		}
		for name, t := range mcpTools {
			finalTools["mcp_"+name] = t
		}
	}

	var reqMessages []llmRequestMessage
	reqMessages = append(reqMessages, llmRequestMessage{
		Role:    "system",
		Content: system,
	})
	for _, m := range messages {
		reqMessages = append(reqMessages, llmRequestMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	var oaTools []llmTool
	for name, t := range finalTools {
		oaTools = append(oaTools, llmTool{
			Type: "function",
			Function: llmToolFunction{
				Name:        name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	client := &http.Client{}
	const maxIterations = 5
	for range maxIterations {
		reqData := map[string]any{
			"stream":   false,
			"model":    model,
			"messages": reqMessages,
		}
		if len(oaTools) > 0 {
			reqData["tools"] = oaTools
		}

		payload, err := json.Marshal(reqData)
		if err != nil {
			return "", fmt.Errorf("error marshaling JSON: %v", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return "", fmt.Errorf("error creating request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("error making request: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			resp.Body.Close()
			return "", fmt.Errorf("LLM HTTP %d: %s", resp.StatusCode, string(body))
		}

		var response llmResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("error decoding JSON response: %v", err)
		}

		if len(response.Choices) == 0 {
			return "", fmt.Errorf("no choices returned in response")
		}

		msg := response.Choices[0].Message

		if len(msg.ToolCalls) == 0 {
			return msg.Content, nil
		}

		reqMessages = append(reqMessages, llmRequestMessage{
			Role:      "assistant",
			Content:   msg.Content,
			ToolCalls: msg.ToolCalls,
		})

		for _, tc := range msg.ToolCalls {
			funcName := tc.Function.Name
			args := tc.Function.Arguments

			var result string
			if tool, ok := finalTools[funcName]; ok {
				res, err := tool.Callback(args)
				if err != nil {
					result = fmt.Sprintf("Error executing function: %v", err)
				} else {
					result = res
				}
			} else {
				result = fmt.Sprintf("Error: function '%s' not found", funcName)
			}

			reqMessages = append(reqMessages, llmRequestMessage{
				Role:       "tool",
				ToolCallID: tc.ID,
				Name:       funcName,
				Content:    result,
			})
		}
	}

	return "", fmt.Errorf("exceeded max tool call iterations")
}

func parseSSEOrJSON(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	trimmed := bytes.TrimSpace(data)
	if len(trimmed) > 0 && trimmed[0] == '{' {
		return trimmed, nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		if after, ok := strings.CutPrefix(line, "data: "); ok {
			payload := after
			if payload != "[DONE]" {
				return []byte(payload), nil
			}
		}
	}

	preview := string(data)
	if len(preview) > 200 {
		preview = preview[:200]
	}
	return nil, fmt.Errorf("no JSON data found in response: %s", preview)
}

func fetchMCPTools(serverURL, token string) (map[string]LLMTool, error) {
	client := &http.Client{}

	listReq := mcpRequest{
		JSONRPC: "2.0",
		Method:  "tools/list",
		Params:  struct{}{},
		ID:      1,
	}
	listBody, err := json.Marshal(listReq)
	if err != nil {
		return nil, fmt.Errorf("marshal tools/list request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(listBody))
	if err != nil {
		return nil, fmt.Errorf("create tools/list request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("tools/list request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("tools/list HTTP %d: %s", resp.StatusCode, string(body))
	}

	rawJSON, err := parseSSEOrJSON(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("parse tools/list response: %w", err)
	}

	var listResp mcpResponse
	if err := json.Unmarshal(rawJSON, &listResp); err != nil {
		return nil, fmt.Errorf("decode tools/list response: %w", err)
	}
	if listResp.Error != nil {
		return nil, fmt.Errorf("MCP error: code %d, message: %s", listResp.Error.Code, listResp.Error.Message)
	}

	var listResult mcpListToolsResult
	if err := json.Unmarshal(listResp.Result, &listResult); err != nil {
		return nil, fmt.Errorf("unmarshal tools/list result: %w", err)
	}

	tools := make(map[string]LLMTool)
	for _, mt := range listResult.Tools {
		tool := mt
		tools[tool.Name] = LLMTool{
			Description: tool.Description,
			Parameters:  tool.InputSchema,
			Callback: func(args string) (string, error) {
				return callMCPTool(serverURL, token, tool.Name, json.RawMessage(args))
			},
		}
	}
	return tools, nil
}

func callMCPTool(serverURL, token, toolName string, args json.RawMessage) (string, error) {
	client := &http.Client{}

	callReq := mcpRequest{
		JSONRPC: "2.0",
		Method:  "tools/call",
		Params: mcpCallToolParams{
			Name:      toolName,
			Arguments: args,
		},
		ID: 1,
	}
	callBody, err := json.Marshal(callReq)
	if err != nil {
		return "", fmt.Errorf("marshal tools/call request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", serverURL, bytes.NewBuffer(callBody))
	if err != nil {
		return "", fmt.Errorf("create tools/call request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json, text/event-stream")
	if token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("tools/call request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("tools/call HTTP %d: %s", resp.StatusCode, string(body))
	}

	rawJSON, err := parseSSEOrJSON(resp.Body)
	if err != nil {
		return "", fmt.Errorf("parse tools/call response: %w", err)
	}

	var callResp mcpResponse
	if err := json.Unmarshal(rawJSON, &callResp); err != nil {
		return "", fmt.Errorf("decode tools/call response: %w", err)
	}
	if callResp.Error != nil {
		return "", fmt.Errorf("MCP error: code %d, message: %s", callResp.Error.Code, callResp.Error.Message)
	}

	var callResult mcpCallToolResult
	if err := json.Unmarshal(callResp.Result, &callResult); err != nil {
		return "", fmt.Errorf("unmarshal tools/call result: %w", err)
	}

	var text string
	for _, content := range callResult.Content {
		if content.Type == "text" {
			text += content.Text
		}
	}
	if callResult.IsError {
		return "", fmt.Errorf("tool returned error: %s", text)
	}
	return text, nil
}
