import { Message } from "../App"

interface MessageProps {
  message: Message,
  isConnectedUser: boolean,
}

export function UserMessage({ message, isConnectedUser }: MessageProps) {
  return (
    <span className="flex flex-col">
      <span className="font-bold">{isConnectedUser ? 'You' : message.user.name}</span>
      <span className="font-medium">{message.message}</span>
    </span>
  )
}