// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package pbx

import (
	"log/slog"

	"github.com/eja/pbx/db"
	"github.com/eja/pbx/google"
	"github.com/eja/pbx/i18n"
	"github.com/eja/pbx/media"
	"github.com/eja/pbx/openai"
	"github.com/eja/pbx/sys"
)

func ASR(fileAudioInput, language string) (string, error) {
	aiSettings := db.Settings()
	transcript := ""
	probeInput, err := media.ProbeAudio(fileAudioInput)
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
			if err := media.MpegAudioOpus(fileAudioInput, fileAudioInputGoogle); err != nil {
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
			if err := media.MpegAudioWhisper(fileAudioInput, fileAudioInputWhisper); err != nil {
				return "", err
			}
		}
		transcript, err = openai.ASR(fileAudioInputWhisper, language)
		if err != nil {
			return "", err
		}
	}

	slog.Debug("ASR", "language", language, "transcript", transcript)
	return transcript, nil
}
