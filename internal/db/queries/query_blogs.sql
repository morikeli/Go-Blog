-- name: CreateBlog :one
INSERT INTO blogs (title, content, user_id, created_at, updated_at)
VALUES ($1, $2, $3, NOW(), NOW())
RETURNING id, title, content, user_id, created_at, updated_at;

-- name: ListBlogs :many
SELECT id, title, content, user_id, created_at, updated_at
FROM blogs
ORDER BY created_at ASC;
