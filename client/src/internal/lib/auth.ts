import { LoginRequestMessage, LogoutRequestMessage, Message, RegisterRequestMessage } from "../../proto/packets"
import { clearTokens, getTokens, saveTokens } from "./tokens"

export interface User {
  id: number,
  name: string,
}

export function isAuthenticated(): boolean {
  return getTokens() !== undefined
}

export async function authenticate(username: string, password: string): Promise<boolean> {
  const loginReq: LoginRequestMessage = LoginRequestMessage.create({ username, password })
  const msg: Message = Message.create({ login: loginReq })
  const binary: Uint8Array = Message.encode(msg).finish()
  
  return fetch("http://localhost:8080/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream"
    },
    body: binary,
    mode: "cors"
  })
  .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
  .then((buffer: ArrayBuffer): boolean => {
    const data = new Uint8Array(buffer)
    const message: Message = Message.decode(data)
    console.log("Server Response:", message)
    console.log(message)

    if (message.denyResponse) throw new Error(message.denyResponse.reason)
    if (!message.jwt) throw new Error("Wrong message type")

    const accessToken: string | undefined = message.jwt.accessToken
    const refreshToken: string | undefined = message.jwt.refreshToken

    if (!accessToken || !refreshToken) {
      clearTokens()
      throw new Error("Error getting access or refresh token")
    }

    saveTokens({ accessToken, refreshToken })
    return true
  })
  .catch(error => {
    throw new Error(error)
  })
}

export async function register(username: string, password: string): Promise<boolean> {
  const registerReq: RegisterRequestMessage = RegisterRequestMessage.create({ username, password })
  const msg: Message = Message.create({ register: registerReq })
  const binary: Uint8Array = Message.encode(msg).finish()
  
  return fetch("http://localhost:8080/login", {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream"
    },
    body: binary,
    mode: "cors"
  })
  .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
  .then((buffer: ArrayBuffer): boolean => {
    const data = new Uint8Array(buffer)
    const message: Message = Message.decode(data)
    if (message.denyResponse) throw new Error(message.denyResponse.reason)
    if (message.okResponse) return true
    return false
  })
  .catch(error => {
    throw new Error(error)
  })
}

export function logout(): void {
  const refreshToken: string = getTokens()!.refreshToken
  const logoutRequest: LogoutRequestMessage = LogoutRequestMessage.create()
  const msg: Message = Message.create<Message>({ logout: logoutRequest })
  const binary: Uint8Array = Message.encode(msg).finish()
  fetch("http://localhost:8080/logout", {
    method: "POST",
    headers: {
      "Content-Type": "application/octet-stream",
      "Authorization": refreshToken,
    },
    body: binary,
    mode: "cors"
  })
  .catch(error => console.error("Error on logout:", error))
  clearTokens()
}
