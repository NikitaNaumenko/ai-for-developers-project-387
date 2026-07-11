package handler

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/nikitanaumenko/calendar/internal/api/gen"
	"github.com/nikitanaumenko/calendar/internal/db"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const bookingWindow = 14 * 24 * time.Hour

type Handler struct {
	store Store
	now   func() time.Time
}

type Store interface {
	CreateBooking(context.Context, db.CreateBookingParams) (db.CreateBookingRow, error)
	CreateEventType(context.Context, db.CreateEventTypeParams) (db.EventType, error)
	GetEventType(context.Context, pgtype.UUID) (db.EventType, error)
	ListBookingsInWindow(context.Context, db.ListBookingsInWindowParams) ([]db.ListBookingsInWindowRow, error)
	ListEventTypes(context.Context) ([]db.EventType, error)
	ListUpcomingBookings(context.Context, pgtype.Timestamptz) ([]db.ListUpcomingBookingsRow, error)
}

func New(store Store) *Handler {
	return &Handler{
		store: store,
		now:   time.Now,
	}
}

func (h *Handler) AdminBookingsListUpcomingBookings(
	ctx context.Context,
	_ gen.AdminBookingsListUpcomingBookingsRequestObject,
) (gen.AdminBookingsListUpcomingBookingsResponseObject, error) {
	rows, err := h.store.ListUpcomingBookings(ctx, timestamptz(h.now().UTC()))
	if err != nil {
		return nil, err
	}

	bookings := make([]gen.Booking, 0, len(rows))
	for _, row := range rows {
		bookings = append(bookings, bookingFromUpcomingRow(row))
	}

	return gen.AdminBookingsListUpcomingBookings200JSONResponse(bookings), nil
}

func (h *Handler) AdminEventTypesCreateEventType(
	ctx context.Context,
	request gen.AdminEventTypesCreateEventTypeRequestObject,
) (gen.AdminEventTypesCreateEventTypeResponseObject, error) {
	if request.Body == nil {
		return gen.AdminEventTypesCreateEventType400JSONResponse(errorResponse("bad_request", "request body is required")), nil
	}

	body := *request.Body
	if err := validateEventType(body.Title, body.Description, body.DurationMinutes); err != nil {
		return gen.AdminEventTypesCreateEventType400JSONResponse(errorResponse("bad_request", err.Error())), nil
	}

	eventType, err := h.store.CreateEventType(ctx, db.CreateEventTypeParams{
		ID:              uuidParam(body.Id),
		Title:           strings.TrimSpace(body.Title),
		Description:     strings.TrimSpace(body.Description),
		DurationMinutes: body.DurationMinutes,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return gen.AdminEventTypesCreateEventType409JSONResponse(errorResponse("event_type_exists", "event type already exists")), nil
		}
		if isCheckViolation(err) {
			return gen.AdminEventTypesCreateEventType400JSONResponse(errorResponse("bad_request", "event type data is invalid")), nil
		}
		return nil, err
	}

	return gen.AdminEventTypesCreateEventType201JSONResponse(eventTypeFromDB(eventType)), nil
}

