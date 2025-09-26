// src/App.tsx
import React from "react";
import { Route, BrowserRouter as Router, Routes } from "react-router-dom";
import Header from "./components/Header";
import { AuthProvider } from "./context/AuthContext";
import CreateEvent from "./pages/CreateEvent";
import EventDetail from "./pages/EventDetail";
import EventList from "./pages/EventList";
import Login from "./pages/Login";
import Register from "./pages/Register";

const App: React.FC = () => {
  return (
    <AuthProvider>
      <Router>
        <div className="min-h-screen bg-gray-100">
          <Header />
          <main className="container mx-auto p-4">
            <Routes>
              <Route path="/" element={<EventList />} />
              <Route path="/login" element={<Login />} />
              <Route path="/register" element={<Register />} />
              <Route path="/create-event" element={<CreateEvent />} />
              <Route path="/events/:eventID" element={<EventDetail />} />
            </Routes>
          </main>
        </div>
      </Router>
    </AuthProvider>
  );
};

export default App;
