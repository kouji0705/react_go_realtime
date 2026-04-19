package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

func main() {
	// /events エンドポイントにハンドラを登録
	http.HandleFunc("/events", sseHandler)

	fmt.Println("SSE Server is running on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// 1. SSEに必要なHTTPヘッダーを設定
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// React（別ポート）から通信できるようにCORSを許可
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// 2. http.Flusherインターフェースに変換
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	fmt.Println("Client connected!")

	// 2秒ごとに時間を刻むティッカーを作成
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// 3. クライアントが切断するまでループ
	for {
		select {
		case <-r.Context().Done():
			// クライアント（ブラウザなど）が接続を切った場合の処理
			fmt.Println("Client disconnected")
			return
		case t := <-ticker.C:
			// データを書き込む (ルール: "data: " で始め、"\n\n" で終わる)
			msg := fmt.Sprintf(`{"time": "%s", "message": "Ping!"}`, t.Format("15:04:05"))
			fmt.Fprintf(w, "data: %s\n\n", msg)

			// バッファに溜めず、即座にクライアントへ押し出す（Flush）
			flusher.Flush()
		}
	}
}