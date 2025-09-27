// src/pages/EventDetail.tsx
import React, { useContext, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import type { Booking, Event } from "../api/api";
import { bookEvent, cancelBooking, confirmBooking, getEvent } from "../api/api";
import { AuthContext } from "../context/AuthContext";

const EventDetail: React.FC = () => {
  const { eventID } = useParams<{ eventID: string }>();
  const [event, setEvent] = useState<Event | null>(null);
  const [booking, setBooking] = useState<Booking | null>(null);
  const [error, setError] = useState("");
  const [message, setMessage] = useState("");
  const authContext = useContext(AuthContext);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchEvent = async () => {
      if (!eventID) return;
      try {
        const data = await getEvent(eventID);
        setEvent(data);
      } catch (err: any) {
        setError(err.response?.data?.error || "Failed to load event");
      }
    };
    fetchEvent();

    // Poll every 10 seconds to check for changes (e.g., expired bookings)
    const interval = setInterval(fetchEvent, 10000);
    return () => clearInterval(interval);
  }, [eventID]);

  const handleBook = async () => {
    if (!authContext?.isAuthenticated) {
      navigate("/login");
      return;
    }
    if (!eventID) return;
    try {
      const response = await bookEvent(eventID);
      setBooking({
        id: response.result.id,
        event_id: eventID,
        user_id: "", // You may need to fetch user_id separately
        status: "pending",
        expires_at: new Date(Date.now() + 10 * 60 * 1000).toISOString(), // Placeholder
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      });
      setMessage(response.result.message);
      // Refresh event
      const updatedEvent = await getEvent(eventID);
      setEvent(updatedEvent);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to book");
    }
  };

  const handleConfirm = async () => {
    if (!eventID || !booking?.id) return;
    try {
      const response = await confirmBooking(eventID, booking.id);
      setMessage(response.result.message);
      setBooking(null); // Or update status
      // Refresh event
      const updatedEvent = await getEvent(eventID);
      setEvent(updatedEvent);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to confirm");
    }
  };

  const handleCancel = async () => {
    if (!eventID || !booking?.id) return;
    try {
      const response = await cancelBooking(eventID, booking.id);
      setMessage(response.result.message);
      setBooking(null);
      // Refresh event
      const updatedEvent = await getEvent(eventID);
      setEvent(updatedEvent);
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to cancel");
    }
  };

  if (!event) return <p>Loading...</p>;

  return (
    <div className="max-w-md mx-auto bg-white p-8 rounded shadow">
      <h2 className="text-2xl mb-4">{event.title}</h2>
      {error && <p className="text-red-500">{error}</p>}
      {message && <p className="text-green-500">{message}</p>}
      <p>Date: {new Date(event.date).toLocaleString()}</p>
      <p>
        Available Seats: {event.available_seats} / {event.total_seats}
      </p>
      <p>Booked: {event.total_seats - event.available_seats}</p>
      <p>Booking TTL: {event.booking_ttl}</p>
      {!booking &&
        authContext?.isAuthenticated &&
        event.available_seats > 0 && (
          <button
            onClick={handleBook}
            className="bg-green-500 text-white p-2 rounded mt-4"
          >
            Book Seat
          </button>
        )}
      {booking && (
        <div className="mt-4">
          <p>Your Booking ID: {booking.id}</p>
          <p>Expires At: {new Date(booking.expires_at).toLocaleString()}</p>
          <button
            onClick={handleConfirm}
            className="bg-blue-500 text-white p-2 rounded mr-2"
          >
            Confirm
          </button>
          <button
            onClick={handleCancel}
            className="bg-red-500 text-white p-2 rounded"
          >
            Cancel
          </button>
        </div>
      )}
    </div>
  );
};

export default EventDetail;
