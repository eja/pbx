// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package google

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"pbx/internal/db"
	"pbx/internal/sys"
)

type TTSRequest struct {
	Input struct {
		Text string `json:"text"`
	} `json:"input"`
	AudioConfig struct {
		AudioEncoding string `json:"audioEncoding"`
	} `json:"audioConfig"`
	Voice struct {
		SsmlGender   string `json:"ssmlGender"`
		LanguageCode string `json:"languageCode"`
	} `json:"voice"`
}

func TTS(filePath string, text string, languageCode string) error {
	aiSettings := db.Settings()
	apiKey := aiSettings["ttsToken"]
	if apiKey == "" {
		apiKey = sys.Options.AiToken
	}
	gender := aiSettings["ttsGender"]
	if gender == "" {
		gender = "FEMALE"
	}

	requestBody := TTSRequest{}
	requestBody.Input.Text = text
	requestBody.AudioConfig.AudioEncoding = "OGG_OPUS"
	requestBody.Voice.SsmlGender = gender
	requestBody.Voice.LanguageCode = languageCode

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request body: %v", err)
	}

	url := fmt.Sprintf("https://texttospeech.googleapis.com/v1/text:synthesize?key=%s", apiKey)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %v", err)
	}

	var respData struct {
		AudioContent string `json:"audioContent"`
	}
	err = json.Unmarshal(respBody, &respData)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %v", err)
	}

	if respData.AudioContent == "" {
		return fmt.Errorf("no audio content in response")
	}

	audioData, err := base64.StdEncoding.DecodeString(respData.AudioContent)
	if err != nil {
		return fmt.Errorf("failed to decode audio content: %v", err)
	}

	err = ioutil.WriteFile(filePath, audioData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write audio file: %v", err)
	}

	return nil
}
