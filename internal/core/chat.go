// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/eja/tibula/log"
	"pbx/internal/db"
	"pbx/internal/google"
	"pbx/internal/i18n"
	"pbx/internal/openai"
	"pbx/internal/sys"
)

const historyTimeout = 300

var history map[string][]sys.TypeChatMessage
var historyTime map[string]time.Time
var historyThread map[string]string
var historyInit bool

func Chat(platform, userId, message, language string) (string, error) {
	var response, system, assistant string
	var err error
	aiSettings := db.Settings()
	action := false

	timeZoneDiff := time.Duration(sys.Number(aiSettings["timezone"]))
	timeZoneNow := time.Now().Add(timeZoneDiff * time.Hour).Format("Monday 02 of January 2006, 03:04 pm")

	log.Debug(tag, "chat request:", language, userId, message)

	if rows, err := db.SystemPrompt(platform); err != nil {
		return "", err
	} else {
		for _, row := range rows {
			system += row["prompt"] + "\n"
		}
	}

	system += fmt.Sprintf("Now is %s.\n", timeZoneNow)
	if platform == "pbx" {
		system += fmt.Sprintf("Users's phone number is %s.\n", userId)
	}
	system += fmt.Sprintf("The user usually speaks in %s, so please answer in that language or the language of the question if not instructed otherwise.\n", i18n.LanguageCodeToInternal(language))
	system += "Always add a new line containing the language code, 2 chars, between square brackets that you have used to answer the question at the end of your response, like this: \n\n[en]\n"

	if strings.HasPrefix(message, "/") {
		action = true
		parameters := strings.Split(message, " ")
		var actionFunction, actionResponse = db.ChatAction(platform, parameters[0], language)
		if actionFunction != "" {
			if AiChatPlugins[actionFunction] != nil {
				response = AiChatPlugins[actionFunction](userId, language, message, actionResponse)
			} else {
				log.Warn(tag, "chat plugin not found", message)
			}
		} else if actionResponse != "" {
			response = actionResponse
		}
	} else {
		db.Log(userId, message)
	}

	if response == "" {
		if !historyInit {
			history = make(map[string][]sys.TypeChatMessage)
			historyTime = make(map[string]time.Time)
			historyThread = make(map[string]string)
			historyInit = true
		}

		if aiSettings["llmProvider"] == "assistant" {

			memoryEnabled := false
			memoryThread := ""
			if sys.Bool(aiSettings["userRestricted"]) {
				if row, err := db.UserGet(userId); err == nil && sys.Bool(row["threadEnabled"]) {
					memoryEnabled = true
					memoryThread := row["thread"]
					historyThread[userId] = memoryThread
				}
			}
			if response, thread, err := openai.Assistant(message, system, historyThread[userId]); err != nil {
				log.Error(tag, err)
				return "", err
			} else {
				historyThread[userId] = thread
				if memoryEnabled && memoryThread == "" {
					db.UserUpdate(userId, "thread", thread)
				}
				return response, nil
			}

		} else {
			if hist, ok := history[userId]; ok && len(hist) > 0 && (time.Now().Sub(historyTime[userId]).Seconds() < historyTimeout) {
				history[userId] = append(history[userId], sys.TypeChatMessage{
					Role:    "user",
					Content: message,
				})
			} else {
				history[userId] = []sys.TypeChatMessage{
					{Role: "user", Content: message},
				}
			}

			if aiSettings["llmProvider"] == "google" || (aiSettings["llmProvider"] == "" && sys.Options.AiProvider == "google") {
				assistant, err = google.Chat(history[userId], system)
			} else {
				assistant, err = openai.Chat(history[userId], system)
			}
			if err != nil {
				log.Error(tag, err)
				return "", err
			}
			historyTime[userId] = time.Now()
			history[userId] = append(history[userId], sys.TypeChatMessage{
				Role:    "assistant",
				Content: assistant,
			})
			response = assistant
		}
	}

	log.Debug(tag, "chat response:", language, userId, response)
	if !action {
		db.Log(userId, response)
	}

	return response, nil
}
