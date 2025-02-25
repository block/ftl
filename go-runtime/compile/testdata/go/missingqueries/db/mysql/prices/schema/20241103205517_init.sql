-- migrate:up
CREATE TABLE prices (
  code VARCHAR(255) NOT NULL,
  price DECIMAL(10, 2) NOT NULL,
  time TIMESTAMP NOT NULL,
  currency VARCHAR(3) NOT NULL
);

-- migrate:down
DROP TABLE prices;
