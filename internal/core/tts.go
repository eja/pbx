// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"crypto/md5"
	"fmt"
	"os"

	"github.com/eja/tibula/log"
	"pbx/internal/db"
	"pbx/internal/google"
	"pbx/internal/i18n"
	"pbx/internal/openai"
	"pbx/internal/sys"
)

func TTS(text, language, fileAudioOutput string) error {
	aiSettings := db.Settings()
	ttsHash := fmt.Sprintf("%x", md5.Sum([]byte(text)))
	ttsCacheFile := sys.Options.Cache + "/tts." + ttsHash
	if _, err := os.Stat(ttsCacheFile); err != nil {
		if aiSettings["ttsProvider"] == "google" {
			if err = google.TTS(fileAudioOutput, text, i18n.LanguageCodeToLocale(language)); err != nil {
				return err
			}
		}
		if aiSettings["ttsProvider"] == "openai" {
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
			log.Trace(tag, "tts using cache for", ttsCacheFile)
		}
	}
	return nil
}
