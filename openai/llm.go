// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/sys"
)

const llmUrl = "https://api.openai.com/v1/chat/completions"
const llmModel = "gpt-4o-mini"

func LLM(messages []sys.TypeChatMessage, system string, tools map[string]LLMTool) (string, error) {
	aiSettings := db.Settings()
	url := aiSettings["llmUrl"]
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
	for name, t := range tools {
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
			return "", fmt.Errorf("Error marshaling JSON: %v", err)
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
		if err != nil {
			return "", fmt.Errorf("Error creating request: %v", err)
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := client.Do(req)
		if err != nil {
			return "", fmt.Errorf("Error making request: %v", err)
		}

		var response llmResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("Error decoding JSON response: %v", err)
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
			if tool, ok := tools[funcName]; ok {
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
