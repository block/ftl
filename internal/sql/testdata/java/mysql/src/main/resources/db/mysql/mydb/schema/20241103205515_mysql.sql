-- migrate:up
CREATE TABLE test_types (
    id BIGINT NOT NULL AUTO_INCREMENT PRIMARY KEY,
    int_val INTEGER NOT NULL,
    float_val DOUBLE NOT NULL,
    text_val TEXT NOT NULL,
    bool_val BOOLEAN NOT NULL,
    time_val TIMESTAMP NOT NULL,
    optional_val TEXT
);

-- migrate:down
DROP TABLE test_types;
