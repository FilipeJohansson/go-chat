import { Message, NewRoomRequestMessage, RoomsRequestMessage } from "../../proto/packets"
import { getTokens } from "./tokens"

export interface Room {
  roomId: number,
  ownerId: string,
  name: string,
}

export async function getAllRooms(): Promise<Room[]> {
  const accessToken: string = getTokens()!.accessToken
  const roomsReq: RoomsRequestMessage = RoomsRequestMessage.create()
  const message: Message = Message.create({ roomsRequest: roomsReq })
  const binary: Uint8Array = Message.encode(message).finish()

  return fetch("http://localhost:8080/rooms", {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream",
      "Authorization": accessToken,
    },
    body: binary,
    mode: "cors"
  })
  .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
  .then((buffer: ArrayBuffer): Room[] => {
    const data = new Uint8Array(buffer)
    const message: Message = Message.decode(data)
    if (!message.roomsResponse) throw new Error("Wrong message type")

    const rooms: Room[] = message.roomsResponse.rooms
    return rooms
  })
  .catch(error => {
    throw new Error(error)
  })
}

export async function createRoom(name: string): Promise<boolean> {
  const accessToken: string = getTokens()!.accessToken
  const newRoomReq: NewRoomRequestMessage = NewRoomRequestMessage.create({ name })
  const message: Message = Message.create({ newRoom: newRoomReq })
  const binary: Uint8Array = Message.encode(message).finish()
  
  return fetch("http://localhost:8080/new-room", {
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
    if (message.okResponse) return true
    else if (message.denyResponse) return false

    throw new Error("Failed to create room")
  })
  .catch(error => {
    throw new Error(error)
  })
}
