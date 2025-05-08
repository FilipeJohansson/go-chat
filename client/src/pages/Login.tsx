import { LoaderCircle } from "lucide-react"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router"
import { clearTokens, getTokens, saveTokens } from "../internal/tokens"
import { LoginRequestMessage, Packet } from "../proto/packets"

export function Login() {
  const navigate = useNavigate()

  const [username, setUsername] = useState<string>('')
  const [password, setPassword] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (getTokens()) navigate("/chat")
  }, [])

  const handleLogin = (): void => {
    setError(undefined)

    if (!username || !password) {
      setError("Blank Username or Password")
      return
    }

    const loginReq: LoginRequestMessage = LoginRequestMessage.create({ username, password })
    const packet: Packet = Packet.create({ loginRequest: loginReq })
    sendLoginPacket(packet)
  }

  const sendLoginPacket = (packet: Packet): void => {
    setLoading(true)

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

      if (packet.denyResponse) setError(packet.denyResponse.reason)

      const accessToken: string | undefined = packet.jwt?.accessToken
      const refreshToken: string | undefined = packet.jwt?.refreshToken

      if (!accessToken || !refreshToken) {
        console.log("login: error getting access or refresh token")
        clearTokens()
        return
      }

      saveTokens({ accessToken, refreshToken })

      navigate("/chat")
    })
    .catch(error => console.error("Erro:", error))
    .finally(() => setLoading(false))
  }

  return (
    <div className="w-screen h-screen bg-[url(/src/assets/background.png)] flex items-center justify-center bg-no-repeat bg-cover bg-center">
      <div className="w-[400px] h-[350px] flex flex-col items-center justify-between p-4 gap-4 bg-white rounded-md">
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

        {error && <div className="text-red-500">
          <span>{error}</span>
        </div>}

        <button
          className="flex flex-row items-center justify-center w-full h-10 text-white bg-blue-500 rounded-xl hover:bg-pink-500 transition"
          onClick={handleLogin}
        >
         {loading
         ? <LoaderCircle className="animate-spin" />
         : <span className="font-semibold">LOGIN</span>}
        </button>
      </div>
    </div>
  )
}