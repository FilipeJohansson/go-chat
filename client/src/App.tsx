import { Route, BrowserRouter as Router, Routes } from "react-router";
import { useWebSocket } from "./internal/useWebSocket";
import { Chat } from "./pages/Chat";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";

export function App() {
  const { isConnected, connectedUser, messages, usersOnline, sendMessage } = useWebSocket()

  const handleSend = (message: string) => {
    if (!isConnected) return
    sendMessage(message)
  }

  return (
    <>
      <Router>
        <Routes>
          <Route index path="/" element={
            <Chat
              connectedUserId={connectedUser?.id ?? -1}
              messages={messages}
              usersOnline={usersOnline}
              onSendMessage={(message: string) => handleSend(message)}
            />
          } />
          <Route index path="/login" element={<Login />} />
          <Route index path="/register" element={<Register />} />
        </Routes>
      </Router>
    </>
  )
}
