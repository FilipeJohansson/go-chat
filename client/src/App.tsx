import { useEffect, useState } from "react"
import { Chat } from "./components/Chat"
import { User, useWebSocket } from "./internal/useWebSocket"

export function App() {
  const { isConnected, connectedUser, messages, sendMessage } = useWebSocket()

  const [usersOnline, setUsersOnline] = useState<User[]>([])

  useEffect(() => {
    if (connectedUser) onUserJoin(connectedUser)
  }, [connectedUser])

  const handleSend = (message: string) => {
    if (!isConnected) return
    sendMessage(message)
  }

  const onUserJoin = (user: User) => {
    setUsersOnline((prevList) => [
      ...prevList,
      user
    ])
  }

  const onUserLeave = (userId: number) => {
    setUsersOnline((prevList) => prevList.filter((u: User) => u.id != userId))
  }

  return (
    <div className="p-5">
      <Chat
        connectedUserId={connectedUser?.id ?? -1}
        messages={messages}
        usersOnline={usersOnline}
        onSendMessage={(message: string) => handleSend(message)}
      />
    </div>
  )
}
