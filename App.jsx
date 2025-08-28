import { useEffect, useState } from "react";

function App() {
  const [messages, setMessages] = useState([]);
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    const ws = new WebSocket("ws://localhost:8080/ws"); // backend WS manzili

    ws.onopen = () => {
      setConnected(true);
      console.log("✅ WS ulandi");
    };

    ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setMessages((prev) => [...prev, data]);
      } catch (e) {
        console.error("❌ JSON parse xato:", e);
      }
    };

    ws.onclose = () => {
      setConnected(false);
      console.log("🔒 WS yopildi");
    };

    return () => ws.close();
  }, []);

  return (
    <div className="min-h-screen bg-gray-50 text-gray-900 p-6">
      <h1 className="text-2xl font-bold mb-4">Soliq Tashkilotlari Ma'lumotlari</h1>

      {!connected && (
        <p className="text-red-500">🔴 Serverga ulanish yo‘q</p>
      )}

      {messages.map((msg, idx) => (
        <div key={idx} className="mb-8 border rounded-lg p-4 bg-white shadow">
          {/* Asosiy javob */}
          {msg?.choices?.[0]?.message?.content && (
            <div className="prose max-w-none mb-4">
              <h2 className="text-lg font-semibold">Natija</h2>
              <p>{msg.choices[0].message.content}</p>
            </div>
          )}

          {/* Manbalar */}
          {msg?.citations?.length > 0 && (
            <div className="mb-4">
              <h3 className="font-medium text-gray-700 mb-2">📚 Manbalar:</h3>
              <ul className="list-disc pl-5 space-y-1">
                {msg.citations.map((url, i) => (
                  <li key={i}>
                    <a
                      href={url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="text-blue-600 hover:underline"
                    >
                      {url}
                    </a>
                  </li>
                ))}
              </ul>
            </div>
          )}

          {/* Qo‘shimcha natijalar */}
          {msg?.search_results?.length > 0 && (
            <div>
              <h3 className="font-medium text-gray-700 mb-2">🔎 Qo‘shimcha natijalar:</h3>
              <div className="grid grid-cols-1 gap-3">
                {msg.search_results.map((res, i) => (
                  <div key={i} className="p-3 border rounded bg-gray-50">
                    <a
                      href={res.url}
                      target="_blank"
                      rel="noopener noreferrer"
                      className="font-semibold text-blue-700 hover:underline"
                    >
                      {res.title}
                    </a>
                    {res.last_updated && (
                      <p className="text-sm text-gray-500">
                        Yangilangan sana: {res.last_updated}
                      </p>
                    )}
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

export default App;
