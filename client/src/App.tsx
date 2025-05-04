import { useState } from "react";
import { Chat } from "./components/Chat";
import { ChatMessage, Packet } from "./proto/packets";

export interface User {
  id: number,
  name: string,
}

export interface Message {
  timestamp: number,
  user: User,
  message: string,
}

export function App() {
  const [connectedUser, setConnectedUser] = useState({ id: 0, name: 'Filipe Johansson' } as User);
  const [usersOnline, setUsersOnline] = useState([{ id: 0, name: 'Filipe Johansson' }] as User[])
  const [messages, setMessages] = useState([{ timestamp: Date.now(), user: { id: 0, name: 'Filipe Johansson'}, message: 'This is a test'}] as Message[])

  let chatMessage: ChatMessage = ChatMessage.create({
    msg: "Hello, Filipe!"
  })
  let packet: Packet = Packet.create({
    senderId: 779,
    chat: chatMessage
  })
  
  let data = Packet.encode(packet).finish()
  console.log('data', data)

  const onSendMessage = (message: string) => {
    onReceiveMessage({ user: connectedUser, message })
  }

  const onReceiveMessage = (data: { user: User, message: string }) => {
    setMessages((prevList) => [
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
        messages={messages}
        usersOnline={usersOnline}
        onSendMessage={(message: string) => onSendMessage(message)}
      />
    </div>
  );
}
