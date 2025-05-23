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

-- name: findMultiple :many
SELECT * FROM requests WHERE `data` IN (sqlc.slice("dataValues"));

-- name: findByDataAndIds :many
SELECT * FROM requests WHERE `data` IN (sqlc.slice("dataValues")) AND `id` IN (sqlc.slice("ids"));

-- name: findTestTypesBySlices :many
SELECT * FROM test_types WHERE `text_val` IN (sqlc.slice("textValues")) AND `int_val` IN (sqlc.slice("intValues")) AND `float_val` IN (sqlc.slice("floatValues"));
