-- name: GetRequestData :many
SELECT data, id FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES (?);

-- name: GetRequestDataOne :one
SELECT data, id FROM requests LIMIT 1;

-- name: GetRequestsWithIDs :many
SELECT data, id FROM requests WHERE id IN(sqlc.slice("ids"))