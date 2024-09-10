// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"pbx/internal/sys"
)

const tag = "[openai]"

type typeTTSRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format"`
	Locale string `json:"locale,omitempty"`
}

type typeASRResponse struct {
	Text string `json:"text"`
}

type typeChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      sys.TypeChatMessage
		Logprobs     interface{} `json:"logprobs"`
		FinishReason string      `json:"finish_reason"`
	} `json:"choices"`
}

type typeAssistantRequestAdditionalMessage struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type typeAssistantRequestMessage struct {
	Role    string                                  `json:"role"`
	Content []typeAssistantRequestAdditionalMessage `json:"content"`
}

type typeAssistantRequest struct {
	AssistantId            string                        `json:"assistant_id"`
	AdditionalInstructions string                        `json:"additional_instructions"`
	AdditionalMessages     []typeAssistantRequestMessage `json:"additional_messages"`
	Stream                 bool                          `json:"stream"`
}

type typeAssistantResponse struct {
	ID        string                             `json:"id"`
	CreatedAt int64                              `json:"created_at"`
	Content   []typeAssistantResponseContentItem `json:"content"`
}

type typeAssistantResponseContentItem struct {
	Type string `json:"type"`
	Text struct {
		Value       string                            `json:"value"`
		Annotations []typeAssistantResponseAnnotation `json:"annotations"`
	} `json:"text"`
}

type typeAssistantResponseAnnotation struct {
	Type       string `json:"type"`
	Text       string `json:"text"`
	StartIndex int    `json:"start_index"`
	EndIndex   int    `json:"end_index"`
}
