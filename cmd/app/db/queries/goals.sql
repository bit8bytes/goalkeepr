-- name: Create :one
INSERT INTO goals (user_id, goal, description, due, visible_to_public, achieved)
VALUES (?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: Get :one
SELECT * FROM goals
WHERE id = ? AND user_id = ?;

-- name: GetAll :many
SELECT * FROM goals
WHERE user_id = ?
ORDER BY due ASC;

-- name: GetAllShared :many
SELECT * FROM goals
WHERE user_id = ? AND visible_to_public = 1
ORDER BY due ASC;

-- name: Update :execresult
UPDATE goals
SET goal = ?, description = ?, due = ?, visible_to_public = ?, achieved = ?
WHERE id = ? AND user_id = ?;

-- name: Delete :execresult
DELETE FROM goals
WHERE id = ? AND user_id = ?;
