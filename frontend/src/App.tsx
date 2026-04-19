import { useEffect, useState } from 'react'

function App() {
  const [messages, setMessages] = useState<string[]>([])

  useEffect(() => {
    // 1. Goのサーバー（8080番）にSSE接続を開始
    const eventSource = new EventSource('http://localhost:8080/events')

    // 2. サーバーからメッセージが届いた時の処理
    eventSource.onmessage = (event) => {
      // event.data にサーバーからの文字列が入っている
      console.log('New message:', event.data)
      setMessages((prev) => [...prev, event.data])
    }

    // 3. エラー（サーバーダウンなど）の処理
    eventSource.onerror = (err) => {
      console.error('SSE Error:', err)
    }

    // 4. クリーンアップ：コンポーネントがアンマウントされたら接続を閉じる
    return () => {
      eventSource.close()
    }
  }, [])

  return (
    <div>
      <h1>SSE Receiver (Step 2)</h1>
      <ul>
        {messages.map((msg, i) => (
          <li key={i}>{msg}</li>
        ))}
      </ul>
    </div>
  )
}

export default App