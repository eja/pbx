// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/eja/tibula/log"
	"pbx/internal/db"
	"pbx/internal/meta"
	"pbx/internal/sys"
	"pbx/internal/telegram"
)

const tag = "[core]"
const maxAudioInputTime = 60
const maxAudioOutputSize = 50 * 1000

func TagsExtract(text string) (string, []string) {
	re := regexp.MustCompile(`\s*\[([^\]]+)\]\s*$`)
	var tags []string

	for {
		matches := re.FindStringSubmatchIndex(text)
		if len(matches) == 0 {
			break
		}
		tag := text[matches[2]:matches[3]]
		tags = append(tags, tag)
		text = strings.TrimSuffix(text[:matches[0]], " ["+tag)
	}

	return text, tags
}

func FilterLanguage(tags []string, language string) string {
	re := regexp.MustCompile(`\[\w\w\]`)
	for _, tag := range tags {
		if re.MatchString(tag) {
			language = tag[1:3]
		}
	}
	return language
}

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
