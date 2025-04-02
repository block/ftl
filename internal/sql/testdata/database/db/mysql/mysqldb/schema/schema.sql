-- migrate:up
CREATE TABLE requests
(
  data TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE authors (
  id         SERIAL PRIMARY KEY,
  bio        text   NULL DEFAULT NULL,
  birth_year int    NULL DEFAULT NULL,
  hometown   text   NULL DEFAULT NULL
);

-- migrate:down
DROP TABLE authors;
DROP TABLE requests;
