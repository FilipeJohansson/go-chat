import { LoaderCircle } from "lucide-react"
import { useEffect, useState } from "react"
import { useNavigate } from "react-router"
import { isAuthenticated, logout } from "../internal/lib/auth"
import { createRoom, getAllRooms, Room } from "../internal/lib/rooms"

export function Lobby() {
  const navigate = useNavigate()
  const [rooms, setRooms] = useState<Room[]>([])
  const [roomName, setRoomName] = useState<string | undefined>(undefined)
  const [error, setError] = useState("")
  const [loadingRooms, setLoadingRooms] = useState<boolean>(false)

  // Fetch rooms
  const fetchRooms = () => {
    setLoadingRooms(true)
    getAllRooms()
      .then((rooms: Room[]) => setRooms(rooms))
      .catch((error) => setError(error))
      .finally(() => setLoadingRooms(false))
  }

  useEffect(() => {
    if (!isAuthenticated()) {
      navigate('/login')
      return
    }

    fetchRooms()

    // Poll for new rooms every 5 seconds
    const interval = setInterval(fetchRooms, 5000)

    return () => clearInterval(interval)
  }, [navigate])

  const handleLogout = () => {
    logout()
    navigate('/login')
  }

  const joinRoom = (roomId: number) => {
    navigate(`/chat?id=${roomId}`)
  }

  const handleCreateRoom = async (e: React.FormEvent) => {
    e.preventDefault()
    setError("")

    try {
      if (!roomName || !roomName.trim()) throw new Error("Please enter a room name")

      createRoom(roomName)
      .then((result) => {
        if (result) {
          fetchRooms()
          setRoomName('')
          return
        }

        setError("Failed to create room")
      })
    } catch (err) {
      setError(err instanceof Error ? err.message : "Failed to create room")
    }
  }

  return (
    <div className="min-h-screen p-4">
      <div className="max-w-4xl mx-auto">
        <div className="flex justify-between items-center mb-6 border-b pb-4">
          <h1 className="text-xl font-bold">Chat Rooms</h1>
          <div className="flex items-center gap-4">
            {/* <span className="text-sm">
              Logged in as <span className="text-blue-800">Filipe</span>
            </span> */}
            <button
              className="px-3 py-1 bg-blue-600 hover:bg-red-600 text-white text-sm rounded transition"
              onClick={handleLogout}
            >
              Sign out
            </button>
          </div>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          <div className="md:col-span-2">
            <div className="border rounded p-4">
              <h2 className="font-bold mb-4 border-b pb-2">Available Rooms ({rooms.length})</h2>
              {loadingRooms
              ? <LoaderCircle className="animate-spin" />
              : rooms.length === 0 ? (
                <p className="text-center py-4">
                  No rooms available. Create one to get started!
                </p>
              ) : (
                <div className="space-y-2 max-h-[400px] overflow-y-auto">
                  {rooms.map((room: Room) => (
                    <div
                      key={room.roomId}
                      className="flex items-center justify-between p-2 border rounded"
                    >
                      <div>
                        {/* <span className="text-gray-500 text-xs">{room.roomId}</span> */}
                        <h3 className="text-blue-500 font-semibold">{room.name}</h3>
                      </div>
                      <button
                        onClick={() => joinRoom(room.roomId)}
                        className="px-3 py-1 bg-blue-500 hover:bg-pink-500 text-white rounded transition"
                      >
                        Join
                      </button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </div>

          <div>
            <div className="border rounded p-4">
              <h2 className="font-bold mb-4 border-b pb-2">Create Room</h2>
              <form onSubmit={handleCreateRoom} className="space-y-4">
                <div>
                  <label className="block mb-1">Room Name</label>
                  <input
                    type="text"
                    value={roomName}
                    onChange={(e) => setRoomName(e.target.value)}
                    placeholder="Enter room name"
                    className="w-full p-2 border rounded"
                  />
                </div>

                {error && <p className="text-red-400 text-sm">{error}</p>}

                <button type="submit" className="w-full py-2 bg-blue-500 hover:bg-pink-500 text-white rounded">
                  Create Room
                </button>
              </form>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
