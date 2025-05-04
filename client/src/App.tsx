import { useState } from "react";
import { Chat } from "./components/Chat";

export interface User {
  id: number,
  name: string,
}

export interface ChatMessage {
  timestamp: number,
  user: User,
  message: string,
}

export function App() {
  const [connectedUser, setConnectedUser] = useState({ id: 0, name: 'Filipe Johansson' } as User);
  const [usersOnline, setUsersOnline] = useState([{ id: 0, name: 'Filipe Johansson' }] as User[])
  const [chatMessages, setChatMessages] = useState([{ timestamp: Date.now(), user: { id: 0, name: 'Filipe Johansson'}, message: 'This is a test'}] as ChatMessage[])

  const onSendMessage = (message: string) => {
    onReceiveMessage({ user: connectedUser, message })
  }

  const onReceiveMessage = (data: { user: User, message: string }) => {
    setChatMessages((prevList) => [
      ...prevList,
      { timestamp: Date.now(), user: data.user, message: data.message }
    ])
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
        connectedUserId={connectedUser.id}
        messages={chatMessages}
        usersOnline={usersOnline}
        onSendMessage={(message: string) => onSendMessage(message)}
      />
    </div>
  );
}
