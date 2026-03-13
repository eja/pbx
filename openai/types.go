// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

const tag = "[openai]"

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

type ttsRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format"`
	Locale string `json:"locale,omitempty"`
}

type asrResponse struct {
	Text string `json:"text"`
}
