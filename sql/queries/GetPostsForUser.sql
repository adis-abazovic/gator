-- name: GetPostsForUser :many
SELECT *
FROM posts
INNER JOIN feeds
ON feeds.user_id = $1 AND feeds.id = posts.feed_id
ORDER BY posts.created_at DESC
LIMIT $2;