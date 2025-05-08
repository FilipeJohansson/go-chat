import { useEffect } from "react";
import { useNavigate } from "react-router";
import { getTokens, saveTokens } from "../internal/tokens";
import { Packet, RefreshRequestMessage } from "../proto/packets";

export function Refresh() {
  const navigate = useNavigate();

  useEffect(() => {
    if (getTokens()) navigate("/chat")
    handleRefresh()
  }, [])

  const handleRefresh = (): void => {
    const refreshToken: string | undefined = getTokens()?.refreshToken
    if (!refreshToken) {
      console.log("Refresh token not set")
      navigate("/login")
      return
    }

    const refreshReq: RefreshRequestMessage = RefreshRequestMessage.create({ refreshToken })
    const packet: Packet = Packet.create({ refreshRequest: refreshReq })
    sendRefreshPacket(packet)
  }

  const sendRefreshPacket = (packet: Packet): void => {
    const binary: Uint8Array = Packet.encode(packet).finish()
    fetch("http://localhost:8080/refresh", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream"
      },
      body: binary,
      mode: "cors"
    })
    .then((response: Response): Promise<ArrayBuffer> => response.arrayBuffer())
    .then((buffer: ArrayBuffer) => {
      const data: Uint8Array = new Uint8Array(buffer)
      const packet: Packet = Packet.decode(data)
      const accessToken: string | undefined = packet.jwt?.accessToken
      const refreshToken: string | undefined = packet.jwt?.refreshToken

      if (!accessToken || !refreshToken) {
        console.log("error getting access or refresh token")
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