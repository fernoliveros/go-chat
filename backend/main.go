package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

var messages []string
var msgChannels []chan struct{}

const FE_URL = "http://localhost:8080"

func main() {
	registerDefaultHandler()
	registerLoginHandler()
	registerSendHandler()
	setupMessagesStream()

	http.ListenAndServe(":8080", nil)
}

func setupMessagesStream() {
	http.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The /messages handler has been invoked!")
		setupCORS(w)

		if r.Method == "OPTIONS" {
			fmt.Println("This is the options conditional")
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
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The send message handler has been invoked!")
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
			Message string `json:"message"`
		}
		if err := json.Unmarshal(jsonBytes, &body); err != nil {
			fmt.Printf("ERROR unmarshalling response body %v", err)

		}
		fmt.Printf("Received post with %v\n", body.Message)

		messages = append(messages, string(body.Message))
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
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The login handler has been invoked!")
		setupCORS(w)

		if r.Method == "OPTIONS" {
			fmt.Println("This is the options conditional")
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
		fmt.Printf("Received post with %v\n", body)

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
	fs := http.FileServer(http.Dir("../dist/browser"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The default handler has been invoked!")
		setupCORS(w)

		if r.Method == "OPTIONS" {
			fmt.Println("This is the options conditional")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != "GET" {
			http.Error(w, "we only allow gets", http.StatusMethodNotAllowed)
		}

		fs.ServeHTTP(w, r)
	})
}

func setupCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", FE_URL)
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
}
