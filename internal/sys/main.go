// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"flag"
	"os"

	"github.com/eja/tibula/sys"
)

var Options typeConfigPbx

func Configure() error {
	flag.StringVar(&Options.MediaPath, "media-path", "/tmp/", "Media temporary folder")
	flag.StringVar(&Options.GoogleToken, "google-token", "", "Google API key")
	flag.StringVar(&Options.GoogleModel, "google-model", "gemini-1.5-flash", "Google LLM model")
	flag.StringVar(&Options.MetaUrl, "meta-url", "", "Meta graph api url")
	flag.StringVar(&Options.MetaUser, "meta-user", "", "Meta user id")
	flag.StringVar(&Options.MetaAuth, "meta-auth", "", "Meta auth")
	flag.StringVar(&Options.MetaToken, "meta-token", "", "Meta token")
	flag.StringVar(&Options.TelegramToken, "telegram-token", "", "Telegram token")
	flag.StringVar(&Options.OpenaiToken, "openai-token", "", "OpenAI LLM token")
	flag.StringVar(&Options.OpenaiUrl, "openai-url", "https://api.openai.com/v1/chat/completions", "OpenAI LLM url")
	flag.StringVar(&Options.OpenaiModel, "openai-model", "gpt-3.5-turbo", "OpenAI LLM model")
	flag.StringVar(&Options.AsteriskHost, "asterisk-host", "127.0.0.1", "Asterisk host")
	flag.IntVar(&Options.AsteriskPort, "asterisk-port", 4573, "Asterisk port")
	flag.StringVar(&Options.AsteriskToken, "asterisk-token", "", "Asterisk token")
	flag.StringVar(&Options.Cache, "cache", "/tmp/", "Cache path")
	flag.StringVar(&Options.MailSender, "mail-sender", "", "Mail sender")
	flag.BoolVar(&Options.Asterisk, "asterisk", false, "start the asterisk agi service")

	if err := sys.Configure(); err != nil {
		return err
	}
	Options.TypeConfig = sys.Options

	if sys.Commands.Start && sys.Options.ConfigFile == "" {
		sys.Options.ConfigFile = Name + ".json"
		if _, err := os.Stat(sys.Options.ConfigFile); err != nil {
			sys.Options.ConfigFile = ""
		}
	}

	if sys.Options.ConfigFile != "" {
		if err := sys.ConfigRead(sys.Options.ConfigFile, &Options); err != nil {
			return err
		}
	}

	return nil
}
