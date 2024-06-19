// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"encoding/json"
	"net/http"

	"github.com/eja/tibula/log"
	"pbx/internal/core"
	"pbx/internal/db"
	"pbx/internal/i18n"
	"pbx/internal/meta"
	"pbx/internal/sys"
)

type typeMetaMessage struct {
	Entry []struct {
		Changes []struct {
			Value struct {
				Messages []struct {
					From string `json:"from"`
					Text *struct {
						Body string `json:"body"`
					} `json:"text,omitempty"`
					Audio *struct {
						ID string `json:"id"`
					} `json:"audio,omitempty"`
					ID string `json:"id"`
				} `json:"messages"`
			} `json:"value"`
		} `json:"changes"`
	} `json:"entry"`
}

func metaRouter(w http.ResponseWriter, r *http.Request) {
	if err := db.Open(); err != nil {
		return
	}

	if r.Method == http.MethodGet {
		hubMode := r.URL.Query().Get("hub.mode")
		verifyToken := r.URL.Query().Get("hub.verify_token")
		if hubMode == "subscribe" && (verifyToken == sys.Options.MetaToken || verifyToken == db.Settings()["metaToken"]) {
			w.Write([]byte(r.URL.Query().Get("hub.challenge")))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Token verification error"))
		}
	} else {
		const platform = "meta"
		var metaMessage typeMetaMessage

		err := json.NewDecoder(r.Body).Decode(&metaMessage)
		if err != nil {
			errMessage := "Error decoding request body"
			http.Error(w, errMessage, http.StatusBadRequest)
			log.Warn("[FB]", errMessage)
			return
		}

		log.Trace("[FB]", "incoming message", metaMessage)
		if len(metaMessage.Entry) > 0 {
			changes := metaMessage.Entry[0].Changes
			if len(changes) > 0 {
				value := changes[0].Value

				if len(value.Messages) > 0 {
					message := value.Messages[0]
					userId := message.From
					chatId := message.ID
					aiSettings := db.Settings()

					user, err := db.UserGet(userId)
					if user == nil && db.Number(aiSettings["userRestricted"]) < 1 {
						user = make(map[string]string)
						user["welcome"] = "1"
						user["language"] = aiSettings["language"]
						user["audio"] = "2"
					}
					if aiSettings["asrProvider"] == "" {
						user["audio"] = "0"
					}

					if err == nil && user != nil {
						if err := meta.SendStatus(chatId, "read"); err != nil {
							log.Warn("[FB]", "status", userId, chatId, err)
						}

						if db.Number(user["welcome"]) < 1 {
							_, actionResponse := db.ChatAction("/welcome", user["language"])
							meta.SendText(userId, actionResponse)
							db.UserUpdate(userId, "welcome", "1")
						}

						if message.Text != nil && message.Text.Body != "" {
							response, err := core.Text(userId, user["language"], message.Text.Body)
							if err != nil {
								response = i18n.Translate(user["language"], "error")
								log.Warn("[FB]", userId, chatId, err)
							}
							if err := meta.SendText(userId, response); err != nil {
								log.Warn("[FB]", userId, err)
							}
						} else if message.Audio != nil {
							if db.Number(user["audio"]) > 0 {
								response, err := core.Audio(
									platform,
									userId,
									user["language"],
									chatId,
									message.Audio.ID,
									db.Number(user["audio"]) > 1,
								)
								if err != nil {
									log.Warn("[FB]", userId, chatId, err)
									if err := meta.SendText(userId, i18n.Translate(user["language"], "error")); err != nil {
										log.Warn("[FB]", userId, chatId, err)
									}
								}
								if response != "" {
									if err := meta.SendText(userId, response); err != nil {
										log.Warn("[FB]", userId, chatId, err)
									}
								}
							} else {
								if err := meta.SendText(userId, i18n.Translate(user["language"], "audio_disabled")); err != nil {
									log.Warn("[FB]", userId, chatId, err)
								}
							}
						}
					} else {
						if err := meta.SendText(userId, i18n.Translate("", "user_unknown")); err != nil {
							log.Warn("[FB]", userId, chatId, err)
						}
					}
				}
			}
		}
		w.WriteHeader(http.StatusOK)
	}
}
