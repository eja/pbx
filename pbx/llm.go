// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/sys"
	"github.com/eja/tibula/log"
)

type LLMTool struct {
	Description string
	Parameters  any
	Callback    func(arguments string) (string, error)
}

type llmTool struct {
	Type     string          `json:"type"`
	Function llmToolFunction `json:"function"`
}

type llmToolFunction struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Parameters  any    `json:"parameters,omitempty"`
}

type llmToolCall struct {
	ID       string              `json:"id"`
	Type     string              `json:"type"`
	Function llmToolCallFunction `json:"function"`
}

type llmToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type llmRequestMessage struct {
	Role       string        `json:"role"`
	Content    string        `json:"content"`
	Name       string        `json:"name,omitempty"`
	ToolCallID string        `json:"tool_call_id,omitempty"`
	ToolCalls  []llmToolCall `json:"tool_calls,omitempty"`
}

type llmResponseMessage struct {
	Role      string        `json:"role"`
	Content   string        `json:"content"`
	ToolCalls []llmToolCall `json:"tool_calls,omitempty"`
}

type llmResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int                `json:"index"`
		Message      llmResponseMessage `json:"message"`
		Logprobs     any                `json:"logprobs"`
		FinishReason string             `json:"finish_reason"`
	} `json:"choices"`
}

func LLM(messages []sys.TypeChatMessage, system string, tools map[string]LLMTool) (string, error) {
	aiSettings := db.Settings()
	url := sys.Options.LlmUrl
	if url == "" {
		url = aiSettings["llmUrl"]
	}
	if url == "" {
		return "", fmt.Errorf("llm url cannot be empty")
	}
	model := aiSettings["llmModel"]
	token := aiSettings["llmToken"]
	if token == "" {
		token = sys.Options.AiToken
	}
	mcpURL := aiSettings["mcpUrl"]
	mcpToken := aiSettings["mcpToken"]

	finalTools := make(map[string]LLMTool)
	maps.Copy(finalTools, tools)

	if mcpURL != "" {
		mcpPrompts, err := fetchMCPPrompts(mcpURL, mcpToken)
		if err != nil {
			log.Warn(tag, "[llm] failed to fetch MCP prompts", err)
		} else {
			for promptName := range mcpPrompts {
				promptText, err := getMCPPrompt(mcpURL, mcpToken, promptName, nil)
				if err == nil && promptText != "" {
					if system != "" {
						system = system + "\n\n" + promptText
					} else {
						system = promptText
					}
				}
			}
		}

		mcpTools, err := fetchMCPTools(mcpURL, mcpToken)
		if err != nil {
			log.Warn(tag, "[llm] failed to fetch MCP tools", err)
		} else {
			for name, t := range mcpTools {
				finalTools["mcp_"+name] = t
			}
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

	httpClient := &http.Client{}
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

		resp, err := httpClient.Do(req)
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
