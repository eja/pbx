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
	flag.StringVar(&Options.MetaUrl, "meta-url", "", "Meta graph api url")
	flag.StringVar(&Options.MetaUser, "meta-user", "", "Meta user id")
	flag.StringVar(&Options.MetaAuth, "meta-auth", "", "Meta auth")
	flag.StringVar(&Options.MetaToken, "meta-token", "", "Meta token")
	flag.StringVar(&Options.TelegramToken, "telegram-token", "", "Telegram token")
	flag.StringVar(&Options.AiToken, "ai-token", "", "AI token")
	flag.StringVar(&Options.AiProvider, "ai-provider", "openai", "AI provider [openai|google|anythingLLM]")
	flag.StringVar(&Options.AsteriskAgi, "asterisk-agi", "127.0.0.1:4573", "Asterisk AGI address")
	flag.StringVar(&Options.AsteriskAri, "asterisk-ari", "http://localhost:8088", "Asterisk ARI url")
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
