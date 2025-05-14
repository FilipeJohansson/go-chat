import { LoaderCircle } from "lucide-react"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router"
import { authenticate } from "../internal/lib/auth"
import { getTokens } from "../internal/lib/tokens"

export function Login() {
  const navigate = useNavigate()

  const [username, setUsername] = useState<string>('')
  const [password, setPassword] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (getTokens()) navigate("/lobby")
  }, [navigate])

  const handleLogin = (): void => {
    setError(undefined)

    if (!username || !password) {
      setError("Blank Username or Password")
      return
    }

    setLoading(true)
    authenticate(username, password)
      .then((resp: boolean) => {
        if (resp) {
          navigate("/lobby")
          return
        }
      })
      .catch((error) => setError(error))
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