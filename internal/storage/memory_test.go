package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nikitanaumenko/calendar/internal/db"
)

func TestMemoryStoreCreatesBookingAndRejectsOverlap(t *testing.T) {
	ctx := context.Background()
	store := NewMemoryStore()
	eventTypeID := uuid.New()

	eventType, err := store.CreateEventType(ctx, db.CreateEventTypeParams{
		ID:              uuidParam(eventTypeID),
		Title:           "Intro",
		Description:     "Intro call",
		DurationMinutes: 30,
	})
	if err != nil {
		t.Fatal(err)
	}

	startsAt := time.Date(2026, 6, 3, 10, 0, 0, 0, time.UTC)
	booking, err := store.CreateBooking(ctx, db.CreateBookingParams{
		EventTypeID: uuidParam(eventTypeID),
		StartsAt:    timestamptz(startsAt),
		EndsAt:      timestamptz(startsAt.Add(30 * time.Minute)),
		GuestName:   "Nikita",
		GuestEmail:  "nikita@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	if booking.EventTypeTitle != eventType.Title {
		t.Fatalf("expected event type title %q, got %q", eventType.Title, booking.EventTypeTitle)
	}

	_, err = store.CreateBooking(ctx, db.CreateBookingParams{
		EventTypeID: uuidParam(eventTypeID),
		StartsAt:    timestamptz(startsAt.Add(15 * time.Minute)),
		EndsAt:      timestamptz(startsAt.Add(45 * time.Minute)),
		GuestName:   "Guest",
		GuestEmail:  "guest@example.com",
	})
	if pgErrorCode(err) != "23P01" {
		t.Fatalf("expected overlap error 23P01, got %v", err)
	}
}

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