func (h *Handler) BookingsCreateBooking(
	ctx context.Context,
	request gen.BookingsCreateBookingRequestObject,
) (gen.BookingsCreateBookingResponseObject, error) {
	if request.Body == nil {
		return gen.BookingsCreateBooking400JSONResponse(errorResponse("bad_request", "request body is required")), nil
	}

	body := *request.Body
	if err := validateBooking(body.GuestName, string(body.GuestEmail)); err != nil {
		return gen.BookingsCreateBooking400JSONResponse(errorResponse("bad_request", err.Error())), nil
	}

	eventType, err := h.store.GetEventType(ctx, uuidParam(body.EventTypeId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.BookingsCreateBooking404JSONResponse(errorResponse("event_type_not_found", "event type not found")), nil
		}
		return nil, err
	}

	startsAt := body.StartsAt.UTC()
	duration := time.Duration(eventType.DurationMinutes) * time.Minute
	endsAt := startsAt.Add(duration)
	windowStartsAt, windowEndsAt := bookingWindowBounds(h.now().UTC())
	if !isValidSlotStart(startsAt, duration, windowStartsAt, windowEndsAt) {
		return gen.BookingsCreateBooking400JSONResponse(errorResponse("invalid_slot", "selected time does not match an available slot")), nil
	}

	occupied, err := h.isSlotOccupied(ctx, startsAt, endsAt)
	if err != nil {
		return nil, err
	}
	if occupied {
		return gen.BookingsCreateBooking409JSONResponse(errorResponse("slot_unavailable", "selected slot is already booked")), nil
	}

	booking, err := h.store.CreateBooking(ctx, db.CreateBookingParams{
		EventTypeID: uuidParam(body.EventTypeId),
		StartsAt:    timestamptz(startsAt),
		EndsAt:      timestamptz(endsAt),
		GuestName:   strings.TrimSpace(body.GuestName),
		GuestEmail:  strings.TrimSpace(string(body.GuestEmail)),
		GuestNote:   optionalTrimmedString(body.GuestNote),
	})
	if err != nil {
		if isExclusionViolation(err) {
			return gen.BookingsCreateBooking409JSONResponse(errorResponse("slot_unavailable", "selected slot is already booked")), nil
		}
		if isForeignKeyViolation(err) {
			return gen.BookingsCreateBooking404JSONResponse(errorResponse("event_type_not_found", "event type not found")), nil
		}
		if isCheckViolation(err) {
			return gen.BookingsCreateBooking400JSONResponse(errorResponse("bad_request", "booking data is invalid")), nil
		}
		return nil, err
	}

	return gen.BookingsCreateBooking201JSONResponse(bookingFromCreateRow(booking)), nil
}

func (h *Handler) EventTypesListEventTypes(
	ctx context.Context,
	_ gen.EventTypesListEventTypesRequestObject,
) (gen.EventTypesListEventTypesResponseObject, error) {
	rows, err := h.store.ListEventTypes(ctx)
	if err != nil {
		return nil, err
	}

	eventTypes := make([]gen.EventType, 0, len(rows))
	for _, row := range rows {
		eventTypes = append(eventTypes, eventTypeFromDB(row))
	}

	return gen.EventTypesListEventTypes200JSONResponse(eventTypes), nil
}

func (h *Handler) SlotsListAvailableSlots(
	ctx context.Context,
	request gen.SlotsListAvailableSlotsRequestObject,
) (gen.SlotsListAvailableSlotsResponseObject, error) {
	eventType, err := h.store.GetEventType(ctx, uuidParam(request.EventTypeId))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.SlotsListAvailableSlots404JSONResponse(errorResponse("event_type_not_found", "event type not found")), nil
		}
		return nil, err
	}

	windowStartsAt, windowEndsAt := bookingWindowBounds(h.now().UTC())
	bookings, err := h.store.ListBookingsInWindow(ctx, db.ListBookingsInWindowParams{
		WindowStartsAt: timestamptz(windowStartsAt),
		WindowEndsAt:   timestamptz(windowEndsAt),
	})
	if err != nil {
		return nil, err
	}

	duration := time.Duration(eventType.DurationMinutes) * time.Minute
	slots := availableSlots(windowStartsAt, windowEndsAt, duration, bookings)

	return gen.SlotsListAvailableSlots200JSONResponse(gen.SlotsResponse{
		EventTypeId:    request.EventTypeId,
		WindowStartsAt: windowStartsAt,
		WindowEndsAt:   windowEndsAt,
		Slots:          slots,
	}), nil
}

func (h *Handler) isSlotOccupied(ctx context.Context, startsAt time.Time, endsAt time.Time) (bool, error) {
	bookings, err := h.store.ListBookingsInWindow(ctx, db.ListBookingsInWindowParams{
		WindowStartsAt: timestamptz(startsAt),
		WindowEndsAt:   timestamptz(endsAt),
	})
	if err != nil {
		return false, err
	}
	return len(bookings) > 0, nil
}

func availableSlots(windowStartsAt time.Time, windowEndsAt time.Time, duration time.Duration, bookings []db.ListBookingsInWindowRow) []gen.AvailableSlot {
	if duration <= 0 {
		return nil
	}

	slots := make([]gen.AvailableSlot, 0)
	for startsAt := nextSlotStart(windowStartsAt, duration); !startsAt.Add(duration).After(windowEndsAt); startsAt = startsAt.Add(duration) {
		endsAt := startsAt.Add(duration)
		if overlapsAny(startsAt, endsAt, bookings) {
			continue
		}
		slots = append(slots, gen.AvailableSlot{
			StartsAt: startsAt,
			EndsAt:   endsAt,
		})
	}
	return slots
}

