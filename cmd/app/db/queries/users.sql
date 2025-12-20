-- name: Create :execresult
INSERT INTO users (email, password_hash, created_at, updated_at)
VALUES (?, ?, unixepoch(), unixepoch());

-- name: GetByEmail :one
SELECT id, email, password_hash, locked_until, created_at, updated_at
FROM users
WHERE email = ?;

-- name: GetByID :one
SELECT id, email, password_hash, locked_until, created_at, updated_at
FROM users
WHERE id = ?;

-- name: Delete :execresult
DELETE FROM users
WHERE id = ?;
