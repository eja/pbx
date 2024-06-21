// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package asterisk

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"pbx/internal/db"
	"pbx/internal/sys"
)

type AriField struct {
	Attribute string `json:"attribute"`
	Value     string `json:"value"`
}

type AriPayload struct {
	Fields []AriField `json:"fields"`
}

func ari(method, token, origin, class, id string, data AriPayload) (ariResponse []AriField, err error) {
	url := fmt.Sprintf("%s/ari/asterisk/config/dynamic/res_pjsip/%s/%s", origin, class, id)
	method = strings.ToUpper(method)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")

	base64Auth := base64.StdEncoding.EncodeToString([]byte(token))
	req.Header.Set("Authorization", "Basic "+base64Auth)

	client := &http.Client{}

	resp, err := client.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request failed with status code: %d\n", resp.StatusCode)
	}
	return
}

func SipUpdate(address, username, password string) (err error) {
	aiSettings := db.Settings()
	token := aiSettings["asteriskToken"]
	if token == "" {
		token = sys.Options.AsteriskToken
	}
	origin := aiSettings["asteriskAri"]
	if origin == "" {
		origin = sys.Options.AsteriskAri
	}
	id := fmt.Sprintf("%x", md5.Sum([]byte(address+username)))

	auth := AriPayload{
		Fields: []AriField{
			{Attribute: "username", Value: username},
			{Attribute: "password", Value: password},
		},
	}
	if _, err = ari("put", token, origin, "auth", id, auth); err != nil {
		return
	}

	aor := AriPayload{
		Fields: []AriField{
			{Attribute: "remove_existing", Value: "yes"},
			{Attribute: "contact", Value: fmt.Sprintf("sip:%s@%s", username, address)},
			{Attribute: "max_contacts", Value: "1"},
		},
	}
	if _, err = ari("put", token, origin, "aor", id, aor); err != nil {
		return
	}

	endpoint := AriPayload{
		Fields: []AriField{
			{Attribute: "context", Value: "agi"},
			{Attribute: "aors", Value: id},
			{Attribute: "auth", Value: id},
			{Attribute: "outbound_auth", Value: id},
			{Attribute: "from_user", Value: username},
			{Attribute: "from_domain", Value: address},
			{Attribute: "allow", Value: "!all,ulaw,alaw"},
		},
	}
	if _, err = ari("put", token, origin, "endpoint", id, endpoint); err != nil {
		return
	}

	/*
		registration := AriPayload{
			Fields: []AriField{
				{Attribute: "outbound_auth", Value: id},
				{Attribute: "endpoint", Value: id},
				{Attribute: "line", Value: "yes"},
				{Attribute: "contact_user", Value: username},
				{Attribute: "server_uri", Value: fmt.Sprintf("sip:%s", address)},
				{Attribute: "client_uri", Value: fmt.Sprintf("sip:%s@%s", username, address)},
			},
		}
		if _, err = ari("put", token, origin, "registration", id, registration); err != nil {
			return
		}
	*/
	return
}
