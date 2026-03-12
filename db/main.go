// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"
)

func UserGet(id string) (tibulaTypeRow, error) {
	return tibulaRow("SELECT * FROM aiUsers WHERE id = ? AND expiration > CURRENT_TIMESTAMP LIMIT 1", id)
}

func UserUpdate(id string, field string, value string) (err error) {
	query := fmt.Sprintf("UPDATE aiUsers SET %s = ? WHERE id = ?", field)
	_, err = tibulaRun(query, value, id)
	return
}

func SystemPrompt(platform string) (tibulaTypeRows, error) {
	return tibulaRows("SELECT prompt FROM aiPrompts WHERE active > 0 AND (platform='' OR platform='all' OR platform=?) ORDER BY power ASC", platform)
}

func ChatAction(platform, action, language string) (function, response string) {
	values, err := tibulaRow(`SELECT function, response FROM aiActions WHERE 
			active>0 AND 
			action=? AND 
			(language=? OR language='') AND 
			(platform='' OR platform='all' OR platform=?) 
			ORDER by language DESC LIMIT 1`,
		action, language, platform)
	if err == nil {
		function = values["function"]
		response = values["response"]
	}
	return
}

func Settings() (values tibulaTypeRow) {
	values, _ = tibulaRow(`SELECT * FROM aiSettings WHERE ejaId=1`)
	return
}

func Log(source, message string) error {
	settings := Settings()
	if tibulaNumber(settings["logs"]) > 0 {
		if _, err := tibulaRun(`INSERT INTO aiLogs (ejaOwner, ejaLog, source, message) VALUES (1,?,?,?)`, tibulaNow(), source, message); err != nil {
			return err
		}
	}
	return nil
}

func DefaultLanguage() (string, error) {
	return tibulaValue("SELECT language FROM aiSettings WHERE ejaId=1")
}

func Translate(label, language string) (string, error) {
	return tibulaValue("SELECT translation FROM aiTranslations WHERE label=? AND language=? LIMIT 1", label, language)
}

func LanguageCodeToLocale(language string) (string, error) {
	return tibulaValue("SELECT locale FROM aiLanguages WHERE code = ? LIMIT 1", language)
}

func LanguageCodeToInternal(language string) (string, error) {
	return tibulaValue("SELECT internal FROM aiLanguages WHERE code = ? LIMIT 1", language)
}
