// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package i18n

import (
	"github.com/eja/pbx/db"
	"github.com/eja/pbx/sys"
)

func DefaultLanguage() string {
	language, err := db.DefaultLanguage()
	if err != nil || language == "" {
		language = sys.Options.Language
	}
	return language
}

func Translate(language string, label string) string {
	if language == "" {
		language = DefaultLanguage()
	}
	value, err := db.Translate(label, language)
	if err != nil || value == "" {
		return "{" + label + "}"
	} else {
		return value
	}
}

func LanguageCodeToLocale(language string) string {
	if language == "" {
		language = DefaultLanguage()
	}

	if locale, err := db.LanguageCodeToLocale(language); err != nil {
		return ""
	} else {
		return locale
	}
}

func LanguageCodeToInternal(language string) string {
	if language == "" {
		language = DefaultLanguage()
	}

	if value, err := db.LanguageCodeToInternal(language); err != nil {
		return ""
	} else {
		return value
	}
}
