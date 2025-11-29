package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"

	nats "github.com/nats-io/nats.go"
)

var messages []string

func main() {
	natSubChannel := "groupchat"
	natPubChannel := "groupchat"
	httpPort := "8080"

	
	registerLoginHandler()

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		fmt.Printf("Error connecting to nats server: %v", err)
		return
	}

	registerSendHandler(nc, natPubChannel)
	setupMessagesStream(nc, natSubChannel)

	startHtmxServer(httpPort)
}

func setupMessagesStream(nc *nats.Conn, natSubChannel string) {
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

		ch := make(chan *nats.Msg, 64)
		sub, err := nc.ChanSubscribe(natSubChannel, ch)
		if err != nil {
			fmt.Printf("Error subscribing to the natsubChannel: %v", err.Error())
		}

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Error initializing flusher, streaming may be unsupported", http.StatusInternalServerError)
			return
		}

		ctx := r.Context()

		for {
			select {
			case <-ctx.Done():
				fmt.Println("Client disconnected")
				sub.Unsubscribe()
				return
			case msg := <-ch:

				fmt.Sprintf("msg not used but received: %s", string(msg.Data))

				messages = append(messages, string(msg.Data))

				ssePayload := fmt.Sprintf("event:message\ndata: %s\n\n", strings.Join(messages, ","))

				// fmt.Printf("About to write to the messages stream:\n %s", ssePayload)
				_, err = w.Write([]byte(ssePayload))
				if err != nil {
					fmt.Printf("Error writting data to the messages stream: %v", err.Error())
				}
				flusher.Flush()
			}
		}
	})
}

func startHtmxServer(port string) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("The default handler has been invoked!")
		tmpl, err := template.ParseFiles("html/index.html")
		if err != nil {
			fmt.Printf("Error parsing index.html %v", err.Error())
			http.Error(w, "Error parsing index.html", http.StatusInternalServerError)
			return
		}

		data := struct{ Message string }{
			Message: "Welcome to simple messaging app",
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			fmt.Printf("Error executing data struct into index.html %v \n", err.Error())
			http.Error(w, "Error executing data struct into index.html", http.StatusInternalServerError)
			return
		}
	})
	http.ListenAndServe(fmt.Sprintf(":%s", port), nil)
}

func registerSendHandler(nc *nats.Conn, natPubChannel string) {
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

		var body struct{ 
			Message string `json:"message"`
		}
		if err := json.Unmarshal(jsonBytes, &body); err != nil {
			fmt.Printf("ERROR unmarshalling response body %v", err)

		}
		fmt.Printf("Received post with %v\n", body.Message)


		nc.Publish(natPubChannel, []byte(body.Message))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) 

		data := struct{ Message string `json:"message"`}{
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

		var body struct{ 
			Username string `json:"username"`
			Password string `json:"password"`
		}
		if err := json.Unmarshal(jsonBytes, &body); err != nil {
			fmt.Printf("ERROR unmarshalling response body %v", err)

		}
		fmt.Printf("Received post with %v\n", body)


		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) 

		data := struct{ Message string `json:"message"`}{
			Message: "successful login",  
		}
		err = json.NewEncoder(w).Encode(data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	})
}

func setupCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:4200")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true") 
}
