package handler

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nikitanaumenko/calendar/internal/api/gen"
	"github.com/nikitanaumenko/calendar/internal/db"
)

func TestAvailableSlotsSkipsOccupiedSlot(t *testing.T) {
	windowStartsAt := time.Date(2026, 6, 1, 9, 0, 0, 0, time.UTC)
	windowEndsAt := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)
	duration := 30 * time.Minute
	bookings := []db.ListBookingsInWindowRow{
		{
			StartsAt: pgtype.Timestamptz{Time: time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC), Valid: true},
			EndsAt:   pgtype.Timestamptz{Time: time.Date(2026, 6, 1, 10, 30, 0, 0, time.UTC), Valid: true},
		},
	}

	slots := availableSlots(windowStartsAt, windowEndsAt, duration, bookings)

	for _, slot := range slots {
		if slot.StartsAt.Equal(bookings[0].StartsAt.Time) {
			t.Fatalf("occupied slot %s was returned", slot.StartsAt)
		}
	}

	if !containsSlot(slots, time.Date(2026, 6, 1, 10, 30, 0, 0, time.UTC)) {
		t.Fatal("adjacent slot after occupied booking should be available")
	}
}

func TestIsValidSlotStartRequiresGeneratedSlotBoundary(t *testing.T) {
	windowStartsAt := time.Date(2026, 6, 1, 9, 5, 0, 0, time.UTC)
	windowEndsAt := windowStartsAt.Add(bookingWindow)
	duration := 30 * time.Minute

	if !isValidSlotStart(time.Date(2026, 6, 1, 9, 30, 0, 0, time.UTC), duration, windowStartsAt, windowEndsAt) {
		t.Fatal("expected next generated slot boundary to be valid")
	}

	if isValidSlotStart(time.Date(2026, 6, 1, 9, 45, 0, 0, time.UTC), duration, windowStartsAt, windowEndsAt) {
		t.Fatal("expected non-generated slot boundary to be invalid")
	}
}

func containsSlot(slots []gen.AvailableSlot, startsAt time.Time) bool {
	for _, slot := range slots {
		if slot.StartsAt.Equal(startsAt) {
			return true
		}
	}
	return false
}
