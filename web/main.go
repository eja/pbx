// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"

	"github.com/eja/pbx/asterisk"
	"github.com/eja/pbx/pbx"
	"github.com/eja/pbx/sys"
	"github.com/eja/tibula/api"
	"github.com/eja/tibula/web"
)

//go:embed assets/*
var embeddedFiles embed.FS

func Router() error {
	assetFS, err := fs.Sub(embeddedFiles, "assets")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(assetFS))
	web.Router.Handle("/pbx/", http.StripPrefix("/pbx/", fileServer))

	if sys.Options.Chat || sys.Options.ChatAudio {
		web.Router.HandleFunc("/chat", chatRouter)
	}
	web.Router.HandleFunc("/meta", metaRouter)
	web.Router.HandleFunc("/tg", telegramRouter)

	api.Plugins["aiChat"] = func(eja api.Api, db api.DbSession) api.Api {
		if eja.Action == "run" && eja.Values["chat"] != "" {
			user := fmt.Sprintf("T.%d", eja.Owner)
			language := eja.Language
			if answer, err := pbx.Text(user, language, eja.Values["chat"]); err != nil {
				eja.Alert = append(eja.Alert, err.Error())
			} else {
				eja.Values["chat"] = answer
			}
		}
		return eja
	}

	api.Plugins["aiSettings"] = func(eja api.Api, db api.DbSession) api.Api {
		data, _ := db.Get(eja.Owner, eja.ModuleId, 1)
		if eja.Action == "run" {
			if data == nil {
				db.New(eja.Owner, eja.ModuleId)
			}
			for k, v := range eja.Values {
				db.Put(eja.Owner, eja.ModuleId, 1, k, v)
			}
		} else {
			eja.Values = data
		}
		eja.Action = "edit"
		eja.ActionType = "Edit"

		return eja
	}

	api.Plugins["aiSip"] = func(eja api.Api, db api.DbSession) api.Api {
		if eja.Action == "delete" {
			if eja.Values["username"] != "" {
				if err := asterisk.SipDelete(eja.Values["username"]); err != nil {
					eja.Alert = append(eja.Alert, "Sync error, please check asterisk")
				} else {
					eja.Info = append(eja.Info, "SIP account removed from Asterisk")
				}
			} else {
				eja.Alert = append(eja.Alert, "Empty values, didn't sync")
			}

		}
		if eja.Action == "save" {
			if eja.Values["address"] != "" && eja.Values["username"] != "" && eja.Values["password"] != "" {
				if err := asterisk.SipUpdate(eja.Values["address"], eja.Values["username"], eja.Values["password"], eja.Values["trunk"], eja.Values["webrtc"]); err != nil {
					eja.Alert = append(eja.Alert, fmt.Sprintf("Sync error: %v", err))
				} else {
					eja.Info = append(eja.Info, "SIP account synced")
				}
			} else {
				eja.Alert = append(eja.Alert, "Missing fields, not syncing")
			}
		}

		return eja
	}

	return nil
}
