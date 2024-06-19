// Copyright (C) 2023-2024 by Ubaldo Porcheddu <ubaldo@eja.it>

package sys

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
)

// Mail sends an email using SMTP without authentication
func Mail(from, to, subject, messageBody string, attachmentPath string) error {
	// Connect to the SMTP server
	client, err := smtp.Dial("localhost:25")
	if err != nil {
		return err
	}
	defer client.Close()

	// Set the sender and receiver
	senderEmail := from
	receiverEmail := to

	if err := client.Mail(senderEmail); err != nil {
		return err
	}
	if err := client.Rcpt(receiverEmail); err != nil {
		return err
	}

	// Open the attachment file if provided
	var attachmentContent []byte
	if attachmentPath != "" {
		// Open the attachment file
		file, err := os.Open(attachmentPath)
		if err != nil {
			return err
		}
		defer file.Close()

		// Read the attachment file content
		attachmentContent, err = os.ReadFile(attachmentPath)
		if err != nil {
			return err
		}
	}

	// Compose the email message
	message := fmt.Sprintf("From: %s\r\n", senderEmail) +
		fmt.Sprintf("To: %s\r\n", receiverEmail) +
		fmt.Sprintf("Subject: %s\r\n", subject) +
		"MIME-version: 1.0;\r\n" +
		fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n\r\n", "boundary") +
		"--boundary\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		messageBody + "\r\n\r\n"

	// If attachment provided, include it in the message
	if attachmentPath != "" {
		// Encode the attachment as base64
		encodedAttachment := base64.StdEncoding.EncodeToString(attachmentContent)

		message += "--boundary\r\n" +
			"Content-Type: application/octet-stream\r\n" +
			"Content-Transfer-Encoding: base64\r\n" +
			"Content-Disposition: attachment; filename=\"" + filepath.Base(attachmentPath) + "\"\r\n\r\n" +
			encodedAttachment + "\r\n\r\n"
	}

	message += "--boundary--\r\n"

	// Send the email data
	w, err := client.Data()
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	return nil
}
