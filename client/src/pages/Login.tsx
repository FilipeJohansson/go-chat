import { LoginRequestMessage, Packet } from "../proto/packets"

export function Login() {

  const t = () => {
    const loginReq: LoginRequestMessage = LoginRequestMessage.create({
      username: 'filipe', password: '123'
    })
    // const registerReq: RegisterRequestMessage = RegisterRequestMessage.create({
    //   username: 'filipe', password: '123'
    // })
    const packet: Packet = Packet.create({ senderId: 0, loginRequest: loginReq })
    // const packet: Packet = Packet.create({ senderId: 0, registerRequest: registerReq })
    send(packet)
  }

  const send = (packet: Packet): void => {
    packet.senderId = 0
    const binary: Uint8Array = Packet.encode(packet).finish()
    fetch("http://localhost:8080/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/octet-stream"
      },
      body: binary,
      mode: "cors"
    })
    .then(response => response.arrayBuffer())
    .then(buffer => {
      const data = new Uint8Array(buffer)
      const packet = Packet.decode(data)
      console.log("Resposta do servidor:", packet)
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
            />
          </div>

          <div className="w-full flex flex-col gap-1">
            <span>Password</span>
            <input
              className="p-1 border-b border-b-2 focus:outline-pink-500"
              type="password"
              placeholder="Type your password"
            />
          </div>
        </div>

        <button
          className="w-full h-10 bg-blue-500 rounded-xl hover:bg-pink-500 transition"
          onClick={t}
        >
          <span className="text-white font-semibold">LOGIN</span>
        </button>
      </div>
    </div>
  )
}