package main

import (
  "encoding/json"
  "fmt"
  "net/http"
  "time"
)

type Message struct {
  Text string `json:"message"`
  User string `json:"username"`
}

var (
  messages = make(chan Message)
  clients  = make([]chan Message, 0)
)

func handleSSE(w http.ResponseWriter, r *http.Request) {
  w.Header().Set("Content-Type", "text/event-stream")
  w.Header().Set("Cache-Control", "no-cache")
  w.Header().Set("Connection", "keep-alive")

  client := make(chan Message)
  defer close(client)

  clients = append(clients, client)

  for msg := range client {

    msgData := make(map[string]interface{})
    msgData["text"] = msg.Text
    msgData["username"] = msg.User
    msgData["createdAt"] = time.Now().UTC()

    msgJSON, err := json.Marshal(msgData)

    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

    fmt.Fprintf(w, "data: %s\n\n", msgJSON)

    if f, ok := w.(http.Flusher); ok {
      f.Flush()
    }
  }
}

func handleNewMessage(w http.ResponseWriter, r *http.Request) {
  var msg Message
  err := json.NewDecoder(r.Body).Decode(&msg)
  if err != nil {
    http.Error(w, err.Error(), http.StatusBadRequest)
    return
  }

  if msg.User == "" {
    http.Error(w, "username is required", http.StatusBadRequest)
  } else if msg.Text == "" {
    http.Error(w, "message is required", http.StatusBadRequest)
  } else {
    messages <- msg

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"status": "success"})
  }
}

func broadcastMessages() {
  for {
    msg := <-messages

    for _, client := range clients {
      client <- msg
    }
  }
}

func main() {
  go broadcastMessages()

  fs := http.FileServer(http.Dir("html"))
  http.Handle("/", fs)

  http.HandleFunc("/sse", handleSSE)

  http.HandleFunc("/message", handleNewMessage)

  http.ListenAndServe(":8080", nil)
}
