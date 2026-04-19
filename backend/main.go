package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

// Hub: 接続中のクライアント管理とメッセージの中継を行う
type Hub struct {
	// 接続中の各クライアントにデータを送るためのチャネルを保持
	// 型: map[チャネル]空構造体
	clients map[chan string]struct{}
	// クライアントの追加・削除・配信を安全に行うためのミューテックス
	mu sync.Mutex
}

// 全クライアントに一斉配信
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
	// テスト用に、ここを叩くと全員に通知が飛ぶエンドポイント
	http.HandleFunc("/notify", notifyHandler)

	fmt.Println("SSE Hub Server running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

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

	// このクライアント専用のチャネルを作成
	clientChan := make(chan string)

	// Hubに登録
	hub.mu.Lock()
	hub.clients[clientChan] = struct{}{}
	hub.mu.Unlock()

	// 接続終了時にHubから削除
	defer func() {
		hub.mu.Lock()
		delete(hub.clients, clientChan)
		hub.mu.Unlock()
		close(clientChan)
		fmt.Println("Client removed from Hub")
	}()

	fmt.Println("New client connected!")

	for {
		select {
		case <-r.Context().Done():
			return
		case msg := <-clientChan:
			// Hubから届いたメッセージをSSE形式で書き出す
			fmt.Fprintf(w, "data: %s\n\n", msg)
			flusher.Flush()
		}
	}
}

// Step 3の肝: このAPIを叩くと接続者全員にPushされる
func notifyHandler(w http.ResponseWriter, r *http.Request) {
	msg := r.URL.Query().Get("msg")
	if msg == "" {
		msg = "誰かがアクションを起こしました！"
	}

	// 全員に通知を飛ばす
	hub.broadcast(fmt.Sprintf(`{"message": "%s"}`, msg))

	fmt.Fprintln(w, "Broadcasted!")
}