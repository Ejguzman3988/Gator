-- +goose Up

CREATE TABLE posts (
  id TEXT PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  title TEXT NOT NULL,
  url TEXT NOT NULL,
  description TEXT,
  published_at timestamp NOT NULL,
  feed_id TEXT NOT NULL REFERENCES feeds (id) ON DELETE CASCADE ON UPDATE CASCADE,
  unique(url)
);

-- +goose Down
DROP TABLE posts;

