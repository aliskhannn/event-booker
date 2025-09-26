// src/components/Header.tsx
import React, { useContext } from "react";
import { Link } from "react-router-dom";
import { AuthContext } from "../context/AuthContext";

const Header: React.FC = () => {
  const authContext = useContext(AuthContext);

  if (!authContext) {
    return null;
  }

  const { isAuthenticated, logout } = authContext;

  return (
    <header className="bg-blue-600 text-white p-4">
      <nav className="container mx-auto flex justify-between items-center">
        <Link to="/" className="text-xl font-bold">
          EventBooker
        </Link>
        <div>
          {isAuthenticated ? (
            <>
              <Link to="/create-event" className="mr-4">
                Create Event
              </Link>
              <button onClick={logout} className="bg-red-500 px-4 py-2 rounded">
                Logout
              </button>
            </>
          ) : (
            <>
              <Link to="/login" className="mr-4">
                Login
              </Link>
              <Link to="/register">Register</Link>
            </>
          )}
        </div>
      </nav>
    </header>
  );
};

export default Header;
