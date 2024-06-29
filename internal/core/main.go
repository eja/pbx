// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"fmt"

	"github.com/eja/tibula/log"
	"pbx/internal/db"
	"pbx/internal/meta"
	"pbx/internal/sys"
	"pbx/internal/telegram"
)

const tag = "[core]"
const maxAudioInputTime = 60
const maxAudioOutputSize = 50 * 1000

func Text(userId string, language string, text string) (response string, err error) {
	platform := "chat"
	var tags []string

	response, err = Chat(platform, userId, text, language)
	if err != nil {
		return
	}
	response, tags = TagsExtract(response)
	response, err = TagsProcess(platform, language, userId, response, tags)

	return
}

func Audio(platform string, userId string, language string, chatId string, mediaId string, tts bool) (response string, err error) {
	var transcript string
	var tags []string

	aiSettings := db.Settings()
	if tts && aiSettings["ttsProvider"] == "" {
		tts = false
		log.Warn(tag, "tts provider empty")

	}
	mediaPath := fmt.Sprintf("%s/%s", sys.Options.MediaPath, mediaId)
	if language == "" {
		language = aiSettings["language"]
	}

	fileAudioInput := mediaPath + ".original.audio.in"
	if platform == "meta" {
		if err = meta.MediaGet(mediaId, fileAudioInput); err != nil {
			return
		}
	}
	if platform == "telegram" {
		if err = telegram.MediaGet(mediaId, fileAudioInput); err != nil {
			return
		}
	}

	transcript, err = ASR(fileAudioInput, language)
	if err != nil {
		return
	}

	responseRaw, err := Chat("chat", userId, transcript, language)
	if err != nil {
		return
	}
	response, tags = TagsExtract(responseRaw)
	response, err = TagsProcess(platform, language, userId, response, tags)
	if err != nil {
		return
	}

	if !tts || len(response) > maxAudioOutputSize {
		return
	}

	fileAudioOutput := mediaPath + ".audio.out"
	ttsLanguage := FilterLanguage(tags, language)
	if err = TTS(response, ttsLanguage, fileAudioOutput); err != nil {
		return
	}

	if platform == "meta" {
		if err = meta.SendAudio(userId, fileAudioOutput); err != nil {
			return
		}
		response = ""
	}
	if platform == "telegram" {
		if err = telegram.SendAudio(chatId, fileAudioOutput, response); err != nil {
			return
		}
		response = ""
	}

	return
}
