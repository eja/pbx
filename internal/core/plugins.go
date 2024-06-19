// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"fmt"
	"strings"

	"github.com/eja/tibula/log"
	"pbx/internal/sys"
)

type TypeAiChatPlugins map[string]func(userId, language, action, output string) string

var AiChatPlugins = TypeAiChatPlugins{
	"reset": func(userId, language, action, output string) string {
		delete(history, userId)
		return output
	},
	"mail": func(userId, language, action, output string) string {
		parameters := strings.Split(action, " ")
		subject := fmt.Sprintf("Call History %s", userId)
		body := ""
		file := ""

		if len(parameters) > 1 && parameters[1] != "" {
			file = parameters[1]
		}

		var sb strings.Builder
		for _, msg := range history[userId] {
			content, _ := TagsExtract(msg.Content)
			if msg.Role == "user" {
				sb.WriteString(fmt.Sprintf("\nQ: %s\n", content))
			}
			if msg.Role == "assistant" {
				sb.WriteString(fmt.Sprintf("A: %s\n", content))
			}
		}
		body = sb.String()
		if output != "" {
			if err := sys.Mail(sys.Options.MailSender, output, subject, body, file); err != nil {
				log.Warn("mail error", err)
			}
		}
		return output
	},
	"ntfy": func(userId, language, action, output string) string {
		parameters := strings.Split(action, " ")
		subject := fmt.Sprintf("%s", userId)
		body := ""
		file := ""

		if len(parameters) > 1 && parameters[1] != "" {
			file = parameters[1]
		}

		var sb strings.Builder
		for _, msg := range history[userId] {
			content, _ := TagsExtract(msg.Content)
			if msg.Role == "user" {
				sb.WriteString(fmt.Sprintf("\nQ: %s\n", content))
			}
			if msg.Role == "assistant" {
				sb.WriteString(fmt.Sprintf("A: %s\n", content))
			}
		}
		body = sb.String()
		if output != "" {
			if err := sys.Ntfy(output, subject, body, file); err != nil {
				log.Warn("mail error", err)
			}
		}
		return output
	},
}
