import { Route, BrowserRouter as Router, Routes } from "react-router";
import { Chat } from "./pages/Chat";
import { Login } from "./pages/Login";
import { Refresh } from "./pages/Refresh";
import { Register } from "./pages/Register";
import { Rooms } from "./pages/rooms";

export function App() {
  return (
    <>
      <Router>
        <Routes>
          <Route index path="/chat" element={<Chat />} />
          <Route index path="/login" element={<Login />} />
          <Route index path="/refresh" element={<Refresh />} />
          <Route index path="/register" element={<Register />} />
          <Route index path="/rooms" element={<Rooms />} />
        </Routes>
      </Router>
    </>
  )
}
