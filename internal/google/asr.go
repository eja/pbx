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

func ASR(filePath, language string) (string, error) {
	aiSettings := db.Settings()
	apiKey := aiSettings["llmToken"]
	if apiKey == "" {
		apiKey = sys.Options.AiToken
	}

	audioBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read audio file: %v", err)
	}

	audioBase64 := base64.StdEncoding.EncodeToString(audioBytes)

	payload := map[string]interface{}{
		"audio": map[string]string{
			"content": audioBase64,
		},
		"config": map[string]interface{}{
			"encoding":        "WEBM_OPUS",
			"sampleRateHertz": 48000,
			"languageCode":    language,
		},
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal JSON payload: %v", err)
	}

	apiURL := fmt.Sprintf("https://speech.googleapis.com/v1/speech:recognize?key=%s", apiKey)

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %v", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send HTTP request: %v", err)
	}
	defer resp.Body.Close()

	respBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("received non-200 response: %s", string(respBytes))
	}

	var respData map[string]interface{}
	err = json.Unmarshal(respBytes, &respData)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response JSON: %v", err)
	}

	results, ok := respData["results"].([]interface{})
	if !ok {
		return "", nil
	}

	firstResult, ok := results[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid result format")
	}

	alternatives, ok := firstResult["alternatives"].([]interface{})
	if !ok || len(alternatives) == 0 {
		return "", fmt.Errorf("no alternatives found in the result")
	}

	firstAlternative, ok := alternatives[0].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid alternative format")
	}

	transcript, ok := firstAlternative["transcript"].(string)
	if !ok {
		return "", fmt.Errorf("transcript not found in the alternative")
	}

	return transcript, nil
}
