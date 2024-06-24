// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package asterisk

import (
	"bytes"
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

func SipUpdate(address, username, password, trunk, webrtc string) (err error) {
	aiSettings := db.Settings()
	token := aiSettings["asteriskToken"]
	if token == "" {
		token = sys.Options.AsteriskToken
	}
	origin := aiSettings["asteriskAri"]
	if origin == "" {
		origin = sys.Options.AsteriskAri
	}

	auth := AriPayload{
		Fields: []AriField{
			{Attribute: "username", Value: username},
			{Attribute: "password", Value: password},
		},
	}
	if _, err = ari("put", token, origin, "auth", username, auth); err != nil {
		return
	}

	aor := AriPayload{
		Fields: []AriField{
			{Attribute: "remove_existing", Value: "yes"},
			{Attribute: "contact", Value: fmt.Sprintf("sip:%s@%s", username, address)},
			{Attribute: "max_contacts", Value: "1"},
		},
	}
	if _, err = ari("put", token, origin, "aor", username, aor); err != nil {
		return
	}

	endpoint := AriPayload{
		Fields: []AriField{
			{Attribute: "context", Value: "agi"},
			{Attribute: "aors", Value: username},
			{Attribute: "from_user", Value: username},
			{Attribute: "from_domain", Value: address},
			{Attribute: "allow", Value: "!all,ulaw,alaw"},
		},
	}
	if db.Number(webrtc) > 0 {
		endpoint.Fields = append(endpoint.Fields, AriField{Attribute: "webrtc", Value: "yes"})
	}
	if db.Number(trunk) > 0 {
		endpoint.Fields = append(endpoint.Fields, AriField{Attribute: "auth", Value: ""})
		endpoint.Fields = append(endpoint.Fields, AriField{Attribute: "outbound_auth", Value: username})
	} else {
		endpoint.Fields = append(endpoint.Fields, AriField{Attribute: "auth", Value: username})
		endpoint.Fields = append(endpoint.Fields, AriField{Attribute: "outbound_auth", Value: ""})
	}
	if _, err = ari("put", token, origin, "endpoint", username, endpoint); err != nil {
		return
	}

	/*
		if db.Number(trunk) > 0 {
			registration := AriPayload{
				Fields: []AriField{
					{Attribute: "outbound_auth", Value: username},
					{Attribute: "endpoint", Value: username},
					{Attribute: "line", Value: "yes"},
					{Attribute: "contact_user", Value: username},
					{Attribute: "server_uri", Value: fmt.Sprintf("sip:%s", address)},
					{Attribute: "client_uri", Value: fmt.Sprintf("sip:%s@%s", username, address)},
				},
			}
			if _, err = ari("put", token, origin, "registration", username, registration); err != nil {
				return
			}
		}
	*/
	return
}
