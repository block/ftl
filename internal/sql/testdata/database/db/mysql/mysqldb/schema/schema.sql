-- migrate:up
CREATE TABLE requests
(
  data TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- migrate:down
DROP TABLE requests;
