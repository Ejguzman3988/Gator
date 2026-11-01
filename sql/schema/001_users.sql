-- +goose Up

CREATE TABLE users (
  id TEXT PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  name TEXT NOT NULL,
  unique (name)
);

-- +goose Down
DROP TABLE users;
