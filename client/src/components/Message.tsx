import { ChatMessage } from "../App"

interface MessageProps {
  message: ChatMessage,
  isConnectedUser: boolean,
}

export function Message({ message, isConnectedUser }: MessageProps) {
  return (
    <span className="flex flex-col">
      <span className="font-bold">{isConnectedUser ? 'You' : message.user.name}</span>
      <span className="font-medium">{message.message}</span>
    </span>
  )
}