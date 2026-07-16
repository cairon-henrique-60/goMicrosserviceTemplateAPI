-- +goose Up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE orders (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), customer_id UUID NOT NULL, status VARCHAR(40) NOT NULL, total NUMERIC(14,2) NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW());
CREATE TABLE order_items (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE, product_id UUID NOT NULL, product_name VARCHAR(180) NOT NULL, quantity INTEGER NOT NULL CHECK(quantity > 0), unit_price NUMERIC(14,2) NOT NULL, subtotal NUMERIC(14,2) NOT NULL);
CREATE TABLE outbox_events (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), aggregate_type VARCHAR(100) NOT NULL, aggregate_id UUID NOT NULL, event_type VARCHAR(150) NOT NULL, payload JSONB NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(), published_at TIMESTAMPTZ);
CREATE INDEX idx_order_outbox_unpublished ON outbox_events(created_at) WHERE published_at IS NULL;
-- +goose Down
DROP TABLE IF EXISTS outbox_events; DROP TABLE IF EXISTS order_items; DROP TABLE IF EXISTS orders;
