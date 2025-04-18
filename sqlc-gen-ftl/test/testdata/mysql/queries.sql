-- name: GetAllTypes :one
SELECT * FROM all_types WHERE id = ?;

-- name: CreateAllTypes :exec
INSERT INTO all_types (
    big_int, small_int, 
    some_decimal, some_numeric, some_float, some_double,
    some_varchar, some_text, some_char, nullable_text,
    some_bool, nullable_bool,
    some_date, some_time, some_timestamp,
    some_blob, some_json
) VALUES (
    ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?
);

-- name: GetSingleColumn :one
SELECT some_text FROM all_types WHERE id = ?;

-- name: GetPartialTable :one
SELECT id, some_text, some_bool FROM all_types WHERE id = ?;

-- name: GetAllTypesMany :many
SELECT * FROM all_types LIMIT 10;
