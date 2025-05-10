import { LogOut, Send } from "lucide-react";
import { useState } from "react";
import { OnlineUser } from "../components/OnlineUser";
import { Painel } from "../components/Painel";
import { UserMessage } from "../components/UserMessage";
import { Message, User, useWebSocket } from "../internal/useWebSocket";

export function Chat() {
  const queryParameters = new URLSearchParams(window.location.search)
  const roomId = queryParameters.get("id")

  const { isConnected, connectedUser, messages, usersOnline, sendMessage, disconnect } = useWebSocket(roomId)

  const [mesageContent, setMessageContent] = useState('')

  const handleSendMessage = () => {
    if (!isConnected) return

    setMessageContent('')
    sendMessage(mesageContent)
  }

  return (
    <div className="flex justify-center w-full p-5">
      <div className="flex flex-col gap-3 p-2 w-[700px] h-[500px] bg-[url(/src/assets/background.png)] bg-no-repeat bg-cover bg-center rounded-md">
        <div className="grid grid-cols-4 gap-3 h-[80%]">
          {/* Messages */}
          <Painel className="col-span-3 flex flex-col overflow-y-auto overflow-x-hidden">
            {messages.map((m: Message) => (
              <UserMessage key={`${m.user.id}_${m.timestamp.getTime()}`} message={m} isConnectedUser={connectedUser?.id === m.user.id} />
            ))}
          </Painel>

          {/* People Online */}
          <Painel className="flex flex-col gap-0.5 overflow-y-auto overflow-x-hidden">
            {usersOnline.map((u: User) => (
              <OnlineUser key={u.id} name={u.name} isConnectedUser={connectedUser?.id === u.id} />
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
            onClick={handleSendMessage}
          >
            <Send className="w-5 h-5" />
          </button>
        </div>

        <div className="flex flex-row justify-between">
          <div className="flex items-center gap-1">
            <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-400' : 'animate-pulse bg-red-400'}`}></div>
            <span className="text-white text-xs">{isConnected ? 'Connected' : 'Disconnected'}</span>
          </div>
          
          <div className="flex items-center gap-1">
            <button
              className="flex fle-row items-center rounded-sm bg-red-500 p-0.5 text-white text-xs"
              onClick={disconnect}
            >
              <LogOut className="w-3 h-3" />
              <span>Leave</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
