// src/pages/EventList.tsx
import React, { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import type { Event } from "../api/api";
import { getEvents } from "../api/api";

const EventList: React.FC = () => {
  const [events, setEvents] = useState<Event[]>([]);
  const [error, setError] = useState("");

  useEffect(() => {
    const fetchEvents = async () => {
      try {
        const data = await getEvents();
        setEvents(data);
      } catch (err) {
        setError("Failed to load events");
      }
    };
    fetchEvents();
  }, []);

  return (
    <div>
      <h2 className="text-2xl mb-4">Events List</h2>
      {error && <p className="text-red-500">{error}</p>}
      <ul>
        {events.map((event) => (
          <li key={event.id} className="mb-4 p-4 bg-white rounded shadow">
            <Link to={`/events/${event.id}`} className="text-blue-500">
              {event.title} - {new Date(event.date).toLocaleString()}
            </Link>
            <p>
              Available Seats: {event.available_seats} / {event.total_seats}
            </p>
            <p>Booked: {event.total_seats - event.available_seats}</p>
          </li>
        ))}
      </ul>
    </div>
  );
};

export default EventList;
