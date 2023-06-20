package main

import (
	"context"
	"log"

	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Message struct {
  Text string `json:"message"`
  User string `json:"username"`
}

var (
  messages = make(chan Message)
  clients  = make([]chan Message, 0)
)

var dbMessages *mongo.Collection
var ctx = context.TODO()

func init() {
  clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")

  client, err := mongo.Connect(ctx, clientOptions)

  if err != nil {
    log.Fatal(err)
  }

  err = client.Ping(ctx, nil)

  if err != nil {
    log.Fatal(err)
  }

  dbMessages = client.Database("chat").Collection("messages")
  cur, err := dbMessages.Find(ctx, bson.D{{}})
  if err != nil {
    log.Fatal(err)
  }
  defer cur.Close(ctx)
  for cur.Next(ctx) {
    var result bson.M
    err := cur.Decode(&result)
    if err != nil {
      log.Fatal(err)
    }
    fmt.Println(result)
  }
  if err := cur.Err(); err != nil {
    log.Fatal(err)
  }

  fmt.Println("Connected to MongoDB!")
}

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

    fmt.Fprintf(w, "event: message\n\ndata: %s\n\n", msgJSON)

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
    _, err := dbMessages.InsertOne(ctx, msg)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }

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
