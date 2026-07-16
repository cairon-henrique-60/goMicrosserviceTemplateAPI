-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE products (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), sku VARCHAR(80) NOT NULL UNIQUE,
 name VARCHAR(180) NOT NULL, description TEXT NOT NULL DEFAULT '',
 price NUMERIC(14,2) NOT NULL CHECK(price >= 0), stock INTEGER NOT NULL DEFAULT 0 CHECK(stock >= 0),
 active BOOLEAN NOT NULL DEFAULT TRUE, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TABLE outbox_events (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(), aggregate_type VARCHAR(100) NOT NULL, aggregate_id UUID NOT NULL,
 event_type VARCHAR(150) NOT NULL, payload JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), published_at TIMESTAMPTZ
);
CREATE INDEX idx_catalog_outbox_unpublished ON outbox_events(created_at) WHERE published_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS outbox_events; DROP TABLE IF EXISTS products;
