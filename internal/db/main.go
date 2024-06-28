// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"fmt"

	"pbx/internal/sys"

	tibula "github.com/eja/tibula/db"
)

var sharedSession tibula.TypeSession

func Open() (tibula.TypeSession, error) {
	if sharedSession == (tibula.TypeSession{}) {
		sharedSession = tibula.Session()
	}
	err := sharedSession.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort)
	return sharedSession, err
}

func UserGet(id string) (tibula.TypeRow, error) {
	db, _ := Open()
	return db.Row("SELECT * FROM aiUsers WHERE id = ? AND expiration > CURRENT_TIMESTAMP LIMIT 1", id)
}

func UserUpdate(id string, field string, value string) (err error) {
	db, _ := Open()
	query := fmt.Sprintf("UPDATE aiUsers SET %s = ? WHERE id = ?", field)
	_, err = db.Run(query, value, id)
	return
}

func SystemPrompt(platform string) (tibula.TypeRows, error) {
	db, _ := Open()
	return db.Rows("SELECT prompt FROM aiPrompts WHERE active > 0 AND (platform='' OR platform='all' OR platform=?) ORDER BY power ASC", platform)
}

func ChatAction(platform, action, language string) (function, response string) {
	db, _ := Open()
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

func Settings() (values tibula.TypeRow) {
	db, _ := Open()
	values, _ = db.Get(1, db.ModuleGetIdByName("aiSettings"), 1)
	return
}

func Log(source, message string) error {
	db, _ := Open()
	settings := Settings()
	if sys.Number(settings["logs"]) > 0 {
		if _, err := db.Run(`INSERT INTO aiLogs (ejaOwner, ejaLog, source, message) VALUES (1,?,?,?)`, db.Now(), source, message); err != nil {
			return err
		}
	}
	return nil
}
