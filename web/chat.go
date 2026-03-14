// Copyright (C) by Ubaldo Porcheddu <ubaldo@eja.it>

package web

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eja/pbx/pbx"
	"github.com/eja/pbx/sys"
	"github.com/eja/tibula/log"
)

func generateChatID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func chatRouter(w http.ResponseWriter, r *http.Request) {
	const tag = "[chat]"
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method == http.MethodGet {
		query := r.URL.Query()
		if query.Get("uuid") == "" {
			newID := generateChatID()
			query.Set("uuid", newID)
			if sys.Options.ChatAudio {
				query.Set("audio", "on")
			}
			r.URL.RawQuery = query.Encode()
			http.Redirect(w, r, r.URL.String(), http.StatusFound)
			return
		}

		htmlContent, err := embeddedFiles.ReadFile("assets/chat.html")
		if err != nil {
			http.Error(w, "File not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(htmlContent)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	chatID := r.URL.Query().Get("uuid")
	if chatID == "" {
		http.Error(w, "Missing uuid parameter", http.StatusBadRequest)
		return
	}

	lang := r.URL.Query().Get("language")

	contentType := r.Header.Get("Content-Type")
	var userInput string
	var isAudio bool

	if strings.Contains(contentType, "application/json") {
		var msg struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		userInput = msg.Text

	} else if strings.Contains(contentType, "multipart/form-data") {
		isAudio = true
		if err := r.ParseMultipartForm(50 << 20); err != nil { // 50 MB max
			http.Error(w, "Error parsing form", http.StatusBadRequest)
			return
		}

		file, _, err := r.FormFile("audio")
		if err != nil {
			http.Error(w, "Error retrieving audio", http.StatusBadRequest)
			return
		}
		defer file.Close()

		tmpIn, err := os.CreateTemp(sys.Options.MediaPath, "web-asr-*.audio")
		if err != nil {
			http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}
		defer os.Remove(tmpIn.Name())

		if _, err := io.Copy(tmpIn, file); err != nil {
			http.Error(w, "Failed to save audio", http.StatusInternalServerError)
			return
		}
		tmpIn.Close()

		userInput, err = pbx.ASR(tmpIn.Name(), lang)
		if err != nil {
			log.Error(tag, "ASR internal error:", err)
			http.Error(w, "STT error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		defer os.Remove(tmpIn.Name() + ".google")
		defer os.Remove(tmpIn.Name() + ".whisper")

	} else {
		http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
		return
	}

	llmResponse, err := pbx.Text(chatID, lang, userInput)
	if err != nil {
		log.Error(tag, "Chat internal error:", err)
		http.Error(w, "LLM Processing error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if isAudio {
		tmpOut, err := os.CreateTemp(sys.Options.MediaPath, "web-tts-*.audio")
		if err != nil {
			http.Error(w, "Failed to create temp file", http.StatusInternalServerError)
			return
		}
		tmpOut.Close()
		defer os.Remove(tmpOut.Name())

		if err := pbx.TTS(llmResponse, lang, tmpOut.Name()); err != nil {
			log.Error(tag, "TTS internal error:", err)
			http.Error(w, "TTS error: "+err.Error(), http.StatusInternalServerError)
			return
		}

		audioData, err := os.ReadFile(tmpOut.Name())
		if err != nil {
			http.Error(w, "Failed to read TTS audio", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "audio/wav")
		w.Write(audioData)
	} else {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"text": llmResponse})
	}
}
