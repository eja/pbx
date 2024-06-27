// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"github.com/eja/tibula/sys"
)

type typeConfigPbx struct {
	sys.TypeConfig
	MediaPath     string `json:"media_path,omitempty"`
	MetaUrl       string `json:"meta_url,omitempty"`
	MetaUser      string `json:"meta_user,omitempty"`
	MetaAuth      string `json:"meta_auth,omitempty"`
	MetaToken     string `json:"meta_token,omitempty"`
	TelegramToken string `json:"telegram_token,omitempty"`
	OpenaiToken   string `json:"openai_token,omitempty"`
	OpenaiModel   string `json:"openai_model,omitempty"`
	Asterisk      bool   `json:"asterisk,omitempty"`
	AsteriskAgi   string `json:"asterisk_agi,omitempty"`
	AsteriskAri   string `json:"asterisk_ari,omitempty"`
	AsteriskToken string `json:"asterisk_token,omitempty"`
	GoogleToken   string `json:"google_token,omitempty"`
	GoogleModel   string `json:"google_model,omitempty"`
	Cache         string `json:"cache,omitempty"`
	MailSender    string `json:"mail_sender,omitempty"`
}

type TypeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
