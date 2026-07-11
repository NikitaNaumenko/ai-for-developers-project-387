-- name: CreateEvent :one
INSERT INTO events (
    title,
    description,
    starts_at,
    ends_at
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id, title, description, starts_at, ends_at, created_at, updated_at;

-- name: ListEvents :many
SELECT id, title, description, starts_at, ends_at, created_at, updated_at
FROM events
ORDER BY starts_at ASC, id ASC
LIMIT $1 OFFSET $2;

-- name: CreateEventType :one
INSERT INTO event_types (
    id,
    title,
    description,
    duration_minutes
) VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING id, title, description, duration_minutes, created_at, updated_at;

-- name: ListEventTypes :many
SELECT id, title, description, duration_minutes, created_at, updated_at
FROM event_types
ORDER BY title ASC, id ASC;

-- name: GetEventType :one
SELECT id, title, description, duration_minutes, created_at, updated_at
FROM event_types
WHERE id = $1;

-- name: ListBookingsInWindow :many
SELECT
    b.id,
    b.event_type_id,
    et.title AS event_type_title,
    b.starts_at,
    b.ends_at,
    b.guest_name,
    b.guest_email,
    b.guest_note,
    b.created_at
FROM bookings b
JOIN event_types et ON et.id = b.event_type_id
WHERE b.starts_at < sqlc.arg(window_ends_at)
  AND b.ends_at > sqlc.arg(window_starts_at)
ORDER BY b.starts_at ASC, b.id ASC;

-- name: CreateBooking :one
INSERT INTO bookings (
    event_type_id,
    starts_at,
    ends_at,
    guest_name,
    guest_email,
    guest_note
) VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    $6
)
RETURNING
    id,
    event_type_id,
    (SELECT title FROM event_types WHERE id = $1) AS event_type_title,
    starts_at,
    ends_at,
    guest_name,
    guest_email,
    guest_note,
    created_at;

-- name: ListUpcomingBookings :many
SELECT
    b.id,
    b.event_type_id,
    et.title AS event_type_title,
    b.starts_at,
    b.ends_at,
    b.guest_name,
    b.guest_email,
    b.guest_note,
    b.created_at
FROM bookings b
JOIN event_types et ON et.id = b.event_type_id
WHERE b.starts_at >= $1
ORDER BY b.starts_at ASC, b.id ASC;
