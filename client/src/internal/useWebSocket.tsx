import { useEffect, useState } from "react"
import { Packet } from "../proto/packets"
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

  useEffect(() => {
    const client = WebSocketClient.getInstance()

    const onConnected = () => {
      setIsConnected(true)
      console.log("Connected to WebSocket server")

      setConnectedUser({ id: 0, name: 'User 0' })
    }

    const onDisconnected = () => {
      setIsConnected(false)
      console.log("Disconnected from WebSocket server")
    }

    const onPacketReceived = (packet: Packet) => {
      if (packet.chat) {
        const message = packet.chat.msg
        setMessages((prevMessages) => [
          ...prevMessages,
          { timestamp: Date.now(), user: { id: packet.senderId, name: `User ${packet.senderId}` }, message },
        ])
      }
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

  const sendMessage = (message: string) => {
    if (!message.trim()) return
    const chatMessage = { msg: message.trim() }
    const packet = Packet.create({ senderId: 0, chat: chatMessage })
    WebSocketClient.getInstance().send(packet)
  }

  return {
    isConnected,
    messages,
    connectedUser,
    sendMessage,
  }
}
