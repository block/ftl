-- migrate:up
CREATE TABLE records (
  id SERIAL PRIMARY KEY
);

-- migrate:down
DROP TABLE records;
