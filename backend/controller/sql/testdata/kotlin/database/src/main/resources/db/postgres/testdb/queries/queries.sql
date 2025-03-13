-- name: InsertRequest :one
INSERT INTO requests (data, id) VALUES ($1, $2) RETURNING *;
