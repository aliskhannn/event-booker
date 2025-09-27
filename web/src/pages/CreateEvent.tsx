// src/pages/CreateEvent.tsx
import React, { useContext, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createEvent } from "../api/api";
import { AuthContext } from "../context/AuthContext";

const CreateEvent: React.FC = () => {
  const [title, setTitle] = useState("");
  const [date, setDate] = useState("");
  const [totalSeats, setTotalSeats] = useState(0);
  const [availableSeats, setAvailableSeats] = useState(0);
  const [bookingTTL, setBookingTTL] = useState("10m");
  const [error, setError] = useState("");
  const navigate = useNavigate();
  const authContext = useContext(AuthContext);

  if (!authContext?.isAuthenticated) {
    navigate("/login");
    return null;
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      // Format date to full RFC3339 (e.g., "2025-09-27T10:27:00Z")
      let formattedDate = "";
      if (date) {
        const dateObj = new Date(date);
        dateObj.setSeconds(0, 0); // Ensure seconds are set
        formattedDate = dateObj.toISOString();
      }
      const event = await createEvent({
        title,
        date: formattedDate,
        total_seats: totalSeats,
        available_seats: availableSeats,
        booking_ttl: bookingTTL,
      });
      navigate(`/events/${event.id}`);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to create event");
    }
  };

  return (
    <div className="max-w-md mx-auto bg-white p-8 rounded shadow">
      <h2 className="text-2xl mb-4">Create Event</h2>
      {error && <p className="text-red-500">{error}</p>}
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          placeholder="Title"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          className="w-full mb-4 p-2 border rounded"
        />
        <input
          type="datetime-local"
          value={date}
          onChange={(e) => setDate(e.target.value)}
          className="w-full mb-4 p-2 border rounded"
        />
        <input
          type="number"
          placeholder="Total Seats"
          value={totalSeats}
          onChange={(e) => setTotalSeats(parseInt(e.target.value) || 0)}
          className="w-full mb-4 p-2 border rounded"
        />
        <input
          type="number"
          placeholder="Available Seats"
          value={availableSeats}
          onChange={(e) => setAvailableSeats(parseInt(e.target.value) || 0)}
          className="w-full mb-4 p-2 border rounded"
        />
        <input
          type="text"
          placeholder="Booking TTL (e.g., 10m)"
          value={bookingTTL}
          onChange={(e) => setBookingTTL(e.target.value)}
          className="w-full mb-4 p-2 border rounded"
        />
        <button
          type="submit"
          className="w-full bg-blue-500 text-white p-2 rounded"
        >
          Create
        </button>
      </form>
    </div>
  );
};

export default CreateEvent;
