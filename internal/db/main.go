// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"

	"github.com/eja/tibula/db"
	"pbx/internal/sys"
)

func Number(value interface{}) int64 {
	return db.Number(value)
}

func UserGet(id string) (db.TypeRow, error) {
	return db.Row("SELECT * FROM aiUsers WHERE id = ? AND expiration > CURRENT_TIMESTAMP LIMIT 1", id)
}

func UserUpdate(id string, field string, value string) (err error) {
	query := fmt.Sprintf("UPDATE aiUsers SET %s = ? WHERE id = ?", field)
	_, err = db.Run(query, value, id)
	return
}

func SystemPrompt(platform string) (db.TypeRows, error) {
	return db.Rows("SELECT prompt FROM aiPrompts WHERE active > 0 AND (platform='' OR platform='all' OR platform=?) ORDER BY power ASC", platform)
}

func ChatAction(platform, action, language string) (function, response string) {
	values, err := db.Row(`SELECT function, response FROM aiActions WHERE 
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

func Open() error {
	if err := db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort); err != nil {
		return err
	}
	return nil
}

func Settings() (values db.TypeRow) {
	values, _ = db.Get(1, db.ModuleGetIdByName("aiSettings"), 1)
	return
}

func Log(source, message string) error {
	settings := Settings()
	if Number(settings["logs"]) > 0 {
		if _, err := db.Run(`INSERT INTO aiLogs (ejaOwner, ejaLog, source, message) VALUES (1,?,?,?)`, db.Now(), source, message); err != nil {
			return err
		}
	}
	return nil
}
