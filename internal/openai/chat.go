// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"pbx/internal/db"
	"pbx/internal/sys"
)

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

func Chat(messages []sys.TypeChatMessage, system string) (string, error) {
	aiSettings := db.Settings()
	url := aiSettings["openaiUrl"]
	if url == "" {
		url = sys.Options.OpenaiUrl
	}
	model := aiSettings["openaiModel"]
	if model == "" {
		model = sys.Options.OpenaiModel
	}
	token := aiSettings["openaiToken"]
	if token == "" {
		token = sys.Options.OpenaiToken
	}

	// Populate the system prompt
	messageSystem := sys.TypeChatMessage{
		Role:    "system",
		Content: system,
	}

	// Define the request payload
	payload, err := json.Marshal(map[string]interface{}{
		"stream":   false,
		"model":    model,
		"messages": append([]sys.TypeChatMessage{messageSystem}, messages...),
	})
	if err != nil {
		return "", fmt.Errorf("Error marshaling JSON: %v", err)
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	// Parse the response
	var response typeChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("Error decoding JSON response: %v", err)
	}

	// Check if there's a valid assistant message
	if len(response.Choices) > 0 {
		assistantMessage := response.Choices[0].Message
		return assistantMessage.Content, nil
	}

	return "", nil
}
