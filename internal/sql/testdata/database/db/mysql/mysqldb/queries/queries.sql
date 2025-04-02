-- name: GetRequestDataMySQL :many
SELECT data FROM requests;

-- name: CreateRequestMySQL :exec
INSERT INTO requests (data) VALUES (?);

-- name: GetAllAuthorsMySQL :many
SELECT * FROM authors;

-- name: GetAuthorByIdMySQL :one
SELECT * FROM authors WHERE id = ?;

-- name: GetAuthorInfoMySQL :one
SELECT bio, hometown FROM authors WHERE id = ?;

-- name: GetManyAuthorsInfoMySQL :many
SELECT bio, hometown FROM authors;
