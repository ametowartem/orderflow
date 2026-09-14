-- +goose Up
CREATE TABLE payments (
    id         UUID            DEFAULT gen_random_uuid() PRIMARY KEY,
    order_id   UUID            NOT NULL UNIQUE,
    amount     NUMERIC (10, 2) NOT NULL,
    status     TEXT            DEFAULT 'processing' NOT NULL,
    created_at TIMESTAMPTZ     DEFAULT now() NOT NULL,
    updated_at TIMESTAMPTZ     DEFAULT now() NOT NULL
);

-- +goose Down
DROP TABLE IF EXISTS payments;