// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/smtp"
	"os"
	"path/filepath"
)

func Mail(from, to, subject, body, attachmentPath string) error {
	client, err := smtp.Dial("localhost:25")
	if err != nil {
		return err
	}
	defer client.Close()

	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}

	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	headers := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-version: 1.0;\r\nContent-Type: multipart/mixed; boundary=boundary\r\n\r\n--boundary\r\nContent-Type: text/plain; charset=utf-8\r\n\r\n%s\r\n\r\n", from, to, subject, body)
	if _, err := fmt.Fprint(w, headers); err != nil {
		return err
	}

	if attachmentPath != "" {
		file, err := os.Open(attachmentPath)
		if err != nil {
			return err
		}
		defer file.Close()

		fmt.Fprintf(w, "--boundary\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\nContent-Disposition: attachment; filename=\"%s\"\r\n\r\n", filepath.Base(attachmentPath))

		enc := base64.NewEncoder(base64.StdEncoding, w)
		if _, err := io.Copy(enc, file); err != nil {
			return err
		}
		if err := enc.Close(); err != nil {
			return err
		}

		fmt.Fprint(w, "\r\n\r\n")
	}

	_, err = fmt.Fprint(w, "--boundary--\r\n")
	return err
}
