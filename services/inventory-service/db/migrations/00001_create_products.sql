-- +goose Up
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    sku TEXT NOT NULL UNIQUE,
    price NUMERIC(10,2) NOT NULL,
    stock_quantity INTEGER NOT NULL DEFAULT 0 CHECK (stock_quantity >= 0)
);

-- +goose Down
DROP TABLE products;