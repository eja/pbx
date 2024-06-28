// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package i18n

import (
	"pbx/internal/db"
	"pbx/internal/sys"
)

func DefaultLanguage() string {
	db, _ := db.Open()
	language, err := db.Value("SELECT language FROM aiSettings WHERE ejaId=1")
	if err != nil || language == "" {
		language = sys.Options.Language
	}
	return language
}

func Translate(language string, label string) string {
	db, _ := db.Open()
	if language == "" {
		language = DefaultLanguage()
	}
	value, err := db.Value("SELECT translation FROM aiTranslations WHERE label=? AND language=?", label, language)
	if err != nil || value == "" {
		return "{" + label + "}"
	} else {
		return value
	}
}

func LanguageCodeToLocale(language string) string {
	db, _ := db.Open()
	if language == "" {
		language = DefaultLanguage()
	}

	if locale, err := db.Value("SELECT locale FROM aiLanguages WHERE code = ?", language); err != nil {
		return ""
	} else {
		return locale
	}
}

func LanguageCodeToInternal(language string) string {
	db, _ := db.Open()
	if language == "" {
		language = DefaultLanguage()
	}

	if value, err := db.Value("SELECT internal FROM aiLanguages WHERE code = ?", language); err != nil {
		return ""
	} else {
		return value
	}
}
