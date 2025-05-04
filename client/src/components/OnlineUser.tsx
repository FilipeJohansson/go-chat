import { User2 } from "lucide-react";

interface OnlineUserProps {
  name: string,
  isConnectedUser: boolean,
}

export function OnlineUser({ name, isConnectedUser }: OnlineUserProps) {
  return (
    <span className={`flex flex-row gap-0.5 items-center ${isConnectedUser && 'font-bold'}`}>
      <User2 className="w-4 h-4" />
      {name}
    </span>
  )
}