// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"pbx/internal/db"
	"pbx/internal/ff"
	"pbx/internal/google"
	"pbx/internal/i18n"
	"pbx/internal/openai"
)

func ASR(fileAudioInput, language string) (string, error) {
	aiSettings := db.Settings()
	transcript := ""
	probeInput, err := ff.ProbeAudio(fileAudioInput)
	if err != nil {
		return "", err
	}

	if aiSettings["asrProvider"] == "google" {
		fileAudioInputGoogle := fileAudioInput
		if db.Number(probeInput["duration"]) > maxAudioInputTime {
			return i18n.Translate(language, "audio_input_limit"), nil
		}
		if probeInput["codec_name"] != "opus" {
			fileAudioInputGoogle = fileAudioInput + ".google"
			if err := ff.MpegAudioOpus(fileAudioInput, fileAudioInputGoogle); err != nil {
				return "", err
			}
		}
		transcript, err = google.ASR(fileAudioInputGoogle, i18n.LanguageCodeToLocale(language))
		if err != nil {
			return "", err
		}
	}

	if aiSettings["asrProvider"] == "openai" {
		fileAudioInputWhisper := fileAudioInput
		if probeInput["codec_name"] != "pcm_s16le" || probeInput["sample_rate"] != "16000" || probeInput["channels"] != "1" {
			fileAudioInputWhisper = fileAudioInput + ".whisper"
			if err := ff.MpegAudioWhisper(fileAudioInput, fileAudioInputWhisper); err != nil {
				return "", err
			}
		}
		transcript, err = openai.ASR(fileAudioInputWhisper, language)
		if err != nil {
			return "", err
		}
	}

	return transcript, nil
}
