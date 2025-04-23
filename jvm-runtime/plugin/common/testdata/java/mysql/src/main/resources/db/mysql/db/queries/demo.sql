-- name: CreateDemoRow :exec
INSERT INTO demo (
    required_string,
    optional_string,
    number_value,
    timestamp_value,
    float_value
) VALUES (
    ?,
    ?,
    ?,
    ?,
    ?
);

-- name: ListDemoRows :many
SELECT
    id,
    required_string,
    optional_string,
    number_value,
    timestamp_value,
    float_value
FROM demo
ORDER BY id;