import { useEffect, useState } from "react"
import { useNavigate } from "react-router"
import { ChatMessage, IdMessage, LogoutRequestMessage, Message as Msg, Packet } from "../proto/packets"
import { Tokens, clearTokens, getTokens } from "./tokens"
import { WebSocketClient } from "./websocket"

export interface User {
  id: number,
  name: string,
}

export interface Room {
  roomId: number,
  ownerId: string,
  name: string,
}

export interface Message {
  timestamp: Date,
  user: User,
  message: string,
}

export function useWebSocket(roomId: string | null) {
  const navigate = useNavigate()

  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<Message[]>([])
  const [connectedRoom, setConnectedRoom] = useState<Room | undefined>(undefined)
  const [connectedUser, setConnectedUser] = useState<User | undefined>(undefined)
  const [usersOnline, setUsersOnline] = useState<User[]>([])

  useEffect(() => {
    if (!roomId) {
      console.log("Error getting room id")
      navigate("/rooms")
    }

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
    client.connect(`ws://localhost:8080/ws?room=${roomId}&token=${accessToken}`)

    return () => {
      client.clear()
    }
  }, [])

  useEffect(() => {
    if (connectedUser) handleRegisterMessage(connectedUser)
  }, [connectedUser])

  const handleIdMsg = (idMsg: IdMessage) => {
    if (!idMsg.room) {
      console.log("Unable to get the room")
      return
    }

    const clientId: number = idMsg.id
    const username: string = idMsg.username
    const roomId: number = idMsg.room.id
    const roomOwnerId: string = idMsg.room.ownerId
    const roomName: string = idMsg.room.name
    setConnectedUser({ id: clientId, name: username})
    setConnectedRoom({ roomId: roomId, ownerId: roomOwnerId, name: roomName })
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
    const roomId: number = connectedRoom!.roomId
    const senderId: number = connectedUser!.id
    const senderUsername: string = connectedUser!.name
    const timestamp: Date = new Date()
    const chatMessage: ChatMessage = { timestamp, senderUsername, msg }
    const packet: Packet = Packet.create<Packet>({ roomId, senderId, chat: chatMessage })
    WebSocketClient.getInstance().send(packet)

    addMessage({ timestamp, user: connectedUser!, message }) // Add a copy of the message on our own side
  }

  const disconnect = () => {
    const refreshToken: string = getTokens()!.refreshToken
    const logoutRequest: LogoutRequestMessage = LogoutRequestMessage.create()
    const message: Msg = Msg.create<Msg>({ logout: logoutRequest })
    const binary: Uint8Array = Msg.encode(message).finish()
    fetch("http://localhost:8080/logout", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        "Authorization": refreshToken,
      },
      body: binary,
      mode: "cors"
    })
    .catch(error => console.error("Erro:", error))
    clearTokens()

    WebSocketClient.getInstance().close(1000, "user logout")
  }

  return {
    isConnected,
    messages,
    usersOnline,
    connectedUser,
    sendMessage,
    disconnect,
  }
}
