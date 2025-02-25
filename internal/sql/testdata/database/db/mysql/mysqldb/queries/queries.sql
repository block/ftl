-- name: GetRequestDataMySQL :many
SELECT data FROM requests;

-- name: CreateRequestMySQL :exec
INSERT INTO requests (data) VALUES (?);
