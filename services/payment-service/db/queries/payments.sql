-- name: GetPaymentByOrderID :one
SELECT * FROM payments WHERE order_id = $1;

-- name: CreatePayment :one
INSERT INTO payments (order_id, amount) VALUES ($1, $2) RETURNING *;

-- name: UpdatePaymentStatus :exec
UPDATE payments SET status = $2 WHERE id = $1;
