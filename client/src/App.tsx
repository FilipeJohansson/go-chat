import { Chat } from "./components/Chat"
import { useWebSocket } from "./internal/useWebSocket"

export function App() {
  const { isConnected, connectedUser, messages, usersOnline, sendMessage } = useWebSocket()

  const handleSend = (message: string) => {
    if (!isConnected) return
    sendMessage(message)
  }

  return (
    <div className="flex justify-center w-full p-5">
      <Chat
        connectedUserId={connectedUser?.id ?? -1}
        messages={messages}
        usersOnline={usersOnline}
        onSendMessage={(message: string) => handleSend(message)}
      />
    </div>
  )
}
