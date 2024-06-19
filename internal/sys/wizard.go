// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"embed"
	"strconv"
	"strings"

	"github.com/eja/tibula/db"
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
	Options.GoogleToken = sys.WizardPrompt("Google api key")
	Options.TelegramToken = sys.WizardPrompt("Telegram token")
	Options.MetaUrl = sys.WizardPrompt("Meta graph api url")
	if Options.MetaUrl != "" {
		Options.MetaUser = sys.WizardPrompt("Meta user id")
		Options.MetaAuth = sys.WizardPrompt("Meta auth")
		Options.MetaToken = sys.WizardPrompt("Meta token")
	}
	Options.OpenaiToken = sys.WizardPrompt("OpenAI LLM key")
	if Options.OpenaiToken != "" {
		Options.OpenaiUrl = sys.WizardPrompt("OpenAI LLM url")
		Options.OpenaiModel = sys.WizardPrompt("OpenAI LLM model")
	}
	asterisk := sys.WizardPrompt("Enable Asterisk AGI server? (N/y)")
	if len(asterisk) > 0 && strings.ToLower(asterisk[0:1]) == "y" {
		Options.Asterisk = true
		Options.AsteriskHost = sys.WizardPrompt("Asterisk host (localhost)")
		asteriskPort := sys.WizardPrompt("Asterisk port (4573)")
		if asteriskPort != "" {
			Options.AsteriskPort, _ = strconv.Atoi(asteriskPort)
		}
		Options.AsteriskToken = sys.WizardPrompt("Asterisk token")
	}
	Options.MailSender = sys.WizardPrompt("Mail Sender")

	db.Assets = dbAssets
	if err := db.Open(Options.DbType, Options.DbName, Options.DbUser, Options.DbPass, Options.DbHost, Options.DbPort); err != nil {
		return err
	}
	if err := db.Setup(""); err != nil {
		return err
	}

	Options.ConfigFile = ""
	return sys.ConfigWrite(configFile, &Options)
}