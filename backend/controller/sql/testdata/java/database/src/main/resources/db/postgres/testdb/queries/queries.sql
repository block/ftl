-- name: CreateRequest :one
INSERT INTO requests (data) VALUES ($1) RETURNING *;

-- name: GetRequestData :many
SELECT data FROM requests;
