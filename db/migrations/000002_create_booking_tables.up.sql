CREATE TABLE event_types (
    id uuid PRIMARY KEY,
    title text NOT NULL CHECK (length(title) BETWEEN 1 AND 200),
    description text NOT NULL,
    duration_minutes integer NOT NULL CHECK (duration_minutes > 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE bookings (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type_id uuid NOT NULL REFERENCES event_types (id),
    starts_at timestamptz NOT NULL,
    ends_at timestamptz NOT NULL,
    guest_name text NOT NULL CHECK (length(guest_name) BETWEEN 1 AND 200),
    guest_email text NOT NULL,
    guest_note text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (ends_at > starts_at),
    EXCLUDE USING gist (tstzrange(starts_at, ends_at, '[)') WITH &&)
);

CREATE INDEX bookings_starts_at_idx ON bookings (starts_at);
CREATE INDEX bookings_event_type_id_idx ON bookings (event_type_id);
