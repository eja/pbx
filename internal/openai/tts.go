// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"pbx/internal/db"
	"pbx/internal/sys"
)

const ttsModel = "tts-1"
const ttsVoice = "alloy"
const ttsUrl = "https://api.openai.com/v1/audio/speech"

type typeTTSRequest struct {
	Model  string `json:"model"`
	Input  string `json:"input"`
	Voice  string `json:"voice"`
	Format string `json:"response_format"`
	Locale string `json:"locale,omitempty"`
}

func TTS(filePath string, text string, languageCode string) error {
	const audioType = "opus"
	aiSettings := db.Settings()
	model := aiSettings["ttsModel"]
	if model == "" {
		model = ttsModel
	}
	token := aiSettings["ttsToken"]
	if token == "" {
		token = aiSettings["llmToken"]
		if token == "" {
			token = sys.Options.AiToken
		}
	}
	url := aiSettings["ttsUrl"]
	if url == "" {
		url = ttsUrl
	}
	voice := aiSettings["ttsVoice"]
	if voice == "" {
		voice = ttsVoice
	}

	requestBody := typeTTSRequest{
		Model:  model,
		Input:  text,
		Voice:  voice,
		Format: audioType,
		Locale: languageCode,
	}

	jsonValue, _ := json.Marshal(requestBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to do request: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	err = ioutil.WriteFile(filePath, body, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}
