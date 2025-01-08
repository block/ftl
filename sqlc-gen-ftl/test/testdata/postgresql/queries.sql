-- name: GetUserByID :one
SELECT id, name, email FROM users WHERE id = $1;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES ($1, $2);

-- name: GetRequestData :many
SELECT data FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES ($1);

