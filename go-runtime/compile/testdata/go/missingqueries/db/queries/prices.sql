
-- name: InsertPrice :exec
INSERT INTO prices (code, price, time, currency) VALUES (?, ?, ?, ?);