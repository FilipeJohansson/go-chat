import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { getTokens, saveTokens } from "../internal/tokens";
import { LoginRequestMessage, Packet } from "../proto/packets";

export function Login() {
  const navigate = useNavigate();

  const [username, setUsername] = useState<string>('')
  const [password, setPassword] = useState<string>('')

  useEffect(() => {
    if (getTokens()) navigate("/chat")
  }, [])

  const handleLogin = (): void => {
    const loginReq: LoginRequestMessage = LoginRequestMessage.create({ username, password })
    const packet: Packet = Packet.create({ loginRequest: loginReq })
    sendLoginPacket(packet)
  }

  const sendLoginPacket = (packet: Packet): void => {
    const binary: Uint8Array = Packet.encode(packet).finish()
    fetch("http://localhost:8080/login", {
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
      const accessToken: string = packet.jwt?.accessToken || ''
      const refreshToken: string = packet.jwt?.refreshToken || ''
      saveTokens({ accessToken, refreshToken })

      navigate("/chat")     
    })
    .catch(error => console.error("Erro:", error));
  }

  return (
    <div className="w-screen h-screen bg-[url(/src/assets/background.png)] flex items-center justify-center bg-no-repeat bg-cover bg-center">
      <div className="w-[400px] h-[300px] flex flex-col items-center justify-between p-4 gap-4 bg-white rounded-md">
        <div>
          <span className="text-xl">Login</span>
        </div>

        <div className="w-full flex flex-col gap-4">
          <div className="w-full flex flex-col gap-1">
            <span>Username</span>
            <input
              className="p-1 border-b border-b-2 focus:outline-pink-500"
              type="text"
              placeholder="Type your username"
              value={username}
              onChange={e => setUsername(e.target.value)}
            />
          </div>

          <div className="w-full flex flex-col gap-1">
            <span>Password</span>
            <input
              className="p-1 border-b border-b-2 focus:outline-pink-500"
              type="password"
              placeholder="Type your password"
              value={password}
              onChange={e => setPassword(e.target.value)}
            />
          </div>
        </div>

        <button
          className="w-full h-10 bg-blue-500 rounded-xl hover:bg-pink-500 transition"
          onClick={handleLogin}
        >
          <span className="text-white font-semibold">LOGIN</span>
        </button>
      </div>
    </div>
  )
}