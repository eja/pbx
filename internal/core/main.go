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

func Text(userId string, language string, text string) (string, error) {
	response, err := Chat("chat", userId, text, language)
	if err == nil {
		response, _ = TagsExtract(response)
	}
	return response, err
}

func Audio(platform string, userId string, language string, chatId string, mediaId string, tts bool) (string, error) {
	var response string
	var transcript string
	var tags []string
	var err error

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
		if err := meta.MediaGet(mediaId, fileAudioInput); err != nil {
			return "", err
		}
	}
	if platform == "telegram" {
		if err := telegram.MediaGet(mediaId, fileAudioInput); err != nil {
			return "", err
		}
	}

	transcript, err = ASR(fileAudioInput, language)
	if err != nil {
		return "", nil
	}

	responseRaw, err := Chat("chat", userId, transcript, language)
	if err != nil {
		return "", err
	}
	response, tags = TagsExtract(responseRaw)

	if !tts || len(response) > maxAudioOutputSize {
		return response, nil
	}

	fileAudioOutput := mediaPath + ".audio.out"
	ttsLanguage := FilterLanguage(tags, language)
	if err := TTS(response, ttsLanguage, fileAudioOutput); err != nil {
		return "", nil
	}

	if platform == "meta" {
		if err := meta.SendAudio(userId, fileAudioOutput); err != nil {
			return "", err
		}
		response = ""
	}
	if platform == "telegram" {
		if err := telegram.SendAudio(chatId, fileAudioOutput, response); err != nil {
			return "", err
		}
		response = ""
	}

	return response, nil
}
