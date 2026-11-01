-- +goose Up

CREATE TABLE feed_follows (
  id TEXT PRIMARY KEY,
  created_at timestamp NOT NULL,
  updated_at timestamp NOT NULL,
  user_id TEXT NOT NULL REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
  feed_id TEXT NOT NULL REFERENCES feeds (id) ON DELETE CASCADE ON UPDATE CASCADE,
  unique (user_id, feed_id)
);

-- +goose Down
DROP TABLE feed_follows;
