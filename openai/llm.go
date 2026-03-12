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

func LLM(messages []sys.TypeChatMessage, system string) (string, error) {
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

	messageSystem := sys.TypeChatMessage{
		Role:    "system",
		Content: system,
	}

	payload, err := json.Marshal(map[string]any{
		"stream":   false,
		"model":    model,
		"messages": append([]sys.TypeChatMessage{messageSystem}, messages...),
	})
	if err != nil {
		return "", fmt.Errorf("Error marshaling JSON: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return "", fmt.Errorf("Error creating request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("Error making request: %v", err)
	}
	defer resp.Body.Close()

	var response typeChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("Error decoding JSON response: %v", err)
	}

	if len(response.Choices) > 0 {
		assistantMessage := response.Choices[0].Message
		return assistantMessage.Content, nil
	}

	return "", nil
}
