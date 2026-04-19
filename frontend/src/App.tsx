import { useEffect, useState } from 'react'

type Message = {
  text: string
  time: string
}

function App() {
  const [messages, setMessages] = useState<Message[]>([])
  const [inputText, setInputText] = useState('')

  useEffect(() => {
    const eventSource = new EventSource('http://localhost:8080/events')

    eventSource.onmessage = (event) => {
      const newMessage = JSON.parse(event.data) as Message
      setMessages((prev) => [...prev, newMessage])
    }

    return () => eventSource.close()
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!inputText) return

    // メッセージをPOST送信（これ自体はただのHTTPリクエスト）
    await fetch('http://localhost:8080/messages', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ text: inputText }),
    })

    setInputText('')
  }

  return (
    <div style={{ padding: '20px' }}>
      <h1>SSE Chat Room</h1>
      
      {/* 送信フォーム */}
      <form onSubmit={handleSubmit} style={{ marginBottom: '20px' }}>
        <input 
          value={inputText} 
          onChange={(e) => setInputText(e.target.value)} 
          placeholder="メッセージを入力..."
        />
        <button type="submit">送信</button>
      </form>

      {/* メッセージリスト */}
      <div style={{ border: '1px solid #ccc', height: '300px', overflowY: 'scroll' }}>
        {messages.map((m, i) => (
          <div key={i} style={{ padding: '5px 10px' }}>
            <strong>[{m.time}]</strong> {m.text}
          </div>
        ))}
      </div>
    </div>
  )
}

export default App