-- migrate:up
CREATE TABLE demo (
    id INT PRIMARY KEY AUTO_INCREMENT,
    required_string VARCHAR(255) NOT NULL,
    optional_string VARCHAR(255),
    number_value INT NOT NULL,
    timestamp_value DATETIME NOT NULL,
    float_value FLOAT NOT NULL
);

-- migrate:down
DROP TABLE demo;