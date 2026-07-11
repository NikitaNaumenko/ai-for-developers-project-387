export type ApiErrorBody = {
  error: {
    code: string;
    message: string;
  };
};

export type EventType = {
  id: string;
  title: string;
  description: string;
  durationMinutes: number;
};

export type AvailableSlot = {
  startsAt: string;
  endsAt: string;
};

export type SlotsResponse = {
  eventTypeId: string;
  windowStartsAt: string;
  windowEndsAt: string;
  slots: AvailableSlot[];
};

export type Booking = {
  id: string;
  eventTypeId: string;
  eventTypeTitle: string;
  startsAt: string;
  endsAt: string;
  guestName: string;
  guestEmail: string;
  guestNote?: string;
  createdAt: string;
};

export type CreateBookingRequest = {
  eventTypeId: string;
  startsAt: string;
  guestName: string;
  guestEmail: string;
  guestNote?: string;
};

export type CreateEventTypeRequest = {
  id: string;
  title: string;
  description: string;
  durationMinutes: number;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = 'ApiError';
    this.status = status;
    this.code = code;
  }
}

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL ?? '/api';

export const calendarApi = {
  listEventTypes: () => request<EventType[]>('/event-types'),
  listSlots: (eventTypeId: string) =>
    request<SlotsResponse>(`/event-types/${encodeURIComponent(eventTypeId)}/slots`),
  createBooking: (payload: CreateBookingRequest) =>
    request<Booking>('/bookings', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  createEventType: (payload: CreateEventTypeRequest) =>
    request<EventType>('/admin/event-types', {
      method: 'POST',
      body: JSON.stringify(payload),
    }),
  listUpcomingBookings: () => request<Booking[]>('/admin/bookings/upcoming'),
};

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_BASE_URL}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
      ...init?.headers,
    },
  });

  if (!response.ok) {
    const body = await readError(response);
    throw new ApiError(response.status, body.error.code, body.error.message);
  }

  return response.json() as Promise<T>;
}

async function readError(response: Response): Promise<ApiErrorBody> {
  try {
    return (await response.json()) as ApiErrorBody;
  } catch {
    return {
      error: {
        code: 'request_failed',
        message: `Request failed with status ${response.status}`,
      },
    };
  }
}
