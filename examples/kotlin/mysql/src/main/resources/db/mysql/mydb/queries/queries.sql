-- name: GetRequestData :many
SELECT data FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES (?);

-- name: GetRequestDataOne :one
SELECT data FROM requests LIMIT 1;
