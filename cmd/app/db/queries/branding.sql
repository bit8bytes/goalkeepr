-- name: GetByUserID :one
SELECT id, user_id, title, description
FROM branding
WHERE user_id = ?;

-- name: CreateOrUpdate :execresult
INSERT INTO branding (user_id, title, description)
VALUES (?, ?, ?)
ON CONFLICT(user_id) DO UPDATE SET
    title = excluded.title,
    description = excluded.description;

-- name: Delete :execresult
DELETE FROM branding
WHERE user_id = ?;
