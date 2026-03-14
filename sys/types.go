// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"github.com/eja/tibula/sys"
)

type typeConfigPbx struct {
	sys.TypeConfig
	MediaPath     string `json:"media_path"`
	MetaUrl       string `json:"meta_url"`
	MetaUser      string `json:"meta_user"`
	MetaAuth      string `json:"meta_auth"`
	MetaToken     string `json:"meta_token"`
	TelegramToken string `json:"telegram_token"`
	AiToken       string `json:"ai_token"`
	AiProvider    string `json:"ai_provider"`
	Asterisk      bool   `json:"asterisk"`
	AsteriskAgi   string `json:"asterisk_agi"`
	AsteriskAri   string `json:"asterisk_ari"`
	AsteriskToken string `json:"asterisk_token"`
	Cache         string `json:"cache"`
	MailSender    string `json:"mail_sender"`
	McpUrl        string `json:"mcp_url"`
	McpToken      string `json:"mcp_token"`
}

type TypeChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`

	Name       string `json:"name,omitempty"`
	ToolCallID string `json:"tool_call_id,omitempty"`
	ToolCalls  any    `json:"tool_calls,omitempty"`
}

var String = sys.String
var Number = sys.Number
var Float = sys.Float
var Bool = sys.Bool
