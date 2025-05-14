import { LoaderCircle } from "lucide-react";
import { useEffect, useState } from "react";
import { useNavigate } from "react-router";
import { register } from "../internal/lib/auth";
import { getTokens } from "../internal/lib/tokens";

export function Register() {
  const navigate = useNavigate();

  const [username, setUsername] = useState<string>('')
  const [password, setPassword] = useState<string>('')
  const [repeatedPassword, setRepeatedPassword] = useState<string>('')
  const [loading, setLoading] = useState<boolean>(false)
  const [error, setError] = useState<string | undefined>(undefined)

  useEffect(() => {
    if (getTokens()) navigate("/chat")
  }, [navigate])

  const handleRegister = (): void => {
    setError(undefined)

    if (password !== repeatedPassword) {
      setError("Passwords must be identical")
      return
    }

    if (!username || !password) {
      setError("Blank Username or Password")
      return
    }

    //! validate username and password min reqs

    setLoading(true)
    register(username, password)
      .catch((error) => setError(error))
      .finally(() => setLoading(false))
  }

  return (
    <div className="w-screen h-screen bg-[url(/src/assets/background.png)] flex items-center justify-center bg-no-repeat bg-cover bg-center">
      <div className="w-[400px] h-[400px] flex flex-col items-center justify-between p-4 gap-4 bg-white rounded-md">
        <div>
          <span className="text-xl">Register</span>
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
          <div className="w-full flex flex-col gap-1">
            <span>Repeat Password</span>
            <input
              className="p-1 border-b border-b-2 focus:outline-pink-500"
              type="password"
              placeholder="Repeat your password"
              value={repeatedPassword}
              onChange={e => setRepeatedPassword(e.target.value)}
            />
          </div>
        </div>

        {error && <div className="text-red-500">
          <span>{error}</span>
        </div>}

        <button
          className="flex flex-row items-center justify-center w-full h-10 text-white bg-blue-500 rounded-xl hover:bg-pink-500 transition"
          onClick={handleRegister}
        >
         {loading
         ? <LoaderCircle className="animate-spin" />
         : <span className="font-semibold">REGISTER</span>}
        </button>
      </div>
    </div>
  )
}