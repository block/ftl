-- name: GetUserByID :one
SELECT id, name, email FROM users WHERE id = ?;

-- name: CreateUser :exec
INSERT INTO users (name, email) VALUES (?, ?);

-- name: GetRequestData :many
SELECT data FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES (?);

