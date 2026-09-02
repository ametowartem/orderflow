-- name: CreateReservation :one
INSERT INTO stock_reservations (product_id, order_id, quantity, expires_at) VALUES ($1, $2, $3, $4) RETURNING *;

-- name: GetReservationByID :one
SELECT * FROM stock_reservations WHERE id = $1;

-- name: UpdateReservationStatus :exec
UPDATE stock_reservations SET status = $2 WHERE id = $1;

-- name: DeleteExpiredReservations :many
DELETE FROM stock_reservations WHERE expires_at < now() AND status = 'reserved' RETURNING *;