func overlapsAny(startsAt time.Time, endsAt time.Time, bookings []db.ListBookingsInWindowRow) bool {
	for _, booking := range bookings {
		if startsAt.Before(booking.EndsAt.Time) && endsAt.After(booking.StartsAt.Time) {
			return true
		}
	}
	return false
}

func isValidSlotStart(startsAt time.Time, duration time.Duration, windowStartsAt time.Time, windowEndsAt time.Time) bool {
	if duration <= 0 {
		return false
	}
	if startsAt.Before(windowStartsAt) || startsAt.Add(duration).After(windowEndsAt) {
		return false
	}
	return startsAt.Equal(nextSlotStart(startsAt.Add(-time.Nanosecond), duration))
}

func bookingWindowBounds(now time.Time) (time.Time, time.Time) {
	windowStartsAt := now.UTC().Truncate(time.Minute)
	return windowStartsAt, windowStartsAt.Add(bookingWindow)
}

func nextSlotStart(after time.Time, duration time.Duration) time.Time {
	after = after.UTC()
	dayStart := time.Date(after.Year(), after.Month(), after.Day(), 0, 0, 0, 0, time.UTC)
	elapsed := after.Sub(dayStart)
	start := dayStart.Add((elapsed / duration) * duration)
	if !start.After(after) {
		start = start.Add(duration)
	}
	return start
}

func validateEventType(title string, description string, durationMinutes int32) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("title is required")
	}
	if len(title) > 200 {
		return errors.New("title must be 200 characters or fewer")
	}
	if strings.TrimSpace(description) == "" {
		return errors.New("description is required")
	}
	if durationMinutes < 1 {
		return errors.New("durationMinutes must be greater than zero")
	}
	return nil
}

func validateBooking(guestName string, guestEmail string) error {
	guestName = strings.TrimSpace(guestName)
	if guestName == "" {
		return errors.New("guestName is required")
	}
	if len(guestName) > 200 {
		return errors.New("guestName must be 200 characters or fewer")
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(guestEmail)); err != nil {
		return errors.New("guestEmail must be a valid email address")
	}
	return nil
}

func eventTypeFromDB(eventType db.EventType) gen.EventType {
	return gen.EventType{
		Id:              uuidFromPg(eventType.ID),
		Title:           eventType.Title,
		Description:     eventType.Description,
		DurationMinutes: eventType.DurationMinutes,
	}
}

func bookingFromCreateRow(row db.CreateBookingRow) gen.Booking {
	return gen.Booking{
		Id:             uuidFromPg(row.ID),
		EventTypeId:    uuidFromPg(row.EventTypeID),
		EventTypeTitle: row.EventTypeTitle,
		StartsAt:       row.StartsAt.Time,
		EndsAt:         row.EndsAt.Time,
		GuestName:      row.GuestName,
		GuestEmail:     openapi_types.Email(row.GuestEmail),
		GuestNote:      row.GuestNote,
		CreatedAt:      row.CreatedAt.Time,
	}
}

func bookingFromUpcomingRow(row db.ListUpcomingBookingsRow) gen.Booking {
	return gen.Booking{
		Id:             uuidFromPg(row.ID),
		EventTypeId:    uuidFromPg(row.EventTypeID),
		EventTypeTitle: row.EventTypeTitle,
		StartsAt:       row.StartsAt.Time,
		EndsAt:         row.EndsAt.Time,
		GuestName:      row.GuestName,
		GuestEmail:     openapi_types.Email(row.GuestEmail),
		GuestNote:      row.GuestNote,
		CreatedAt:      row.CreatedAt.Time,
	}
}

func errorResponse(code string, message string) gen.ErrorResponse {
	return gen.ErrorResponse{
		Error: gen.ErrorBody{
			Code:    code,
			Message: message,
		},
	}
}

func optionalTrimmedString(value *string) *string {
	if value == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}
	return &trimmed
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

func isUniqueViolation(err error) bool {
	return pgErrorCode(err) == "23505"
}

func isForeignKeyViolation(err error) bool {
	return pgErrorCode(err) == "23503"
}

func isCheckViolation(err error) bool {
	return pgErrorCode(err) == "23514"
}

func isExclusionViolation(err error) bool {
	return pgErrorCode(err) == "23P01"
}

func pgErrorCode(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code
	}
	return ""
}
