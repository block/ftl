-- migrate:up
CREATE TABLE foo (a text, b integer, c DATETIME, d DATE);

-- migrate:down
DROP TABLE foo;
