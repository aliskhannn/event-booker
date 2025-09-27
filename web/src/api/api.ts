// src/api.ts
import axios from "axios";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080/api";

const api = axios.create({
  baseURL: API_BASE_URL,
});

// Interceptor to add JWT token to requests
api.interceptors.request.use((config) => {
  const token = localStorage.getItem("token");
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

export interface RegisterRequest {
  email: string;
  password: string;
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface CreateEventRequest {
  title: string;
  date: string; // ISO string
  total_seats: number;
  available_seats: number;
  booking_ttl: string; // e.g., "10m"
}

export interface Event {
  id: string; // uuid as string
  title: string;
  date: string;
  total_seats: number;
  available_seats: number;
  booking_ttl: string;
  created_at: string;
  updated_at: string;
}

export interface Booking {
  id: string;
  event_id: string;
  user_id: string;
  status: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
}

// Backend wraps responses in { result: {...} }
export interface AuthResponse {
  result: {
    token: string;
  };
}

export interface BookResponse {
  result: {
    id: string;
    message: string;
  };
}

export interface ActionResponse {
  result: {
    message: string;
  };
}

export interface EventsResponse {
  result: {
    events: Event[];
  };
}

export interface EventResponse {
  result: {
    event: Event;
  };
}

export interface CreateEventResponse {
  result: {
    id: string;
  };
}

// Auth
export const register = async (
  data: RegisterRequest
): Promise<AuthResponse> => {
  const response = await api.post("/auth/register", data);
  return response.data;
};

export const login = async (data: LoginRequest): Promise<AuthResponse> => {
  const response = await api.post("/auth/login", data);
  return response.data;
};

// Events
export const getEvents = async (): Promise<Event[]> => {
  const response = await api.get<EventsResponse>("/events");
  return response.data.result.events || [];
};

export const getEvent = async (eventID: string): Promise<Event> => {
  const response = await api.get<EventResponse>(`/events/${eventID}`);
  return response.data.result.event;
};

export const createEvent = async (
  data: CreateEventRequest
): Promise<{ id: string }> => {
  const response = await api.post<CreateEventResponse>("/events", data);
  return response.data.result;
};

export const bookEvent = async (eventID: string): Promise<BookResponse> => {
  const response = await api.post<BookResponse>(`/events/${eventID}/book`);
  return response.data;
};

export const confirmBooking = async (
  eventID: string,
  bookingID: string
): Promise<ActionResponse> => {
  const response = await api.post<ActionResponse>(
    `/events/${eventID}/booking/${bookingID}/confirm`
  );
  return response.data;
};

export const cancelBooking = async (
  eventID: string,
  bookingID: string
): Promise<ActionResponse> => {
  const response = await api.post<ActionResponse>(
    `/events/${eventID}/booking/${bookingID}/cancel`
  );
  return response.data;
};
