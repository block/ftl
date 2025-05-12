-- name: CreateRequest :one
INSERT INTO requests (data) VALUES ($1) RETURNING *;

-- name: GetRequestData :many
SELECT data FROM requests;

-- name: findMultiple :many
SELECT * FROM requests WHERE data = ANY($1::text[]);

-- name: findByDataAndIds :many
SELECT * FROM requests WHERE data = ANY($1::text[]) AND id = ANY($2::int[]);
