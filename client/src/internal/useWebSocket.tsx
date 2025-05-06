import { useEffect, useState } from "react"
import { ChatMessage, IdMessage, Packet } from "../proto/packets"
import { WebSocketClient } from "./websocket"

export interface User {
  id: number,
  name: string,
}

export interface Message {
  timestamp: number,
  user: User,
  message: string,
}

export function useWebSocket() {
  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [connectedUser, setConnectedUser] = useState<User | undefined>(undefined)
  const [usersOnline, setUsersOnline] = useState<User[]>([])

  useEffect(() => {
    const client = WebSocketClient.getInstance()

    const onConnected = () => {
      setIsConnected(true)
      console.log("Connected to WebSocket server")
    }

    const onDisconnected = () => {
      setIsConnected(false)
      console.log("Disconnected from WebSocket server")

      //? Automatic reconnection
    }

    const onPacketReceived = (packet: Packet) => {
      console.log("Packet received", packet)
      const senderId = packet.senderId
      if (packet.id) handleIdMsg(packet.id)
      else if (packet.chat) handleChatMessage(senderId, packet.chat)
      else if (packet.register) handleRegisterMessage(packet.register.id)
      else if (packet.unregister) handleUnregisterMessage(packet.unregister.id)
    }

    client.configure({
      onConnected,
      onDisconnected,
      onPacket: onPacketReceived,
    })

    client.connect("ws://localhost:8080/ws")

    return () => {
      client.clear()
    }
  }, [])

  useEffect(() => {
    if (connectedUser) handleRegisterMessage(connectedUser.id)
  }, [connectedUser])

  const handleIdMsg = (idMsg: IdMessage) => {
    const clientId = idMsg.id
    setConnectedUser({ id: clientId, name: `Client ${clientId}`})
  }

  const handleChatMessage = (senderId: number, message: ChatMessage) => {
    addMessage(senderId, message.msg)
  }

  const handleRegisterMessage = (id: number) => {
    const newUser: User = { id, name: `Client ${id}` }
    setUsersOnline((prev) => {
      const merged = [...prev, newUser]

      const unique = Array.from(
        new Map(merged.map(user => [user.id, user])).values()
      )

      return unique.sort((a, b) => (
        a.id === connectedUser?.id ? -1 :
        b.id === connectedUser?.id ? 1 :
        0
      ))
    })
  }

  const handleUnregisterMessage = (unregisterId: number) => {
    setUsersOnline((prev) => [
      ...prev
    ].filter((user: User) => user.id !== unregisterId))
  }

  const addMessage = (senderId: number, message: string) => {
    setMessages((prevMessages) => [
      ...prevMessages,
      { timestamp: Date.now(), user: { id: senderId, name: `Client ${senderId}` }, message },
    ])
  }

  const sendMessage = (message: string) => {
    const msg = message.trim()
    if (!msg) return
    const senderId = connectedUser!.id
    const chatMessage = { msg }
    const packet = Packet.create({ senderId: senderId, chat: chatMessage })
    WebSocketClient.getInstance().send(packet)

    addMessage(senderId, msg) // Add a copy of the message on our own side
  }

  return {
    isConnected,
    messages,
    usersOnline,
    connectedUser,
    sendMessage,
  }
}
