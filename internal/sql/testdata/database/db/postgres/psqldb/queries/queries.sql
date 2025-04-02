-- name: GetRequestDataPSQL :many
SELECT data FROM requests;

-- name: CreateRequestPSQL :exec
INSERT INTO requests (data) VALUES ($1);

-- name: GetAllAuthorsPSQL :many
SELECT * FROM authors;

-- name: GetAuthorByIdPSQL :one
SELECT * FROM authors WHERE id = $1;

-- name: GetAuthorInfoPSQL :one
SELECT bio, hometown FROM authors WHERE id = $1;

-- name: GetManyAuthorsInfoPSQL :many
SELECT bio, hometown FROM authors;
