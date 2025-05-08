import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { ChatMessage, IdMessage, Packet } from "../proto/packets";
import { Tokens, clearTokens, getTokens } from "./tokens";
import { WebSocketClient } from "./websocket";

export interface User {
  id: number,
  name: string,
}

export interface Message {
  timestamp: Date,
  user: User,
  message: string,
}

export function useWebSocket() {
  const navigate = useNavigate();

  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [connectedUser, setConnectedUser] = useState<User | undefined>(undefined)
  const [usersOnline, setUsersOnline] = useState<User[]>([])

  useEffect(() => {
    const client: WebSocketClient = WebSocketClient.getInstance()

    const onConnected = () => {
      setIsConnected(true)
      console.log("Connected to WebSocket server")
    }

    const onDisconnected = () => {
      setIsConnected(false)
      console.log("Disconnected from WebSocket server")

      navigate("/refresh")
    }

    const onPacketReceived = (packet: Packet) => {
      console.log("Packet received", packet)
      if (packet.id) handleIdMsg(packet.id)
      else if (packet.chat) handleChatMessage({ id: packet.senderId, name: packet.chat.senderUsername }, packet.chat)
      else if (packet.register) handleRegisterMessage({ id: packet.register.id, name: packet.register.username })
      else if (packet.unregister) handleUnregisterMessage(packet.unregister.id)
    }

    client.configure({
      onConnected,
      onDisconnected,
      onPacket: onPacketReceived,
    })

    const tokens: Tokens | undefined = getTokens()
    if (!tokens) {
      clearTokens()
      navigate("/login")
      console.log("Unable to find tokens to connect")
      return
    }

    const accessToken: string = tokens.accessToken
    client.connect(`ws://localhost:8080/ws?token=${accessToken}`)

    return () => {
      client.clear()
    }
  }, [])

  useEffect(() => {
    if (connectedUser) handleRegisterMessage(connectedUser)
  }, [connectedUser])

  const handleIdMsg = (idMsg: IdMessage) => {
    const clientId: number = idMsg.id
    const username: string = idMsg.username
    setConnectedUser({ id: clientId, name: username})
  }

  const handleChatMessage = (user: User, msg: ChatMessage) => {
    const message: Message = { timestamp: msg.timestamp || new Date(), user, message: msg.msg }
    addMessage(message)
  }

  const handleRegisterMessage = (user: User) => {
    const newUser: User = user
    setUsersOnline((prev) => {
      const merged: User[] = [...prev, newUser]

      const unique: User[] = Array.from(
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

  const addMessage = (message: Message) => {
    setMessages((prevMessages) => [
      ...prevMessages,
      message
    ])
  }

  const sendMessage = (message: string) => {
    const msg: string = message.trim()
    if (!msg) return
    const senderId: number = connectedUser!.id
    const senderUsername: string = connectedUser!.name
    const timestamp: Date = new Date()
    const chatMessage: ChatMessage = { timestamp, senderUsername, msg }
    const packet: Packet = Packet.create({ senderId, chat: chatMessage })
    WebSocketClient.getInstance().send(packet)

    addMessage({ timestamp, user: connectedUser!, message }) // Add a copy of the message on our own side
  }

  return {
    isConnected,
    messages,
    usersOnline,
    connectedUser,
    sendMessage,
  }
}
