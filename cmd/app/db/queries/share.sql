-- name: Create :execresult
INSERT INTO share (user_id, public_id)
VALUES (?, ?);

-- name: CountByUserID :one
SELECT COUNT(*) FROM share WHERE user_id = ?;

-- name: GetAll :many
SELECT id, user_id, public_id
FROM share
WHERE user_id = ?;

-- name: GetUserIDByPublicID :one
SELECT user_id FROM share WHERE public_id = ?;

-- name: Delete :execresult
DELETE FROM share WHERE id = ?;
