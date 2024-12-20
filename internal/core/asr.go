// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"github.com/eja/tibula/log"
	"pbx/internal/av"
	"pbx/internal/db"
	"pbx/internal/google"
	"pbx/internal/i18n"
	"pbx/internal/openai"
	"pbx/internal/sys"
)

func ASR(fileAudioInput, language string) (string, error) {
	aiSettings := db.Settings()
	transcript := ""
	probeInput, err := av.ProbeAudio(fileAudioInput)
	if err != nil {
		return "", err
	}

	if aiSettings["asrProvider"] == "google" {
		fileAudioInputGoogle := fileAudioInput
		if sys.Number(probeInput["duration"]) > maxAudioInputTime {
			return i18n.Translate(language, "audio_input_limit"), nil
		}
		if probeInput["codec_name"] != "opus" {
			fileAudioInputGoogle = fileAudioInput + ".google"
			if err := av.MpegAudioOpus(fileAudioInput, fileAudioInputGoogle); err != nil {
				return "", err
			}
		}
		transcript, err = google.ASR(fileAudioInputGoogle, i18n.LanguageCodeToLocale(language))
		if err != nil {
			return "", err
		}

	} else {

		fileAudioInputWhisper := fileAudioInput
		if probeInput["codec_name"] != "pcm_s16le" || probeInput["sample_rate"] != "16000" || probeInput["channels"] != "1" {
			fileAudioInputWhisper = fileAudioInput + ".whisper"
			if err := av.MpegAudioWhisper(fileAudioInput, fileAudioInputWhisper); err != nil {
				return "", err
			}
		}
		transcript, err = openai.ASR(fileAudioInputWhisper, language)
		if err != nil {
			return "", err
		}
	}

	log.Debug(tag, "asr", language, transcript)
	return transcript, nil
}
