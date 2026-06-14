CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE products (
    id          UUID            PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(255)    NOT NULL,
    description TEXT            NOT NULL DEFAULT '',
    price       NUMERIC(10, 2)  NOT NULL CHECK (price >= 0),
    quantity    INTEGER         NOT NULL DEFAULT 0 CHECK (quantity >= 0),
    created_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ     NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_products_name       ON products (name);
CREATE INDEX idx_products_created_at ON products (created_at DESC);
