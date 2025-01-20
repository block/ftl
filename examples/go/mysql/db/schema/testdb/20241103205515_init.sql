-- migrate:up
CREATE TABLE authors (
  id         SERIAL PRIMARY KEY,
  bio        text   NULL DEFAULT NULL,
  birth_year int    NULL DEFAULT NULL,
  hometown   text   NULL DEFAULT NULL
);

-- migrate:down
DROP TABLE authors;
