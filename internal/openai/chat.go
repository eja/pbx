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

const llmUrl = "https://api.openai.com/v1/chat/completions"
const llmModel = "gpt-4o-mini"

func Chat(messages []sys.TypeChatMessage, system string) (string, error) {
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
