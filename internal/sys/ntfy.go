// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"fmt"
	"net/http"
	"os"
	"strings"
)

func Ntfy(link, title, message, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	req, err := http.NewRequest("PUT", link, file)
	if err != nil {
		return err
	}
	req.Header.Set("X-Title", strings.Trim(fmt.Sprintf("%q", title), `"`))
	req.Header.Set("X-Message", strings.Trim(fmt.Sprintf("%q", message), `"`))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err
	}

	return nil
}
