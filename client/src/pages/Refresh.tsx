import { useEffect } from "react";
import { useNavigate } from "react-router";
import { clearTokens, getTokens, saveTokens } from "../internal/lib/tokens";
import { Message, RefreshRequestMessage } from "../proto/packets";

export function Refresh() {
  const navigate = useNavigate();

  useEffect(() => {
    handleRefresh()
  }, [])

  const handleRefresh = (): void => {
    const refreshToken: string | undefined = getTokens()?.refreshToken
    if (!refreshToken) {
      console.log("Refresh token not set")
      navigate("/login")
      return
    }

    clearTokens()

    const refreshReq: RefreshRequestMessage = RefreshRequestMessage.create()
    const message: Message = Message.create({ refresh: refreshReq })
    sendRefreshPacket(message, refreshToken)
  }

  const sendRefreshPacket = (message: Message, refreshToken: string): void => {
    const binary: Uint8Array = Message.encode(message).finish()
    fetch("http://localhost:8080/refresh", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream",
        "Authorization": refreshToken,
      },
      body: binary,
      mode: "cors"
    })
    .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
    .then((buffer: ArrayBuffer) => {
      const data: Uint8Array = new Uint8Array(buffer)
      const message: Message = Message.decode(data)
      if (!message.jwt) {
        console.log("Message not type JWT message", message)
        navigate("/login")
        return
      }

      const accessToken: string | undefined = message.jwt.accessToken
      const refreshToken: string | undefined = message.jwt.refreshToken

      if (!accessToken || !refreshToken) {
        console.log("refresh: error getting access or refresh token")
        clearTokens()
        navigate("/login")
        return
      }

      saveTokens({ accessToken, refreshToken })
      navigate("/chat")     
    })
    .catch(error => console.error("Erro:", error));
  }

  return (
    <></>
  )
}