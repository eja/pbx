// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

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
	AiToken       string `json:"ai_token,omitempty"`
	LlmUrl        string `json:"llm_url,omitempty"`
	Asterisk      bool   `json:"asterisk,omitempty"`
	AsteriskAgi   string `json:"asterisk_agi,omitempty"`
	AsteriskAri   string `json:"asterisk_ari,omitempty"`
	AsteriskToken string `json:"asterisk_token,omitempty"`
	Cache         string `json:"cache,omitempty"`
	MailSender    string `json:"mail_sender,omitempty"`
	McpUrl        string `json:"mcp_url,omitempty"`
	McpToken      string `json:"mcp_token,omitempty"`
	Chat          bool   `json:"chat,omitempty"`
	ChatAudio     bool   `json:"chat_audio,omitempty"`
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
