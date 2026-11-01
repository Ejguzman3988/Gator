-- name: CreateFeed :one
INSERT INTO feeds (id, created_at, updated_at, name, url, user_id)
VALUES (
  $1,
  $2,
  $3,
  $4,
  $5,
  $6
)
RETURNING *;

-- name: ListFeeds :many
SELECT f.name AS "feed_name", f.url AS "feed_url", u.name AS "user_name" FROM feeds f
JOIN users u ON u.id = f.user_id;

-- name: FindFeedByUrl :one
SELECT * FROM feeds 
WHERE url = $1;

-- name: CreateFeedFollow :one
WITH inserted_feed_follow AS (
  INSERT INTO feed_follows (id, created_at, updated_at, user_id, feed_id)
    VALUES (
      $1, --id
      $2, --created_at
      $3, --updated_at
      $4, --user_id
      $5  --feed_id
    )
    RETURNING *
)
SELECT 
  ff.*,
  f.name as feed_name,
  u.name as user_name
FROM inserted_feed_follow ff
JOIN users u ON u.id = ff.user_id
JOIN feeds f ON f.id = ff.feed_id;

-- name: UnfollowFeed :exec
DELETE FROM feed_follows ff where ff.feed_id = $1 AND  ff.user_id = $2;

-- name: GetFeedFollowsForUser :many
SELECT f.id as feed_id, f.name as feed_name, u.name as user_name from feeds f
JOIN feed_follows ff ON f.id = ff.feed_id
JOIN users u ON u.id = ff.user_id
WHERE ff.user_id = $1;


-- name: MarkFeedFetched :one
UPDATE feeds
SET last_fetched_at = $2, updated_at = $2
WHERE id = $1
RETURNING *;

-- name: GetNextFeedToFetch :one
SELECT * FROM feeds
ORDER BY last_fetched_at ASC NULLS FIRST;

-- name: UpdateLastFetchedAt :exec
UPDATE feeds 
SET last_fetched_at = $2 
WHERE id = $1;
