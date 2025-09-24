-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS bookings
(
    id         UUID PRIMARY KEY                                               DEFAULT gen_random_uuid(),
    event_id   UUID REFERENCES events (id),
    user_id    UUID REFERENCES users (id),
    status     TEXT CHECK ( status IN ('pending', 'confirmed', 'cancelled') ) DEFAULT 'pending',
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ                                                    DEFAULT NOW(),
    updated_at TIMESTAMPTZ                                                    DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS bookings;
-- +goose StatementEnd
