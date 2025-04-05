-- name: GetRequestData :many
SELECT data FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES (?);

-- name: GetAllAuthors :many
SELECT * FROM authors;

-- name: GetAuthorById :one
SELECT * FROM authors WHERE id = ?;

-- name: GetAuthorInfo :one
SELECT bio, hometown FROM authors WHERE id = ?;

-- name: GetManyAuthorsInfo :many
SELECT bio, hometown FROM authors WHERE id IN (?);
