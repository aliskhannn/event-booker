// src/api.ts
import axios from "axios";

const API_BASE_URL = "http://localhost:8080/api";

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
  date: string; // ISO string or backend format
  total_seats: number;
  available_seats: number;
  booking_ttl: string; // e.g., "10m"
}

export interface Event {
  id: string; // uuid as string
  title: string;
  date: string; // time.Time as string
  total_seats: number;
  available_seats: number;
  booking_ttl: string; // time.Duration as string
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

// Assuming backend returns { token: string } on login/register
export interface AuthResponse {
  token: string;
}

// Assuming book returns the booking
export interface BookResponse {
  booking: Booking;
}

// Assuming confirm/cancel return success or something
export interface ActionResponse {
  message: string;
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
// Note: Assuming GET /events exists for listing, as required by TZ for viewing list.
// If not, backend needs to add it. For now, implement as if it does, returning Event[]
export const getEvents = async (): Promise<Event[]> => {
  const response = await api.get("/events");
  return response.data;
};

export const getEvent = async (eventID: string): Promise<Event> => {
  const response = await api.get(`/events/${eventID}`);
  return response.data;
};

export const createEvent = async (data: CreateEventRequest): Promise<Event> => {
  const response = await api.post("/events", data);
  return response.data;
};

export const bookEvent = async (eventID: string): Promise<BookResponse> => {
  const response = await api.post(`/events/${eventID}/book`);
  return response.data;
};

export const confirmBooking = async (
  eventID: string,
  bookingID: string
): Promise<ActionResponse> => {
  const response = await api.post(
    `/events/${eventID}/booking/${bookingID}/confirm`
  );
  return response.data;
};

export const cancelBooking = async (
  eventID: string,
  bookingID: string
): Promise<ActionResponse> => {
  const response = await api.post(
    `/events/${eventID}/booking/${bookingID}/cancel`
  );
  return response.data;
};
