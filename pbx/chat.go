// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"fmt"
	"strings"
	"time"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/i18n"
	"github.com/eja/pbx/sys"
)

const historyTimeout = 300

var history map[string][]sys.TypeChatMessage
var historyTime map[string]time.Time
var historyThread map[string]string
var historyInit bool

var ChatProcess func(userId, system, message, language string, history []sys.TypeChatMessage, tools map[string]LLMTool) (string, error)

func Chat(platform, userId, message, language string) (string, error) {
	var response, system, assistant string
	var err error
	aiSettings := db.Settings()
	action := false

	timeZoneDiff := time.Duration(sys.Number(aiSettings["timezone"]))
	timeZoneNow := time.Now().Add(timeZoneDiff * time.Hour).Format("Monday 02 of January 2006, 03:04 pm")

	log().Debug("chat request", "language", language, "user", userId, "message", message)

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
			if Plugins[actionFunction] != nil {
				response = Plugins[actionFunction](userId, language, message, actionResponse)
			} else {
				log().Warn("chat plugin not found", "message", message)
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

		if ChatProcess != nil {
			assistant, err = ChatProcess(userId, system, message, language, history[userId], Tools)
		} else {
			assistant, err = LLM(history[userId], system, Tools)
		}
		if err != nil {
			log().Error("chat error", "error", err)
			return "", err
		}
		response = assistant

		historyTime[userId] = time.Now()
		history[userId] = append(history[userId], sys.TypeChatMessage{
			Role:    "assistant",
			Content: response,
		})
	}

	log().Debug("chat response", "language", language, "user", userId, "response", response)
	if !action {
		db.Log(userId, response)
	}

	return response, nil
}
