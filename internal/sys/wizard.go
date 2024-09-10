// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"embed"
	"strings"

	tibulaDb "github.com/eja/tibula/db"
	"github.com/eja/tibula/sys"
)

//go:embed all:assets
var dbAssets embed.FS

func Wizard() error {
	configFile := sys.Options.ConfigFile
	if err := sys.ConfigRead(configFile, &Options); err != nil {
		return err
	}

	Options.MediaPath = sys.WizardPrompt("Media temporary folder path")
	Options.Cache = sys.WizardPrompt("Cache folder path")
	Options.TelegramToken = sys.WizardPrompt("Telegram token")
	Options.MetaUrl = sys.WizardPrompt("Meta graph api url")
	if Options.MetaUrl != "" {
		Options.MetaUser = sys.WizardPrompt("Meta user id")
		Options.MetaAuth = sys.WizardPrompt("Meta auth")
		Options.MetaToken = sys.WizardPrompt("Meta token")
	}
	Options.AiToken = sys.WizardPrompt("AI API key")
	Options.AiProvider = sys.WizardPrompt("AI Provider [openai|google|assistant]")
	asterisk := sys.WizardPrompt("Enable Asterisk AGI server? (N/y)")
	if len(asterisk) > 0 && strings.ToLower(asterisk[0:1]) == "y" {
		Options.Asterisk = true
		Options.AsteriskAgi = sys.WizardPrompt("Asterisk AGI address (localhost:4573)")
		Options.AsteriskToken = sys.WizardPrompt("Asterisk token")
		Options.AsteriskAri = sys.WizardPrompt("Asterisk ARI url")
	}
	Options.MailSender = sys.WizardPrompt("Mail Sender")

	tibulaDb.Assets = dbAssets

	db := tibulaDb.Session()
	if err := db.Open(Options.DbType, Options.DbName, Options.DbUser, Options.DbPass, Options.DbHost, Options.DbPort); err != nil {
		return err
	}
	if err := db.Setup(""); err != nil {
		return err
	}

	Options.ConfigFile = ""
	return sys.ConfigWrite(configFile, &Options)
}
