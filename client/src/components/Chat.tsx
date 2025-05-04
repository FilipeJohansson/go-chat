import { Send } from "lucide-react";
import { useState } from "react";
import { Message, User } from "../App";
import { OnlineUser } from "./OnlineUser";
import { Painel } from "./Painel";
import { UserMessage } from "./UserMessage";

interface ChatProps {
  connectedUserId: number,
  messages: Message[],
  usersOnline: User[],
  onSendMessage: (message: string) => void,
}

export function Chat({ connectedUserId, messages, usersOnline, onSendMessage }: ChatProps) {
  const [mesageContent, setMessageContent] = useState('')

  const sendMessage = () => {
    setMessageContent('')
    onSendMessage(mesageContent)
  }

  return (
    <div className="flex flex-col gap-3 p-2 w-[700px] h-[500px] bg-[url(/src/assets/background.png)] bg-no-repeat bg-cover bg-center rounded-md">
      <div className="grid grid-cols-4 gap-3 h-[80%]">
        {/* Messages */}
        <Painel className="col-span-3 flex flex-col">
          {messages.map((m: Message) => (
            <UserMessage key={`${m.user.id}_${m.timestamp}`} message={m} isConnectedUser={connectedUserId === m.user.id} />
          ))}
        </Painel>

        {/* People Online */}
        <Painel className="flex flex-col gap-0.5">
          {usersOnline.map((u: User) => (
            <OnlineUser key={u.id} name={u.name} isConnectedUser={connectedUserId === u.id} />
          ))}
        </Painel>
      </div>

      {/* Text Input */}
      <div className="flex gap-3 h-[20%]">
        <textarea
          value={mesageContent}
          onChange={e => setMessageContent(e.target.value)}
          className="p-1 w-full font-medium resize-none bg-zinc-100 bg-opacity-30 border border-zinc-100 border-opacity-50 rounded-md"
        />
        <button
          className="px-1.5 text-white text-sm bg-blue-700 border border-blue-900 rounded-md hover:bg-blue-500"
          onClick={sendMessage}
        >
          <Send className="w-5 h-5" />
        </button>
      </div>
    </div>
  );
}
