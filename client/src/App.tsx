import { Route, BrowserRouter as Router, Routes } from "react-router";
import { Chat } from "./pages/Chat";
import { Login } from "./pages/Login";
import { Register } from "./pages/Register";

export function App() {
  return (
    <>
      <Router>
        <Routes>
          <Route index path="/chat" element={<Chat />} />
          <Route index path="/login" element={<Login />} />
          <Route index path="/register" element={<Register />} />
        </Routes>
      </Router>
    </>
  )
}
