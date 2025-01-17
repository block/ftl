-- name: GetAllTypes :one
SELECT * FROM all_types WHERE id = $1;

-- name: CreateAllTypes :exec
INSERT INTO all_types (
    big_int, small_int, 
    some_decimal, some_numeric, some_float, some_double,
    some_varchar, some_text, some_char, nullable_text,
    some_bool, nullable_bool,
    some_date, some_time, some_timestamp,
    some_blob, some_json
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
);

