// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package asterisk

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

	if resp.StatusCode == http.StatusOK {
		fmt.Println("Request was successful")
	} else {
		fmt.Printf("Request failed with status code: %d\n", resp.StatusCode)
	}
	return
}
