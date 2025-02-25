-- name: GetRequestDataPSQL :many
SELECT data FROM requests;

-- name: CreateRequestPSQL :exec
INSERT INTO requests (data) VALUES ($1);
