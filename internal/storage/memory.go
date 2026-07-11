package storage

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nikitanaumenko/calendar/internal/db"
)

type MemoryStore struct {
	mu         sync.RWMutex
	eventTypes map[uuid.UUID]db.EventType
	bookings   map[uuid.UUID]db.CreateBookingRow
	now        func() time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		eventTypes: make(map[uuid.UUID]db.EventType),
		bookings:   make(map[uuid.UUID]db.CreateBookingRow),
		now:        time.Now,
	}
}

func (s *MemoryStore) CreateEventType(_ context.Context, arg db.CreateEventTypeParams) (db.EventType, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuidFromPg(arg.ID)
	if _, exists := s.eventTypes[id]; exists {
		return db.EventType{}, pgError("23505")
	}

	now := timestamptz(s.now())
	eventType := db.EventType{
		ID:              arg.ID,
		Title:           arg.Title,
		Description:     arg.Description,
		DurationMinutes: arg.DurationMinutes,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	s.eventTypes[id] = eventType
	return eventType, nil
}

func (s *MemoryStore) GetEventType(_ context.Context, id pgtype.UUID) (db.EventType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventType, ok := s.eventTypes[uuidFromPg(id)]
	if !ok {
		return db.EventType{}, pgx.ErrNoRows
	}
	return eventType, nil
}

func (s *MemoryStore) ListEventTypes(context.Context) ([]db.EventType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	eventTypes := make([]db.EventType, 0, len(s.eventTypes))
	for _, eventType := range s.eventTypes {
		eventTypes = append(eventTypes, eventType)
	}
	sort.Slice(eventTypes, func(i, j int) bool {
		if eventTypes[i].Title == eventTypes[j].Title {
			return uuidFromPg(eventTypes[i].ID).String() < uuidFromPg(eventTypes[j].ID).String()
		}
		return eventTypes[i].Title < eventTypes[j].Title
	})
	return eventTypes, nil
}

func (s *MemoryStore) CreateBooking(_ context.Context, arg db.CreateBookingParams) (db.CreateBookingRow, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	eventType, ok := s.eventTypes[uuidFromPg(arg.EventTypeID)]
	if !ok {
		return db.CreateBookingRow{}, pgError("23503")
	}

	for _, booking := range s.bookings {
		if overlaps(arg.StartsAt.Time, arg.EndsAt.Time, booking.StartsAt.Time, booking.EndsAt.Time) {
			return db.CreateBookingRow{}, pgError("23P01")
		}
	}

	id := uuid.New()
	booking := db.CreateBookingRow{
		ID:             uuidParam(id),
		EventTypeID:    arg.EventTypeID,
		EventTypeTitle: eventType.Title,
		StartsAt:       timestamptz(arg.StartsAt.Time),
		EndsAt:         timestamptz(arg.EndsAt.Time),
		GuestName:      arg.GuestName,
		GuestEmail:     arg.GuestEmail,
		GuestNote:      arg.GuestNote,
		CreatedAt:      timestamptz(s.now()),
	}
	s.bookings[id] = booking
	return booking, nil
}

func (s *MemoryStore) ListBookingsInWindow(_ context.Context, arg db.ListBookingsInWindowParams) ([]db.ListBookingsInWindowRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookings := make([]db.ListBookingsInWindowRow, 0)
	for _, booking := range s.bookings {
		if !overlaps(booking.StartsAt.Time, booking.EndsAt.Time, arg.WindowStartsAt.Time, arg.WindowEndsAt.Time) {
			continue
		}
		bookings = append(bookings, bookingWindowRow(booking))
	}
	sortBookingWindowRows(bookings)
	return bookings, nil
}

func (s *MemoryStore) ListUpcomingBookings(_ context.Context, startsAt pgtype.Timestamptz) ([]db.ListUpcomingBookingsRow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	bookings := make([]db.ListUpcomingBookingsRow, 0)
	for _, booking := range s.bookings {
		if booking.StartsAt.Time.Before(startsAt.Time) {
			continue
		}
		bookings = append(bookings, upcomingBookingRow(booking))
	}
	sortUpcomingBookingRows(bookings)
	return bookings, nil
}

func bookingWindowRow(booking db.CreateBookingRow) db.ListBookingsInWindowRow {
	return db.ListBookingsInWindowRow{
		ID:             booking.ID,
		EventTypeID:    booking.EventTypeID,
		EventTypeTitle: booking.EventTypeTitle,
		StartsAt:       booking.StartsAt,
		EndsAt:         booking.EndsAt,
		GuestName:      booking.GuestName,
		GuestEmail:     booking.GuestEmail,
		GuestNote:      booking.GuestNote,
		CreatedAt:      booking.CreatedAt,
	}
}

func upcomingBookingRow(booking db.CreateBookingRow) db.ListUpcomingBookingsRow {
	return db.ListUpcomingBookingsRow{
		ID:             booking.ID,
		EventTypeID:    booking.EventTypeID,
		EventTypeTitle: booking.EventTypeTitle,
		StartsAt:       booking.StartsAt,
		EndsAt:         booking.EndsAt,
		GuestName:      booking.GuestName,
		GuestEmail:     booking.GuestEmail,
		GuestNote:      booking.GuestNote,
		CreatedAt:      booking.CreatedAt,
	}
}

func sortBookingWindowRows(bookings []db.ListBookingsInWindowRow) {
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].StartsAt.Time.Equal(bookings[j].StartsAt.Time) {
			return uuidFromPg(bookings[i].ID).String() < uuidFromPg(bookings[j].ID).String()
		}
		return bookings[i].StartsAt.Time.Before(bookings[j].StartsAt.Time)
	})
}

func sortUpcomingBookingRows(bookings []db.ListUpcomingBookingsRow) {
	sort.Slice(bookings, func(i, j int) bool {
		if bookings[i].StartsAt.Time.Equal(bookings[j].StartsAt.Time) {
			return uuidFromPg(bookings[i].ID).String() < uuidFromPg(bookings[j].ID).String()
		}
		return bookings[i].StartsAt.Time.Before(bookings[j].StartsAt.Time)
	})
}

func overlaps(aStartsAt time.Time, aEndsAt time.Time, bStartsAt time.Time, bEndsAt time.Time) bool {
	return aStartsAt.Before(bEndsAt) && aEndsAt.After(bStartsAt)
}

func uuidParam(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPg(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}

func timestamptz(value time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: value.UTC(), Valid: true}
}

func pgError(code string) error {
	return &pgconn.PgError{Code: code}
}
