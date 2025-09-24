-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS events
(
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title           TEXT      NOT NULL,
    date            TIMESTAMPTZ NOT NULL,
    total_seats     INT       NOT NULL,
    available_seats INT       NOT NULL,
    booking_ttl     INTERVAL         DEFAULT '30 minutes',
    created_at      TIMESTAMPTZ        DEFAULT NOW(),
    updated_at      TIMESTAMPTZ        DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
