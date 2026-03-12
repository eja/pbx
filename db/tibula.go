// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package db

import (
	"github.com/eja/pbx/sys"

	tibula "github.com/eja/tibula/db"
)

type tibulaTypeRun = tibula.TypeRun
type tibulaTypeRow = tibula.TypeRow
type tibulaTypeRows = tibula.TypeRows
type tibulaTypeField = tibula.TypeField
type tibulaTypeCommand = tibula.TypeCommand

var tibulaSession = new(tibula.TypeSession)
var (
	tibulaNow      = tibulaSession.Now
	tibulaPassword = tibulaSession.Password
	tibulaString   = tibulaSession.String
	tibulaNumber   = tibulaSession.Number
	tibulaFloat    = tibulaSession.Float
	tibulaBool     = tibulaSession.Bool
)

func tibulaConnection() (tibula.TypeSession, error) {
	var db tibula.TypeSession
	err := db.Open(sys.Options.DbType, sys.Options.DbName, sys.Options.DbUser, sys.Options.DbPass, sys.Options.DbHost, sys.Options.DbPort)
	return db, err
}

func tibulaRun(query string, args ...any) (result tibula.TypeRun, err error) {
	db, err := tibulaConnection()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Run(query, args...)
}

func tibulaValue(query string, args ...any) (value string, err error) {
	db, err := tibulaConnection()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Value(query, args...)
}

func tibulaRow(query string, args ...any) (row tibula.TypeRow, err error) {
	db, err := tibulaConnection()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Row(query, args...)
}

func tibulaRows(query string, args ...any) (rows tibula.TypeRows, err error) {
	db, err := tibulaConnection()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Rows(query, args...)
}

func tibulaCols(query string, args ...any) (cols []string, err error) {
	db, err := tibulaConnection()
	if err != nil {
		return
	}
	defer db.Close()

	return db.Cols(query, args...)
}
