-- +goose Up

CREATE TABLE feeds (
  id TEXT PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  unique (url),
  user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
);

-- +goose Down
DROP TABLE feeds;
