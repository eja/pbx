// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/eja/tibula/log"
	"pbx/internal/db"
	"pbx/internal/sys"
)

const assistantUrl = "https://api.openai.com/v1"

var assistantThreadId map[string]string

func Assistant(message, system, threadIn string) (response, thread string, err error) {
	var data []byte
	aiSettings := db.Settings()
	url := aiSettings["llmUrl"]
	if url == "" {
		url = assistantUrl
	}
	model := aiSettings["llmModel"]
	if model == "" {
		model = llmModel
	}
	token := aiSettings["llmToken"]
	if token == "" {
		token = sys.Options.AiToken
	}

	thread = threadIn
	if thread == "" {
		data, err = assistantPost(url+"/threads", token, []byte("{}"))
		if err != nil {
			return "", "", err
		} else {
			var jsonResponse map[string]interface{}
			if err = json.Unmarshal(data, &jsonResponse); err != nil {
				return
			}
			if idValue, ok := jsonResponse["id"].(string); ok {
				thread = idValue
			}
		}
	}
	if thread == "" {
		err = fmt.Errorf("no valid thread id")
		return
	}

	run := typeAssistantRequest{
		AssistantId:            model,
		AdditionalInstructions: system,
		Stream:                 true,
		AdditionalMessages: []typeAssistantRequestMessage{
			{
				Role: "user",
				Content: []typeAssistantRequestAdditionalMessage{
					{Type: "text", Text: message},
				},
			},
		},
	}
	payload, err := json.Marshal(run)
	if err != nil {
		err = err
		return
	}
	data, err = assistantPost(url+"/threads/"+thread+"/runs", token, []byte(payload))
	if err != nil {
		return
	}
	re := regexp.MustCompile(`event:\s*thread.message.completed\s*data:\s*(.*)`)
	matches := re.FindStringSubmatch(string(data))

	if len(matches) > 1 {
		log.Trace(tag, "assistant response", url, matches[1])
		var jsonResponse typeAssistantResponse
		if err = json.Unmarshal([]byte(matches[1]), &jsonResponse); err != nil {
			return
		}
		if len(jsonResponse.Content) > 0 {
			response = jsonResponse.Content[0].Text.Value
			log.Debug(tag, "annotations", jsonResponse.Content[0].Text.Annotations)
			for _, row := range jsonResponse.Content[0].Text.Annotations {
				response = assistantReplaceAnnotation(response, row.StartIndex, row.EndIndex, ' ')
			}
		}
	}

	return
}

func assistantPost(url, token string, payload []byte) (response []byte, err error) {
	log.Trace(tag, "assistant request", url, string(payload))
	client := &http.Client{}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return response, fmt.Errorf("failed http request: %s", resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

func assistantReplaceAnnotation(s string, start, stop int, replace rune) string {
	if start < 0 {
		start = 0
	}
	if stop > len(s) {
		stop = len(s)
	}
	if start > stop {
		return s
	}

	runes := []rune(s)

	for i := start; i < stop && i < len(runes); i++ {
		runes[i] = replace
	}

	return string(runes)
}
