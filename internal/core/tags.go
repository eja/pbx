// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package core

import (
	"regexp"
	"strings"

	"github.com/eja/tibula/log"
	"pbx/internal/sys"
)

func TagsProcess(platform, language, userId, message string, tags []string) (response string, err error) {
	log.Trace(tag, "processing", tags, message)
	response = message
	for _, item := range tags {
		lower := strings.ToLower(item)
		if strings.HasPrefix(lower, "ntfy:") {
			if err = sys.Ntfy(item[5:], userId, message, ""); err != nil {
				log.Warn(tag, "ntfy sending problem", err)
				if response, err = Chat(platform, userId, "/ntfy_error", language); err != nil {
					return
				}
			} else {
				if response, err = Chat(platform, userId, "/ntfy_sent", language); err != nil {
					return
				}
			}
		}
	}
	return
}

func TagsExtract(text string) (string, []string) {
	re := regexp.MustCompile(`\s*\[([^\]]+)\]\s*$`)
	var tags []string

	for {
		matches := re.FindStringSubmatchIndex(text)
		if len(matches) == 0 {
			break
		}
		tag := text[matches[2]:matches[3]]
		tags = append(tags, tag)
		text = strings.TrimSuffix(text[:matches[0]], " ["+tag)
	}

	return text, tags
}

func FilterLanguage(tags []string, language string) string {
	re := regexp.MustCompile(`\[\w\w\]`)
	for _, tag := range tags {
		if re.MatchString(tag) {
			language = tag[1:3]
		}
	}
	return language
}
