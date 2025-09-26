export interface User {
  id: string;
  email: string;
  name: string;
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

export interface Event {
  id: string;
  title: string;
  date: string;
  total_seats: number;
  available_seats: number;
  booking_ttl: string;
  created_at: string;
  updated_at: string;
}
