package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
)

type Hub struct {
	clients map[chan string]struct{}
	mu      sync.Mutex
}

func (h *Hub) broadcast(msg string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for clientChan := range h.clients {
		clientChan <- msg
	}
}

var hub = &Hub{
	clients: make(map[chan string]struct{}),
}

func main() {
	http.HandleFunc("/events", sseHandler)
	// メッセージ投稿用エンドポイント
	http.HandleFunc("/messages", messageHandler)

	fmt.Println("Chat Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func messageHandler(w http.ResponseWriter, r *http.Request) {
	// CORS設定
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == "POST" {
		var body struct {
			Text string `json:"text"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// JSONを文字列のまま全員に一斉配信
		msgJson, _ := json.Marshal(map[string]string{
			"text": body.Text,
			"time": "18:30", // 本来は現在時刻など
		})
		hub.broadcast(string(msgJson))

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "Sent!")
	}
}

// sseHandlerはStep 3と同じものを使用
func sseHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	clientChan := make(chan string)
	hub.mu.Lock()
	hub.clients[clientChan] = struct{}{}
	hub.mu.Unlock()

	defer func() {
		hub.mu.Lock()
		delete(hub.clients, clientChan)
		hub.mu.Unlock()
		close(clientChan)
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}