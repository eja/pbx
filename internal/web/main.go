// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"fmt"
	"strings"

	"github.com/eja/tibula/api"
	"github.com/eja/tibula/db"
	"github.com/eja/tibula/web"
	"pbx/internal/asterisk"
	"pbx/internal/core"
	"pbx/internal/sys"
)

func Router() error {
	web.Router.HandleFunc("/meta", metaRouter)
	web.Router.HandleFunc("/tg", telegramRouter)

	api.Plugins["aiChat"] = func(eja api.TypeApi) api.TypeApi {
		if eja.Action == "run" && eja.Values["chat"] != "" {
			user := fmt.Sprintf("T.%d", eja.Owner)
			language := eja.Language
			if answer, err := core.Text(user, language, eja.Values["chat"]); err != nil {
				eja.Alert = append(eja.Alert, err.Error())
			} else {
				eja.Values["chat"] = answer
			}
		}
		return eja
	}

	api.Plugins["aiSettings"] = func(eja api.TypeApi) api.TypeApi {
		data, _ := db.Get(eja.Owner, eja.ModuleId, 1)
		if eja.Action == "run" {
			if data == nil {
				db.New(eja.Owner, eja.ModuleId)
			}
			for k, v := range eja.Values {
				add := true
				if strings.Contains("#llmProvider#ttsProvider#asrProvider#", "#"+k+"#") {
					if v == "google" && sys.Options.GoogleToken == "" && eja.Values["googleToken"] == "" {
						add = false
						eja.Values[k] = ""
					}
					if v == "openai" && sys.Options.OpenaiToken == "" && eja.Values["openaiToken"] == "" {
						add = false
						eja.Values[k] = ""
					}
				}
				if add {
					db.Put(eja.Owner, eja.ModuleId, 1, k, v)
				} else {
					eja.Alert = append(eja.Alert, "Missing provider credentials, cannot enable")
				}
			}
		} else {
			eja.Values = data
		}

		return eja
	}

	api.Plugins["aiSip"] = func(eja api.TypeApi) api.TypeApi {
		if eja.Action == "save" {
			if eja.Values["address"] != "" && eja.Values["username"] != "" && eja.Values["password"] != "" {
				if err := asterisk.SipUpdate(eja.Values["address"], eja.Values["username"], eja.Values["password"]); err != nil {
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
