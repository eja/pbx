// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"pbx/internal/db"
	"pbx/internal/sys"
)

const asrModel = "whisper-1"
const asrUrl = "https://api.openai.com/v1/audio/transcriptions"

type typeASRResponse struct {
	Text string `json:"text"`
}

func ASR(filePath string, languageCode string) (string, error) {
	const audioType = "ogg"
	aiSettings := db.Settings()
	model := aiSettings["asrModel"]
	if model == "" {
		model = asrModel
	}
	token := aiSettings["asrToken"]
	if token == "" {
		token = aiSettings["aiToken"]
		if token == "" {
			token = sys.Options.AiToken
		}
	}
	url := aiSettings["asrUrl"]
	if url == "" {
		url = asrUrl
	}

	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.WriteField("model", model)
	writer.WriteField("language", languageCode)
	filePart, err := writer.CreatePart(map[string][]string{
		"Content-Disposition": {fmt.Sprintf("form-data; name=\"file\"; filename=\"%s.%s\"", filepath.Base(filePath), audioType)},
		"Content-Type":        {fmt.Sprintf("audio/%s", audioType)},
	})
	if err != nil {
		return "", fmt.Errorf("creating form file")
	}
	_, err = io.Copy(filePart, file)
	if err != nil {
		return "", fmt.Errorf("copying file into part")
	}
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("closing form writer: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, url, &body)
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	var openAIResponse typeASRResponse
	err = json.Unmarshal(responseBody, &openAIResponse)
	if err != nil {
		return "", err
	}

	return openAIResponse.Text, nil
}
