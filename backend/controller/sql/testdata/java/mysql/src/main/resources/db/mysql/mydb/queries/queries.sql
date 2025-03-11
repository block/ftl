-- name: GetRequestData :many
SELECT data FROM requests;

-- name: CreateRequest :exec
INSERT INTO requests (data) VALUES (?);

-- name: InsertTestTypes :exec
INSERT INTO test_types (int_val, float_val, text_val, bool_val, time_val, optional_val) 
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetTestType :one
SELECT id, int_val, float_val, text_val, bool_val, time_val, optional_val 
FROM test_types 
WHERE id = ?;

-- name: GetAllTestTypes :many
SELECT id, int_val, float_val, text_val, bool_val, time_val, optional_val 
FROM test_types;
