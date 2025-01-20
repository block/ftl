CREATE TABLE all_types (
    -- Integer types
    id SERIAL PRIMARY KEY,
    big_int BIGINT NOT NULL,
    small_int SMALLINT NOT NULL,
    
    -- Numeric/Decimal types
    some_decimal DECIMAL(10,2) NOT NULL,
    some_numeric NUMERIC(10,2) NOT NULL,
    some_float FLOAT NOT NULL,
    some_double DOUBLE PRECISION NOT NULL,
    
    -- Character types
    some_varchar VARCHAR(255) NOT NULL,
    some_text TEXT NOT NULL,
    some_char CHAR(10) NOT NULL,
    nullable_text TEXT,
    
    -- Boolean type
    some_bool BOOLEAN NOT NULL,
    nullable_bool BOOLEAN,
    
    -- Date/Time types
    some_date DATE NOT NULL,
    some_time TIME NOT NULL,
    some_timestamp TIMESTAMP NOT NULL,
    
    -- Binary type
    some_blob BLOB NOT NULL,
    
    -- JSON type
    some_json JSON NOT NULL
);
