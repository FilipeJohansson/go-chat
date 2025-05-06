export function Register() {
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
          <div className="w-full flex flex-col gap-1">
            <span>Repeat Password</span>
            <input
              className="p-1 border-b border-b-2 focus:outline-pink-500"
              type="password"
              placeholder="Repeat your password"
            />
          </div>
        </div>

        <button className="w-full h-10 bg-blue-500 rounded-xl hover:bg-pink-500 transition">
          <span className="text-white font-semibold">REGISTER</span>
        </button>
      </div>
    </div>
  )
}