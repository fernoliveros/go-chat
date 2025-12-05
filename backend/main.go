package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

var messages []string
var msgChannels []chan struct{}

var CORS_ALLOWED_ORIGIN string

func main() {
	// could use https://github.com/joho/godotenv for setting env vars
	CORS_ALLOWED_ORIGIN = os.Getenv("CORS_ALLOWED_ORIGIN")

	registerDefaultHandler()
	registerLoginHandler()
	registerSendHandler()
	setupMessagesStream()

	http.ListenAndServe(":8080", nil)
}

func setupMessagesStream() {
	http.HandleFunc("/api/messages", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Error initializing flusher, streaming may be unsupported", http.StatusInternalServerError)
			return
		}

		myChan := make(chan struct{})
		msgChannels = append(msgChannels, myChan)
		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Client disconnected")
				return
			case <-myChan:
				ssePayload := fmt.Sprintf("event:message\ndata: %s\n\n", strings.Join(messages, ","))
				_, err := w.Write([]byte(ssePayload))
				if err != nil {
					fmt.Printf("Error writting data to the messages stream: %v", err.Error())
				}
				flusher.Flush()
			}
		}
	})
}

func registerSendHandler() {
	http.HandleFunc("/api/send", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "we only allow posts", http.StatusMethodNotAllowed)
		}

		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("ERROR reading response body %v", err)
		}

		var body struct {
			Message  string `json:"message"`
			Username string `json:"username"`
		}
		if err := json.Unmarshal(jsonBytes, &body); err != nil {
			fmt.Printf("ERROR unmarshalling response body %v", err)

		}

		userMessage := fmt.Sprintf("%s: %s", body.Username, body.Message)
		messages = append(messages, userMessage)

		for _, channel := range msgChannels {
			channel <- struct{}{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		data := struct {
			Message string `json:"message"`
		}{
			Message: "successfully sent message",
		}
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func registerLoginHandler() {
	http.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "POST" {
			http.Error(w, "we only allow posts", http.StatusMethodNotAllowed)
		}

		jsonBytes, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Printf("ERROR reading response body %v", err)
		}

		var body struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(jsonBytes, &body); err != nil {
			fmt.Printf("ERROR unmarshalling response body %v", err)

		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		data := struct {
			Message string `json:"message"`
		}{
			Message: "successful login",
		}
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func registerDefaultHandler() {
	buildPath := "../dist/browser"
	fs := http.FileServer(http.Dir(buildPath))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		setupCORS(w)

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		if _, err := os.Stat(filepath.Join(buildPath, r.URL.Path)); os.IsNotExist(err) {
			http.ServeFile(w, r, filepath.Join(buildPath, "index.html"))
			return
		}
		fs.ServeHTTP(w, r)
	})
}

func setupCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
