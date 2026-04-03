// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"crypto/md5"
	"fmt"
	"log/slog"
	"os"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/google"
	"github.com/eja/pbx/i18n"
	"github.com/eja/pbx/openai"
	"github.com/eja/pbx/sys"
)

func TTS(rawText, language, fileAudioOutput string) error {
	text, err := sys.MarkdownToText([]byte(rawText))
	if err != nil {
		return err
	}

	aiSettings := db.Settings()
	ttsHash := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	ttsCacheFile := sys.Options.Cache + "/tts." + ttsHash
	if _, err := os.Stat(ttsCacheFile); err != nil {
		if aiSettings["ttsProvider"] == "google" {
			if err = google.TTS(fileAudioOutput, text, i18n.LanguageCodeToLocale(language)); err != nil {
				return err
			}
		} else {
			if err = openai.TTS(fileAudioOutput, text, i18n.LanguageCodeToLocale(language)); err != nil {
				return err
			}
		}
		if _, err := os.Stat(fileAudioOutput); err != nil {
			return fmt.Errorf("[core] tts file not found")
		}
		if err := sys.FileCopy(fileAudioOutput, ttsCacheFile); err != nil {
			return err
		}
	} else {
		if err := sys.FileCopy(ttsCacheFile, fileAudioOutput); err != nil {
			return err
		} else {
			slog.Debug("TTS, using cache", "file", ttsCacheFile)
		}
	}
	slog.Debug("TTS, processed", "language", language, "text", text)
	return nil
}
