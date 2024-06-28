// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eja/tibula/log"
	"pbx/internal/core"
	"pbx/internal/db"
	"pbx/internal/i18n"
	"pbx/internal/sys"
	"pbx/internal/telegram"
)

type typeTelegramMessage struct {
	Message struct {
		From struct {
			Id           int    `json:"id"`
			LanguageCode string `json:"language_code"`
		} `json:"from"`
		Chat struct {
			Id int `json:"id"`
		} `json:"chat"`
		Text  string `json:"text,omitempty"`
		Voice struct {
			FileId string `json:"file_id"`
		} `json:"voice,omitempty"`
		Context struct {
			Forwarded bool `json:"forwarded"`
		} `json:"context,omitempty"`
	} `json:"message"`
}

func telegramRouter(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		const platform = "telegram"
		var telegramMessage typeTelegramMessage

		err := json.NewDecoder(r.Body).Decode(&telegramMessage)
		if err != nil {
			errMessage := "Error decoding request body"
			http.Error(w, errMessage, http.StatusBadRequest)
			log.Warn("[TG]", errMessage)
			return
		}

		log.Trace("[TG]", "incoming message", telegramMessage)
		userId := fmt.Sprintf("TG.%d", telegramMessage.Message.From.Id)
		chatId := fmt.Sprintf("%d", telegramMessage.Message.Chat.Id)
		chatLanguage := telegramMessage.Message.From.LanguageCode
		aiSettings := db.Settings()

		user, err := db.UserGet(userId)
		if user == nil && sys.Number(aiSettings["userRestricted"]) < 1 {
			user = make(map[string]string)
			user["welcome"] = "1"
			user["language"] = aiSettings["language"]
			user["audio"] = "2"
		}
		if aiSettings["asrProvider"] == "" {
			user["audio"] = "0"
		}

		if err == nil && user != nil {
			if sys.Number(user["welcome"]) < 1 {
				_, actionResponse := db.ChatAction("chat", "/welcome", user["language"])
				telegram.SendText(chatId, actionResponse)
				db.UserUpdate(userId, "welcome", "1")
			}

			if text := telegramMessage.Message.Text; text != "" {
				response, err := core.Text(userId, user["language"], text)
				if err != nil {
					response = i18n.Translate(user["language"], "error")
					log.Warn("[TG]", userId, chatId, err)
				}
				if err := telegram.SendText(chatId, response); err != nil {
					log.Warn("[TG]", userId, chatId, err)
				}
			}

			if voice := telegramMessage.Message.Voice; voice.FileId != "" {
				if sys.Number(user["audio"]) > 0 {
					response, err := core.Audio(
						platform,
						userId,
						user["language"],
						chatId,
						voice.FileId,
						sys.Number(user["audio"]) > 1,
					)
					if err != nil {
						log.Warn("[TG]", userId, chatId, err)
						if err := telegram.SendText(chatId, i18n.Translate(chatLanguage, "error")); err != nil {
							log.Warn("[TG]", userId, chatId, err)
						}
					}
					if response != "" {
						if err := telegram.SendText(chatId, response); err != nil {
							log.Warn("[TG]", userId, chatId, err)
						}
					}
				} else {
					if err := telegram.SendText(chatId, i18n.Translate(user["language"], "audio_disabled")); err != nil {
						log.Warn("[TG]", userId, chatId, err)
					}
				}
			}
		} else {
			if err := telegram.SendText(chatId, i18n.Translate(chatLanguage, "user_unknown")); err != nil {
				log.Warn("[TG]", userId, chatId, err)
			}
		}
	}
}
