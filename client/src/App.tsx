import { useState } from "react";
import { Chat } from "./components/Chat";
import { Packet } from "./proto/packets";

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

  let data: Uint8Array = new Uint8Array([8, 139, 6, 18, 16, 10, 14, 72, 101, 108, 108, 111, 44, 32, 70, 105, 108, 105, 112, 101, 33])
  let packet = Packet.decode(data)
  console.log(packet)

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
