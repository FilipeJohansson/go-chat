import { useEffect, useState } from "react"
import { useNavigate } from "react-router"
import { getTokens } from "../internal/tokens"
import { Room } from "../internal/useWebSocket"
import { Message, NewRoomRequestMessage, RoomsRequestMessage } from "../proto/packets"

export function Rooms() {
  const navigate = useNavigate()
  const [rooms, setRooms] = useState<Room[]>([])
  const [roomName, setRoomName] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (!getTokens()?.accessToken) {
      console.log("Refresh token not set")
      return
    }

    getRooms()
  }, [])

  const getRooms = () => {
    const accessToken: string = getTokens()!.accessToken
    const roomsReq: RoomsRequestMessage = RoomsRequestMessage.create()
    const message: Message = Message.create({ roomsRequest: roomsReq })
    const binary: Uint8Array = Message.encode(message).finish()
    fetch("http://localhost:8080/rooms", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        "Authorization": accessToken,
      },
      body: binary,
      mode: "cors"
    })
    .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
    .then((buffer: ArrayBuffer) => {
      const data = new Uint8Array(buffer)
      const message: Message = Message.decode(data)
      if (!message.roomsResponse) return
      console.log("rooms", message.roomsResponse)

      const rooms: Room[] = message.roomsResponse.rooms
      setRooms(rooms)
    })
    .catch(error => console.error("Erro:", error))
  }

  const goToRoom = (roomId: number) => {
    navigate(`/chat?id=${roomId}`)
  }

  const createRoom = () => {
    const accessToken: string = getTokens()!.accessToken
    const newRoomReq: NewRoomRequestMessage = NewRoomRequestMessage.create({ name: roomName })
    const message: Message = Message.create({ newRoom: newRoomReq })
    const binary: Uint8Array = Message.encode(message).finish()
    fetch("http://localhost:8080/new-room", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        "Authorization": accessToken,
      },
      body: binary,
      mode: "cors"
    })
    .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
    .then((buffer: ArrayBuffer) => {
      const data = new Uint8Array(buffer)
      const message: Message = Message.decode(data)
      if (!message.okResponse) return

      getRooms()
    })
    .catch(error => console.error("Erro:", error))
  }

  return (
    <div className="flex  flex-col gap-10 p-20">
      <div className="flex flex-row gap-3">
        <input
          className="p-1 border rounded-md focus:outline-blue-500"
          type="text"
          placeholder="Room name"
          onChange={e => setRoomName(e.target.value)}
        />
        <button
          className="px-3 py-1 border bg-blue-500 hover:bg-pink-500 rounded-md transition"
          onClick={createRoom}
        >
          <span className="text-white">Create Room</span>
        </button>
      </div>
      <div>
        <span className="text-md font-semibold">Available Rooms ({rooms.length})</span>
        <div className="grid grid-cols-4 gap-4 mt-5">
          {rooms.map((room: Room) => (
            <div key={room.roomId} className="flex flex-row items-center justify-between p-4 border rounded-md">
              <div className="flex flex-col">
                <span className="text-gray-500 text-xs">{room.roomId}</span>
                <span className="text-blue-500 font-semibold">{room.name}</span>
              </div>
              <button
                className="px-3 py-1 border bg-blue-500 hover:bg-pink-500 rounded-md transition"
                onClick={e => goToRoom(room.roomId)}
              >
                <span className="text-white">Join</span>
              </button>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}