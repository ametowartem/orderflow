-- name: GetProductByID :one
SELECT * FROM products WHERE id = $1;

-- name: GetProductForUpdate :one
SELECT * FROM products WHERE id = $1 FOR UPDATE;

-- name: DecrementStock :exec
UPDATE products SET stock_quantity = stock_quantity - $2 WHERE id = $1;
