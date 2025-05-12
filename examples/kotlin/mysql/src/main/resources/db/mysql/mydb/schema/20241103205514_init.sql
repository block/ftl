-- migrate:up
CREATE TABLE requests
(
  id int NOT NULL AUTO_INCREMENT,
  data TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id)
);
-- migrate:down
DROP TABLE requests;